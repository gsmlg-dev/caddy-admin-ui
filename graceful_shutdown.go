package caddy_admin_ui

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// ServerManager handles graceful shutdown of the Caddy Admin UI
type ServerManager struct {
	servers      []*http.Server
	connections  map[net.Conn]struct{}
	webShells    map[string]*OptimizedWebShell
	shutdownTimeout time.Duration
	mutex        sync.RWMutex
	logger       *zap.Logger
	wg           sync.WaitGroup
}

func NewServerManager(logger *zap.Logger) *ServerManager {
	return &ServerManager{
		connections:     make(map[net.Conn]struct{}),
		webShells:       make(map[string]*OptimizedWebShell),
		shutdownTimeout: 30 * time.Second,
		logger:          logger,
	}
}

// AddServer adds an HTTP server to the manager
func (sm *ServerManager) AddServer(server *http.Server) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.servers = append(sm.servers, server)
}

// RegisterConnection tracks an active connection
func (sm *ServerManager) RegisterConnection(conn net.Conn) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.connections[conn] = struct{}{}
}

// UnregisterConnection removes a tracked connection
func (sm *ServerManager) UnregisterConnection(conn net.Conn) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.connections, conn)
}

// RegisterWebShell tracks an active web shell
func (sm *ServerManager) RegisterWebShell(id string, ws *OptimizedWebShell) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.webShells[id] = ws
}

// UnregisterWebShell removes a tracked web shell
func (sm *ServerManager) UnregisterWebShell(id string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.webShells, id)
}

// WaitForShutdown waits for shutdown signals and gracefully shuts down
func (sm *ServerManager) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for signal
	sig := <-sigChan
	sm.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), sm.shutdownTimeout)
	defer cancel()

	// Start graceful shutdown
	if err := sm.Shutdown(ctx); err != nil {
		sm.logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	sm.logger.Info("Graceful shutdown completed")
}

// Shutdown gracefully shuts down all servers and connections
func (sm *ServerManager) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// Close all web shells first
	sm.mutex.Lock()
	for id, ws := range sm.webShells {
		sm.logger.Info("Closing web shell", zap.String("id", id))
		if ws != nil {
			ws.cancel()
		}
	}
	sm.webShells = make(map[string]*OptimizedWebShell)
	sm.mutex.Unlock()

	// Close all connections with a timeout
	sm.mutex.Lock()
	conns := make([]net.Conn, 0, len(sm.connections))
	for conn := range sm.connections {
		conns = append(conns, conn)
	}
	sm.mutex.Unlock()

	// Close connections with timeout
	connClosed := make(chan struct{})
	go func() {
		for _, conn := range conns {
			conn.Close()
		}
		close(connClosed)
	}()

	select {
	case <-connClosed:
		sm.logger.Info("All connections closed")
	case <-ctx.Done():
		sm.logger.Warn("Timeout waiting for connections to close")
	}

	// Shutdown HTTP servers
	var wg sync.WaitGroup
	for _, server := range sm.servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()

			sm.logger.Info("Shutting down server", zap.String("addr", srv.Addr))
			if err := srv.Shutdown(ctx); err != nil {
				sm.logger.Error("Error shutting down server",
					zap.String("addr", srv.Addr),
					zap.Error(err))
				shutdownErr = err
			} else {
				sm.logger.Info("Server shutdown complete", zap.String("addr", srv.Addr))
			}
		}(server)
	}

	// Wait for all servers to shutdown or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		sm.logger.Info("All servers shutdown successfully")
	case <-ctx.Done():
		sm.logger.Error("Timeout waiting for servers to shutdown")
		shutdownErr = ctx.Err()
	}

	return shutdownErr
}

// HealthCheckHandler provides health check endpoint
func (sm *ServerManager) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	status := struct {
		Status       string    `json:"status"`
		Timestamp    time.Time `json:"timestamp"`
		Version      string    `json:"version"`
		Connections  int       `json:"active_connections"`
		WebShells    int       `json:"active_webshells"`
		Uptime       string    `json:"uptime"`
	}{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     caddy.CaddyVersion(),
		Connections: len(sm.connections),
		WebShells:   len(sm.webShells),
		Uptime:      "unknown", // TODO: Track start time
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON response
	if err := writeJSON(w, status); err != nil {
		sm.logger.Error("Error writing health check response", zap.Error(err))
	}
}

// ReadyCheckHandler provides readiness check endpoint
func (sm *ServerManager) ReadyCheckHandler(w http.ResponseWriter, r *http.Request) {
	sm.mutex.RLock()
	ready := len(sm.servers) > 0
	sm.mutex.RUnlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Not Ready"))
	}
}

// MetricsHandler provides basic metrics endpoint
func (sm *ServerManager) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	metrics := `# HELP caddy_admin_ui_connections_active Current number of active connections
# TYPE caddy_admin_ui_connections_active gauge
caddy_admin_ui_connections_active ` + floatToString(float64(len(sm.connections))) + `

# HELP caddy_admin_ui_webshells_active Current number of active web shells
# TYPE caddy_admin_ui_webshells_active gauge
caddy_admin_ui_webshells_active ` + floatToString(float64(len(sm.webShells))) + `

# HELP caddy_admin_ui_uptime_seconds Server uptime in seconds
# TYPE caddy_admin_ui_uptime_seconds counter
caddy_admin_ui_uptime_seconds ` + floatToString(time.Since(startTime).Seconds()) + `

`

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metrics))
}

// Global server manager instance
var (
	serverManager *ServerManager
	startTime     = time.Now()
)

// InitServerManager initializes the global server manager
func InitServerManager(logger *zap.Logger) {
	serverManager = NewServerManager(logger)
}

// GetServerManager returns the global server manager
func GetServerManager() *ServerManager {
	return serverManager
}

// Helper functions
func writeJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func floatToString(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// ConnectionTrackingMiddleware tracks connections for graceful shutdown
func (sm *ServerManager) ConnectionTrackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track the connection
		hijacker, ok := w.(http.Hijacker)
		if ok {
			conn, _, err := hijacker.Hijack()
			if err == nil {
				sm.RegisterConnection(conn)
				defer sm.UnregisterConnection(conn)
			}
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}