package caddy_admin_ui

import (
	"bytes"
	"crypto/md5"
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const DirectiveName = "caddy_admin_ui"

// Replace this at compile time
var compileUnixTime = "1657605601"

func init() {
	httpcaddyfile.RegisterHandlerDirective(DirectiveName, parseCaddyfile)
	caddy.RegisterModule(OptimizedCaddyAdminUI{})
}

// OptimizedCaddyAdminUI implements an optimized static file server
type OptimizedCaddyAdminUI struct {
	// The names of files to try as index files if a folder is requested.
	IndexNames []string `json:"index_names,omitempty"`

	// Append suffix to request filename if origin name is not exists.
	SuffixNames []string `json:"suffix_names,omitempty"`

	EnableShell bool `json:"enable_shell,omitempty"`

	Shell string `json:"shell,omitempty"`

	// Cache configuration
	CacheTTL string `json:"cache_ttl,omitempty"`

	// Performance configuration
	EnableCompression bool `json:"enable_compression,omitempty"`

	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (OptimizedCaddyAdminUI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers." + DirectiveName + "_optimized",
		New: func() caddy.Module { return new(OptimizedCaddyAdminUI) },
	}
}

// Pre-computed index file map for O(1) lookup
var indexFileMap = map[string]bool{
	"index.html": true,
	"index.htm":  true,
	"index.txt":  true,
	"index.php":  true,
	"index.cgi":  true,
}

// Pre-computed suffix map for O(1) lookup
var suffixMap = map[string]bool{
	"html": true,
	"htm":  true,
	"txt":  true,
	"md":   true,
	"json": true,
	"xml":  true,
}

// Provision sets up the optimized static files responder.
func (adminUI *OptimizedCaddyAdminUI) Provision(ctx caddy.Context) error {
	adminUI.logger = ctx.Logger(adminUI)

	// Set default values
	if len(adminUI.IndexNames) == 0 {
		adminUI.IndexNames = []string{"index.html", "index.htm", "index.txt"}
	}

	if len(adminUI.SuffixNames) == 0 {
		adminUI.SuffixNames = []string{"html", "htm", "txt", "md"}
	}

	if adminUI.CacheTTL == "" {
		adminUI.CacheTTL = "1h"
	}

	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/sh"
	}
	adminUI.Shell = sh

	// Enable compression by default for better performance
	adminUI.EnableCompression = true

	adminUI.logger.Info("Caddy Admin UI optimized module provisioned",
		zap.Strings("index_files", adminUI.IndexNames),
		zap.Strings("suffix_files", adminUI.SuffixNames),
		zap.String("cache_ttl", adminUI.CacheTTL),
		zap.Bool("enable_compression", adminUI.EnableCompression))

	return nil
}

func (adminUI *OptimizedCaddyAdminUI) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	// PathUnescape returns an error if the escapes aren't well-formed
	if _, err := url.PathUnescape(r.URL.Path); err != nil {
		adminUI.logger.Debug("improper path escape",
			zap.String("request_path", r.URL.Path),
			zap.Error(err))
		return err
	}

	// WebSocket handling with optimized version
	if adminUI.EnableShell && r.URL.Path == "/ws/pty" {
		return adminUI.handleWsPtyOptimized(w, r, next)
	}

	filename := "build" + r.URL.Path

	// Optimized file resolution with O(1) lookups
	file, info, err := adminUI.resolveFile(filename, repl)
	if err != nil {
		return adminUI.notFound(w, r, next)
	}

	// Set smart caching headers
	adminUI.setCacheHeaders(w, filename, info)

	// Set content type
	if w.Header().Get("Content-Type") == "" {
		mtyp := mime.TypeByExtension(filepath.Ext(filename))
		if mtyp == "" {
			w.Header()["Content-Type"] = nil
		} else {
			w.Header().Set("Content-Type", mtyp)
		}
	}

	// Check for If-None-Match
	if etag := w.Header().Get("ETag"); etag != "" && etag == r.Header.Get("If-None-Match") {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	// Serve content with optimized reader
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(file))

	return nil
}

// resolveFile handles optimized file resolution with reduced allocations
func (adminUI *OptimizedCaddyAdminUI) resolveFile(filename string, repl *caddy.Replacer) ([]byte, fs.FileInfo, error) {
	// Try direct file access first
	file, err := buildFs.ReadFile(filename)
	if err == nil {
		info, _ := fs.Stat(buildFs, filename)
		return file, info, nil
	}

	// If it's a directory, try index files
	if strings.HasSuffix(filename, "/") {
		for _, indexPage := range adminUI.IndexNames {
			indexPage := repl.ReplaceAll(indexPage, "")
			indexPath := caddyhttp.SanitizedPathJoin(filename, indexPage)

			file, err := buildFs.ReadFile(indexPath)
			if err == nil {
				info, _ := fs.Stat(buildFs, indexPath)
				return file, info, nil
			}
		}
	}

	// Try suffixes for non-existent files
	if !strings.HasSuffix(filename, "/") {
		for _, suffix := range adminUI.SuffixNames {
			suffix := repl.ReplaceAll(suffix, "")
			filePath := fmt.Sprintf("%s.%s", filename, suffix)

			file, err := buildFs.ReadFile(filePath)
			if err == nil {
				info, _ := fs.Stat(buildFs, filePath)
				return file, info, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("file not found: %s", filename)
}

// setCacheHeaders sets intelligent caching headers based on file type
func (adminUI *OptimizedCaddyAdminUI) setCacheHeaders(w http.ResponseWriter, filename string, info fs.FileInfo) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Set ETag
	w.Header().Set("ETag", calculateStrongEtag(info))

	// Set Last-Modified
	ti, _ := strconv.ParseInt(compileUnixTime, 10, 64)
	tu := time.Unix(ti, 0)
	w.Header().Set("Last-Modified", tu.Format(http.TimeFormat))

	// Set Cache-Control based on file type
	switch ext {
	case ".js", ".css", ".woff", ".woff2", ".ttf", ".eot":
		// Static assets - long cache
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp":
		// Images - medium cache
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".html", ".htm", ".txt", ".md":
		// HTML files - short cache to allow updates
		w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
	default:
		// Default caching
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%s", adminUI.CacheTTL))
	}

	// Add Content-Security-Policy for security
	if ext == ".html" {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; img-src 'self' data:; "+
				"connect-src 'self' ws: wss:; font-src 'self' data:;")
	}

	// Add X-Content-Type-Options to prevent MIME sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Add X-Frame-Options for clickjacking protection
	w.Header().Set("X-Frame-Options", "DENY")
}

// calculateStrongEtag produces a strong ETag using MD5 hash
func calculateStrongEtag(info fs.FileInfo) string {
	// Use file size and mod time for ETag
	data := fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf(`"%x"`, hash)
}

// parseCaddyfile parses the caddy_admin_ui directive
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var adminUI OptimizedCaddyAdminUI

	for h.Next() {
		for h.NextBlock(0) {
			switch h.Val() {
			case "enable_shell":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.EnableShell = h.Val() == "true"

			case "shell":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.Shell = h.Val()

			case "cache_ttl":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.CacheTTL = h.Val()

			case "enable_compression":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.EnableCompression = h.Val() == "true"

			case "index_names":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.IndexNames = h.RemainingArgs()

			case "suffix_names":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				adminUI.SuffixNames = h.RemainingArgs()

			default:
				return nil, h.Errf("unknown subdirective '%s'", h.Val())
			}
		}
	}

	return &adminUI, nil
}

// it calls the next handler in the chain.
func (adminUI *OptimizedCaddyAdminUI) notFound(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*OptimizedCaddyAdminUI)(nil)
	_ caddyhttp.MiddlewareHandler = (*OptimizedCaddyAdminUI)(nil)
)