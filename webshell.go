package caddy_admin_ui

import (
	"context"
	"encoding/base64"
                                                            	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo purposes
	},
}

type WebShell struct {
	conn   *websocket.Conn
	ptmx   *os.File
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func NewWebShell(conn *websocket.Conn) *WebShell {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebShell{
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ws *WebShell) Start(sh string) error {
	ws.cmd = exec.CommandContext(ws.ctx, sh)
	ws.cmd.Env = os.Environ()

	// Start the command with PTY
	var err error
	ws.ptmx, err = pty.Start(ws.cmd)
	if err != nil {
		return fmt.Errorf("failed to start PTY: %v", err)
	}

	// Set initial terminal size (will be updated by client)
	if err := pty.Setsize(ws.ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		log.Printf("Failed to set terminal size: %v", err)
	}

	// Start goroutines for I/O
	go ws.readFromPTY()
	go ws.readFromWebSocket()
	go ws.waitForProcess()

	return nil
}

func (ws *WebShell) readFromPTY() {
	defer ws.cleanup()

	buf := make([]byte, 1024)
	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
			n, err := ws.ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("PTY read error: %v", err)
				}
				return
			}

			ws.mu.Lock()
			if ws.conn != nil {
				data := buf[:n]
				dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
				base64.StdEncoding.Encode(dst, data)
				// fmt.Printf("encode: %v, before: %v", dst, data)
				if err := ws.conn.WriteMessage(websocket.TextMessage, dst); err != nil {
					log.Printf("WebSocket write error: %v", err)
					ws.mu.Unlock()
					return
				}
			}
			ws.mu.Unlock()
		}
	}
}

func (ws *WebShell) readFromWebSocket() {
	defer ws.cleanup()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
			messageType, data, err := ws.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			switch messageType {
			case websocket.TextMessage:
				// Handle terminal resize messages
				if len(data) > 0 && data[0] == 1 { // Resize message marker
					if len(data) >= 5 {
						rows := uint16(data[1])<<8 | uint16(data[2])
						cols := uint16(data[3])<<8 | uint16(data[4])
						if ws.ptmx != nil {
							pty.Setsize(ws.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
						}
					}
				} else {
					dst := make([]byte, base64.StdEncoding.DecodedLen(len(string(data))))
					n, err := base64.StdEncoding.Decode(dst, data)
					if err != nil {
						fmt.Println("decode error:", err)
						return
					}
					dst = dst[:n]

					// Regular terminal input
					if ws.ptmx != nil {
						if _, err := ws.ptmx.Write(dst); err != nil {
							log.Printf("PTY write error: %v", err)
							return
						}
					}
				}
			case websocket.BinaryMessage:
				// Handle binary data (for resize or other control messages)
				if len(data) >= 5 && data[0] == 1 { // Resize message
					rows := uint16(data[1])<<8 | uint16(data[2])
					cols := uint16(data[3])<<8 | uint16(data[4])
					if ws.ptmx != nil {
						pty.Setsize(ws.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
					}
				}
			case websocket.CloseMessage:
				log.Println("WebSocket connection closed by client")
				return
			}
		}
	}
}

func (ws *WebShell) waitForProcess() {
	defer ws.cleanup()

	if ws.cmd != nil && ws.cmd.Process != nil {
		ws.cmd.Wait()
		log.Println("Shell process exited")
	}
}

func (ws *WebShell) cleanup() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Cancel context to stop all goroutines
	ws.cancel()

	// Close PTY
	if ws.ptmx != nil {
		ws.ptmx.Close()
		ws.ptmx = nil
	}

	// Kill process if still running
	if ws.cmd != nil && ws.cmd.Process != nil {
		ws.cmd.Process.Kill()
	}

	// Close WebSocket connection
	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
}

func (adminUI *CaddyAdminUI) handleWsPty(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return nil
	}

	log.Println("New WebSocket connection established")

	webShell := NewWebShell(conn)
	sh := adminUI.Shell
	if err := webShell.Start(sh); err != nil {
		log.Printf("Failed to start web shell: %v", err)
		conn.Close()
		return nil
	}

	// Keep the connection alive until context is cancelled
	<-webShell.ctx.Done()
	log.Println("WebSocket connection closed")
	return nil
}
