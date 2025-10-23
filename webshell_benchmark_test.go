package caddy_admin_ui

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// BenchmarkWebSocketBase64 benchmarks the old base64 approach
func BenchmarkWebSocketBase64(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Simulate PTY data with base64 encoding
		testData := []byte("Hello, World! This is a test message for benchmarking.\n")
		for i := 0; i < b.N; i++ {
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(testData)))
			base64.StdEncoding.Encode(dst, testData)
			conn.WriteMessage(websocket.TextMessage, dst)
		}
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Read messages
	for i := 0; i < b.N; i++ {
		_, _, err := conn.ReadMessage()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWebSocketBinary benchmarks the new binary protocol
func BenchmarkWebSocketBinary(b *testing.B) {
	// Create test server with binary protocol
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Simulate PTY data with binary protocol
		testData := []byte("Hello, World! This is a test message for benchmarking.\n")
		for i := 0; i < b.N; i++ {
			message := make([]byte, 1+len(testData))
			message[0] = msgTypeData
			copy(message[1:], testData)
			conn.WriteMessage(websocket.BinaryMessage, message)
		}
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Read messages
	for i := 0; i < b.N; i++ {
		_, _, err := conn.ReadMessage()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWebSocketBuffered benchmarks buffered writes
func BenchmarkWebSocketBuffered(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Send many small messages
		testData := []byte("x")
		for i := 0; i < b.N; i++ {
			message := make([]byte, 1+len(testData))
			message[0] = msgTypeData
			copy(message[1:], testData)
			conn.WriteMessage(websocket.BinaryMessage, message)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := conn.ReadMessage()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentConnections tests connection handling under load
func BenchmarkConcurrentConnections(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		// Simple echo server
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(messageType, data)
		}
	}))
	defer server.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Run with multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send and receive one message
			conn.WriteMessage(websocket.TextMessage, []byte("test"))
			conn.ReadMessage()
		}()
	}
	wg.Wait()
}

// BenchmarkRateLimiter benchmarks the rate limiter
func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000, 100) // 1000 req/s, burst 100
	clientIP := "192.168.1.1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		limiter.Allow(clientIP)
	}
}

// BenchmarkFileResolution benchmarks the optimized file resolution
func BenchmarkFileResolution(b *testing.B) {
	adminUI := &OptimizedCaddyAdminUI{}
	adminUI.IndexNames = []string{"index.html", "index.htm"}
	adminUI.SuffixNames = []string{"html", "htm"}

	// Mock replacer
	repl := &caddy.Replacer{}
	ctx := context.WithValue(context.Background(), caddy.ReplacerCtxKey, repl)

	// Create mock request
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)
	next := caddyhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		adminUI.resolveFile("build/test", repl)
	}
}

// BenchmarkETagCalculation benchmarks ETag generation
func BenchmarkETagCalculation(b *testing.B) {
	// Create mock file info
	info := &mockFileInfo{
		size:    1024,
		modTime: time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		calculateStrongEtag(info)
	}
}

// Mock file info for testing
type mockFileInfo struct {
	size    int64
	modTime time.Time
	name    string
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *mockFileInfo) IsDir() bool        { return false }
func (fi *mockFileInfo) Sys() interface{}   { return nil }

// TestWebSocketMessageProtocol tests the binary message protocol
func TestWebSocketMessageProtocol(t *testing.T) {
	tests := []struct {
		name     string
		msgType  byte
		payload  []byte
		expected []byte
	}{
		{"Data Message", msgTypeData, []byte("hello"), []byte{0x00, 'h', 'e', 'l', 'l', 'o'}},
		{"Resize Message", msgTypeResize, []byte{0x00, 0x18, 0x00, 0x50}, []byte{0x01, 0x00, 0x18, 0x00, 0x50}},
		{"Ping Message", msgTypePing, nil, []byte{0x02}},
		{"Pong Message", msgTypePong, nil, []byte{0x03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := make([]byte, 1+len(tt.payload))
			message[0] = tt.msgType
			if len(tt.payload) > 0 {
				copy(message[1:], tt.payload)
			}

			if !bytes.Equal(message, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, message)
			}
		})
	}
}

// TestBufferPool tests the buffer pool implementation
func TestBufferPool(t *testing.T) {
	// Get buffer from pool
	buf1 := bufferPool.Get().([]byte)
	if len(buf1) != readBufferSize {
		t.Errorf("Expected buffer size %d, got %d", readBufferSize, len(buf1))
	}

	// Modify buffer
	buf1[0] = 0x42

	// Return buffer to pool
	bufferPool.Put(buf1)

	// Get another buffer
	buf2 := bufferPool.Get().([]byte)
	if buf2[0] != 0x42 {
		t.Error("Buffer pool should reuse buffers")
	}

	// Return buffer
	bufferPool.Put(buf2)
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	limiter := NewRateLimiter(10, 20) // 10 req/s, burst 20
	clientIP := "192.168.1.1"

	// Should allow burst of 20 requests
	allowed := 0
	for i := 0; i < 25; i++ {
		if limiter.Allow(clientIP) {
			allowed++
		}
	}

	if allowed != 20 {
		t.Errorf("Expected 20 allowed requests, got %d", allowed)
	}

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should allow more requests
	if !limiter.Allow(clientIP) {
		t.Error("Should allow request after refill")
	}
}