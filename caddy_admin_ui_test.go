package caddy_admin_ui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func TestCaddyModule(t *testing.T) {
	adminUI := CaddyAdminUI{}
	moduleInfo := adminUI.CaddyModule()

	if moduleInfo.ID != "http.handlers.caddy_admin_ui" {
		t.Errorf("Expected module ID 'http.handlers.caddy_admin_ui', got %s", moduleInfo.ID)
	}

	if moduleInfo.New == nil {
		t.Error("Expected New function to be non-nil")
	}
}

func TestProvision(t *testing.T) {
	adminUI := CaddyAdminUI{}

	// Test that the struct can be created and has reasonable defaults
	if adminUI.IndexNames != nil {
		t.Log("IndexNames already initialized")
	}
}

func TestCalculateEtag(t *testing.T) {
	// Create a mock file info
	mockFile := &mockFileInfo{
		name:    "test.txt",
		size:    1024,
		modTime: time.Now(),
		isDir:   false,
	}

	etag := calculateEtag(mockFile)
	if etag == "" {
		t.Error("Expected non-empty etag")
	}

	// Check etag format
	if !strings.HasPrefix(etag, "\"") || !strings.HasSuffix(etag, "\"") {
		t.Errorf("Expected etag to be quoted, got %s", etag)
	}
}

func TestOpenFile(t *testing.T) {
	adminUI := CaddyAdminUI{
		logger: zap.NewNop(), // Use a no-op logger to avoid nil pointer
	}

	// Test opening non-existent file
	w := httptest.NewRecorder()
	_, err := adminUI.openFile("nonexistent.txt", w)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestNotFound(t *testing.T) {
	adminUI := CaddyAdminUI{
		logger: zap.NewNop(),
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/nonexistent", nil)

	// Create a mock next handler
	called := false
	next := caddyhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		called = true
		return nil
	})

	err := adminUI.notFound(w, r, next)
	if err != nil {
		t.Errorf("notFound returned error: %v", err)
	}

	if !called {
		t.Error("Expected next handler to be called")
	}
}

func TestParseCaddyfile(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		enableShell bool
		shell       string
		wantErr     bool
	}{
		{
			name:  "basic directive",
			input: "caddy_admin_ui",
		},
		{
			name:        "enable shell true",
			input:       "caddy_admin_ui { enable_shell true }",
			enableShell: true,
		},
		{
			name:        "enable shell false",
			input:       "caddy_admin_ui { enable_shell false }",
			enableShell: false,
		},
		{
			name:  "custom shell",
			input: "caddy_admin_ui { shell /bin/bash }",
			shell: "/bin/bash",
		},
		{
			name:        "both options",
			input:       "caddy_admin_ui { enable_shell true shell /bin/zsh }",
			enableShell: true,
			shell:       "/bin/zsh",
		},
		{
			name:    "unknown directive",
			input:   "caddy_admin_ui { unknown_directive }",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := httpcaddyfile.Helper{
				Dispenser: caddyfile.NewTestDispenser(tt.input),
			}

			handler, err := parseCaddyfile(h)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCaddyfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				adminUI, ok := handler.(*CaddyAdminUI)
				if !ok {
					t.Error("Expected handler to be *CaddyAdminUI")
					return
				}

				if tt.enableShell != adminUI.EnableShell {
					t.Errorf("EnableShell = %v, want %v", adminUI.EnableShell, tt.enableShell)
				}

				if tt.shell != "" && tt.shell != adminUI.Shell {
					t.Errorf("Shell = %v, want %v", adminUI.Shell, tt.shell)
				}
			}
		})
	}
}

func TestMapDirOpenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		path string
		want error
	}{
		{
			name: "not exist error",
			err:  os.ErrNotExist,
			path: "/nonexistent/path",
			want: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapDirOpenError(tt.err, tt.path)
			if got != tt.want {
				t.Errorf("mapDirOpenError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Test that the module can be created
	adminUI := CaddyAdminUI{}
	moduleInfo := adminUI.CaddyModule()
	if moduleInfo.ID != "http.handlers.caddy_admin_ui" {
		t.Error("Module ID not set correctly")
	}
}

// Test helpers

type mockFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
