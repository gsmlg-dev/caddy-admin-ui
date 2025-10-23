package caddy_admin_ui

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

const (
	// Message types
	msgTypeData    = 0x00
	msgTypeResize  = 0x01
	msgTypePing    = 0x02
	msgTypePong    = 0x03

	// Performance constants
	writeBufferSize  = 8192
	readBufferSize   = 8192
	flushInterval    = 10 * time.Millisecond
	maxConnections   = 100
)

var (
	// Optimized upgrader with compression and larger buffers
	upgraderOptimized = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Implement proper origin validation for production
			// For now, allow specific origins or check against config
			origin := r.Header.Get("Origin")
			return origin == "" || origin == "http://localhost:3000" || origin == "http://localhost:80"
		},
	}

	// Pool of buffers to reduce GC pressure
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, readBufferSize)
		},
	}

	// Connection pool management
	activeConnections int64
)

// WebSocketManager manages active connections with rate limiting
type WebSocketManager struct {
	maxConnections int64
	connections    map[string]*OptimizedWebShell
	mutex          sync.RWMutex
	rateLimiter    *RateLimiter
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		maxConnections: maxConnections,
		connections:    make(map[string]*OptimizedWebShell),
		rateLimiter:    NewRateLimiter(10, 20), // 10 req/s with burst of 20
	}
}

func (wm *WebSocketManager) CanAcceptConnection(r *http.Request) bool {
	clientIP := getClientIP(r)
	return wm.rateLimiter.Allow(clientIP) &&
		   atomic.LoadInt64(&activeConnections) < wm.maxConnections
}

func (wm *WebSocketManager) RegisterConnection(id string, ws *OptimizedWebShell) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	wm.connections[id] = ws
	atomic.AddInt64(&activeConnections, 1)
}

func (wm *WebSocketManager) UnregisterConnection(id string) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	delete(wm.connections, id)
	atomic.AddInt64(&activeConnections, -1)
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	clients map[string]*tokenBucket
	mutex   sync.RWMutex
	rate    float64
	burst   int
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
	mutex    sync.Mutex
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*tokenBucket),
		rate:    rate,
		burst:   burst,
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.RLock()
	bucket, exists := rl.clients[clientIP]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		bucket = &tokenBucket{
			tokens: float64(rl.burst),
			lastTime: time.Now(),
		}
		rl.clients[clientIP] = bucket
		rl.mutex.Unlock()
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens += elapsed * rl.rate
	bucket.lastTime = now

	if bucket.tokens > float64(rl.burst) {
		bucket.tokens = float64(rl.burst)
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}

// OptimizedWebShell is the optimized version of WebShell
type OptimizedWebShell struct {
	conn         *websocket.Conn
	ptmx         *os.File
	cmd          *exec.Cmd
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
	writeBuffer  *bufio.Writer
	flushTimer   *time.Timer
	manager      *WebSocketManager
	connID       string
	lastActivity time.Time
}

func NewOptimizedWebShell(conn *websocket.Conn, manager *WebSocketManager) *OptimizedWebShell {
	ctx, cancel := context.WithCancel(context.Background())

	// Create buffered writer for batching
	writeBuffer := bufio.NewWriterSize(conn, writeBufferSize)

	return &OptimizedWebShell{
		conn:        conn,
		ctx:         ctx,
		cancel:      cancel,
		writeBuffer: writeBuffer,
		manager:     manager,
		connID:      generateConnectionID(),
		lastActivity: time.Now(),
	}
}

func (ows *OptimizedWebShell) Start(shell string) error {
	ows.cmd = exec.CommandContext(ows.ctx, shell)
	ows.cmd.Env = os.Environ()

	// Start the command with PTY
	var err error
	ows.ptmx, err = pty.Start(ows.cmd)
	if err != nil {
		return fmt.Errorf("failed to start PTY: %v", err)
	}

	// Set initial terminal size
	if err := pty.Setsize(ows.ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		log.Printf("Failed to set terminal size: %v", err)
	}

	// Register with manager
	ows.manager.RegisterConnection(ows.connID, ows)

	// Start goroutines for I/O
	go ows.readFromPTY()
	go ows.readFromWebSocket()
	go ows.waitForProcess()
	go ows.flushLoop()

	return nil
}

func (ows *OptimizedWebShell) readFromPTY() {
	defer ows.cleanup()

	// Get buffer from pool
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	for {
		select {
		case <-ows.ctx.Done():
			return
		default:
			// Update activity timestamp
			ows.mutex.Lock()
			ows.lastActivity = time.Now()
			ows.mutex.Unlock()

			n, err := ows.ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("PTY read error: %v", err)
				}
				return
			}

			ows.sendBinaryData(msgTypeData, buf[:n])
		}
	}
}

func (ows *OptimizedWebShell) readFromWebSocket() {
	defer ows.cleanup()

	// Set read deadline for ping/pong
	ows.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ows.conn.SetPongHandler(func(string) error {
		ows.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ows.ctx.Done():
			return
		default:
			messageType, data, err := ows.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			// Update activity timestamp
			ows.mutex.Lock()
			ows.lastActivity = time.Now()
			ows.mutex.Unlock()

			switch messageType {
			case websocket.BinaryMessage:
				ows.handleBinaryMessage(data)
			case websocket.TextMessage:
				// Fallback for legacy clients - handle base64
				ows.handleTextMessage(data)
			case websocket.CloseMessage:
				log.Println("WebSocket connection closed by client")
				return
			case websocket.PingMessage:
				ows.sendBinaryData(msgTypePong, nil)
			}
		}
	}
}

func (ows *OptimizedWebShell) handleBinaryMessage(data []byte) {
	if len(data) == 0 {
		return
	}

	msgType := data[0]
	payload := data[1:]

	switch msgType {
	case msgTypeData:
		// Regular terminal input
		ows.mutex.RLock()
		if ows.ptmx != nil {
			if _, err := ows.ptmx.Write(payload); err != nil {
				log.Printf("PTY write error: %v", err)
			}
		}
		ows.mutex.RUnlock()

	case msgTypeResize:
		// Terminal resize message
		if len(payload) >= 4 {
			rows := binary.BigEndian.Uint16(payload[0:2])
			cols := binary.BigEndian.Uint16(payload[2:4])

			ows.mutex.RLock()
			if ows.ptmx != nil {
				if err := pty.Setsize(ows.ptmx, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
					log.Printf("Failed to resize terminal: %v", err)
				}
			}
			ows.mutex.RUnlock()
		}

	case msgTypePing:
		ows.sendBinaryData(msgTypePong, nil)
	}
}

func (ows *OptimizedWebShell) handleTextMessage(data []byte) {
	// Legacy base64 handling for backward compatibility
	if len(data) > 0 && data[0] == 1 {
		// Resize message in old format
		if len(data) >= 5 {
			rows := uint16(data[1])<<8 | uint16(data[2])
			cols := uint16(data[3])<<8 | uint16(data[4])

			ows.mutex.RLock()
			if ows.ptmx != nil {
				pty.Setsize(ows.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
			}
			ows.mutex.RUnlock()
		}
	} else {
		// Decode base64 data (legacy)
		// This is slower but maintains compatibility
		decoded := make([]byte, len(data))
		n, err := base64.StdEncoding.Decode(decoded, data)
		if err != nil {
			log.Printf("Base64 decode error: %v", err)
			return
		}

		ows.mutex.RLock()
		if ows.ptmx != nil {
			if _, err := ows.ptmx.Write(decoded[:n]); err != nil {
				log.Printf("PTY write error: %v", err)
			}
		}
		ows.mutex.RUnlock()
	}
}

func (ows *OptimizedWebShell) sendBinaryData(msgType byte, data []byte) {
	ows.mutex.Lock()
	defer ows.mutex.Unlock()

	if ows.conn == nil {
		return
	}

	// Create message with type prefix
	message := make([]byte, 1+len(data))
	message[0] = msgType
	if len(data) > 0 {
		copy(message[1:], data)
	}

	// Write to buffered writer
	if _, err := ows.writeBuffer.Write(message); err != nil {
		log.Printf("WebSocket buffer write error: %v", err)
		return
	}

	// Flush immediately for small data or if buffer is getting full
	if len(data) < 64 || ows.writeBuffer.Buffered() > writeBufferSize/2 {
		if err := ows.writeBuffer.Flush(); err != nil {
			log.Printf("WebSocket flush error: %v", err)
		}
	}
}

func (ows *OptimizedWebShell) flushLoop() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ows.ctx.Done():
			// Final flush
			ows.mutex.Lock()
			if ows.writeBuffer != nil {
				ows.writeBuffer.Flush()
			}
			ows.mutex.Unlock()
			return
		case <-ticker.C:
			ows.mutex.Lock()
			if ows.writeBuffer != nil && ows.writeBuffer.Buffered() > 0 {
				if err := ows.writeBuffer.Flush(); err != nil {
					log.Printf("WebSocket periodic flush error: %v", err)
				}
			}
			ows.mutex.Unlock()
		}
	}
}

func (ows *OptimizedWebShell) waitForProcess() {
	defer ows.cleanup()

	if ows.cmd != nil && ows.cmd.Process != nil {
		ows.cmd.Wait()
		log.Printf("Shell process %d exited", ows.cmd.Process.Pid)
	}
}

func (ows *OptimizedWebShell) cleanup() {
	ows.cancel()

	ows.manager.UnregisterConnection(ows.connID)

	ows.mutex.Lock()
	defer ows.mutex.Unlock()

	// Close PTY
	if ows.ptmx != nil {
		ows.ptmx.Close()
		ows.ptmx = nil
	}

	// Kill process
	if ows.cmd != nil && ows.cmd.Process != nil {
		ows.cmd.Process.Kill()
	}

	// Close WebSocket
	if ows.conn != nil {
		// Send close message
		ows.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ows.conn.Close()
		ows.conn = nil
	}

	// Close buffer
	if ows.writeBuffer != nil {
		ows.writeBuffer.Flush()
		ows.writeBuffer = nil
	}
}

// Helper functions
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < idx && xff[commaIdx] != ','; commaIdx++ {
				continue
			}
			return xff[:commaIdx]
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func generateConnectionID() string {
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(),
		atomic.AddInt64(&activeConnections, 1))
}

// Optimized handler with rate limiting and connection management
var wsManager = NewWebSocketManager()

func (adminUI *CaddyAdminUI) handleWsPtyOptimized(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Check rate limits
	if !wsManager.CanAcceptConnection(r) {
		log.Printf("Rate limit exceeded for client: %s", getClientIP(r))
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return nil
	}

	// Upgrade to WebSocket
	conn, err := upgraderOptimized.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return nil
	}

	log.Printf("New WebSocket connection from %s (total: %d)",
		getClientIP(r), atomic.LoadInt64(&activeConnections))

	// Create optimized web shell
	webShell := NewOptimizedWebShell(conn, wsManager)
	shell := adminUI.Shell
	if err := webShell.Start(shell); err != nil {
		log.Printf("Failed to start web shell: %v", err)
		conn.Close()
		return nil
	}

	// Keep connection alive
	<-webShell.ctx.Done()
	log.Printf("WebSocket connection closed for %s", webShell.connID)

	return nil
}