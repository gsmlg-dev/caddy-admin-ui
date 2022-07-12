package caddy_admin_ui

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	weakrand "math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
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
	weakrand.Seed(time.Now().UnixNano())

	caddy.RegisterModule(CaddyAdminUI{})
}

// CaddyAdminUI implements a static file server responder for Caddy.
type CaddyAdminUI struct {
	// The names of files to try as index files if a folder is requested.
	IndexNames []string `json:"index_names,omitempty"`

	// Append suffix to request filename if origin name is not exists.
	SuffixNames []string `json:"suffix_names,omitempty"`

	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (CaddyAdminUI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers." + DirectiveName,
		New: func() caddy.Module { return new(CaddyAdminUI) },
	}
}

// Provision sets up the static files responder.
func (adminUI *CaddyAdminUI) Provision(ctx caddy.Context) error {
	adminUI.logger = ctx.Logger(adminUI)

	adminUI.IndexNames = []string{"index.html", "index.htm", "index.txt"}

	adminUI.SuffixNames = []string{"html", "htm", "txt"}

	return nil
}

//go:embed build/*
var buildFs embed.FS

func (adminUI *CaddyAdminUI) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	// PathUnescape returns an error if the escapes aren't well-formed,
	// meaning the count % matches the RFC. Return early if the escape is
	// improper.
	if _, err := url.PathUnescape(r.URL.Path); err != nil {
		adminUI.logger.Debug("improper path escape",
			zap.String("request_path", r.URL.Path),
			zap.Error(err))
		return err
	}
	filename := "build" + r.URL.Path

	adminUI.logger.Debug("sanitized path join",
		zap.String("request_path", r.URL.Path),
		zap.String("result", filename))

	// get information about the file
	opF, err := buildFs.Open(filename)
	if err != nil {
		adminUI.logger.Debug("open file error",
			zap.String("error", err.Error()),
			zap.String("File", fmt.Sprintf("%v", opF)),
			zap.String("FileSystem", fmt.Sprintf("%v", buildFs)))
		err = mapDirOpenError(err, filename)
		if os.IsNotExist(err) {
			var info fs.FileInfo
			if len(adminUI.IndexNames) > 0 {
				for _, indexPage := range adminUI.IndexNames {
					indexPage := repl.ReplaceAll(indexPage, "")
					indexPath := caddyhttp.SanitizedPathJoin(filename, indexPage)

					opF, err = buildFs.Open(indexPath)
					if err != nil {
						continue
					}
					info, _ = opF.Stat()
					filename = indexPath
					// implicitIndexFile = true
					adminUI.logger.Debug("located index file", zap.String("filename", filename))
					break
				}
			}
			if info == nil && strings.HasSuffix(filename, "/") == false {
				suffixList := []string{"html", "htm", "txt"}
				for _, suffix := range suffixList {
					suffix := repl.ReplaceAll(suffix, "")
					filePath := fmt.Sprintf("%s.%s", filename, suffix)

					opF, err = buildFs.Open(filePath)
					if err != nil {
						continue
					}
					info, _ = opF.Stat()
					filename = filePath
					// implicitIndexFile = true
					adminUI.logger.Debug("located file with suffix", zap.String("filename", filename), zap.String("suffix", suffix))
					break
				}
			}
			if info == nil {
				return adminUI.notFound(w, r, next)
			}
		} else if os.IsPermission(err) {
			return caddyhttp.Error(http.StatusForbidden, err)
		}
	}
	info, err := opF.Stat()
	if err != nil {
		return caddyhttp.Error(http.StatusInternalServerError, err)
	}

	// if the request mapped to a directory, see if
	// there is an index file we can serve
	var implicitIndexFile bool
	if info.IsDir() && len(adminUI.IndexNames) > 0 {
		for _, indexPage := range adminUI.IndexNames {
			indexPage := repl.ReplaceAll(indexPage, "")
			indexPath := caddyhttp.SanitizedPathJoin(filename, indexPage)

			opF, err := buildFs.Open(indexPath)
			if err != nil {
				continue
			}
			indexInfo, _ := opF.Stat()

			// don't rewrite the request path to append
			// the index file, because we might need to
			// do a canonical-URL redirect below based
			// on the URL as-is

			// we've chosen to use this index file,
			// so replace the last file info and path
			// with that of the index file
			info = indexInfo
			filename = indexPath
			implicitIndexFile = true
			adminUI.logger.Debug("located index file", zap.String("filename", filename))
			break
		}
	}

	// if still referencing a directory, delegate
	// to browse or return an error
	if info.IsDir() {
		adminUI.logger.Debug("no index file in directory",
			zap.String("path", filename),
			zap.Strings("index_filenames", adminUI.IndexNames))
		return adminUI.notFound(w, r, next)
	}

	// if URL canonicalization is enabled, we need to enforce trailing
	// slash convention: if a directory, trailing slash; if a file, no
	// trailing slash - not enforcing this can break relative hrefs
	// in HTML (see https://github.com/caddyserver/caddy/issues/2741)
	if info == nil {
		// Only redirect if the last element of the path (the filename) was not
		// rewritten; if the admin wanted to rewrite to the canonical path, they
		// would have, and we have to be very careful not to introduce unwanted
		// redirects and especially redirect loops!
		// See https://github.com/caddyserver/caddy/issues/4205.
		origReq := r.Context().Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
		if path.Base(origReq.URL.Path) == path.Base(r.URL.Path) {
			if implicitIndexFile && !strings.HasSuffix(origReq.URL.Path, "/") {
				to := origReq.URL.Path + "/"
				adminUI.logger.Debug("redirecting to canonical URI (adding trailing slash for directory)",
					zap.String("from_path", origReq.URL.Path),
					zap.String("to_path", to))
				return redirect(w, r, to)
			} else if !implicitIndexFile && strings.HasSuffix(origReq.URL.Path, "/") {
				to := origReq.URL.Path[:len(origReq.URL.Path)-1]
				adminUI.logger.Debug("redirecting to canonical URI (removing trailing slash for file)",
					zap.String("from_path", origReq.URL.Path),
					zap.String("to_path", to))
				return redirect(w, r, to)
			}
		}
	}

	var file []byte

	// no precompressed file found, use the actual file
	if file == nil {
		adminUI.logger.Debug("opening file", zap.String("filename", filename))

		// open the file
		file, err = adminUI.openFile(filename, w)
		if err != nil {
			if herr, ok := err.(caddyhttp.HandlerError); ok &&
				herr.StatusCode == http.StatusNotFound {
				return adminUI.notFound(w, r, next)
			}
			return err // error is already structured
		}
	}

	// set the ETag - note that a conditional If-None-Match request is handled
	// by http.ServeContent below, which checks against this ETag value
	w.Header().Set("ETag", calculateEtag(info))

	// set last modify since
	ti, _ := strconv.ParseInt(compileUnixTime, 10, 64)
	tu := time.Unix(ti, 0)
	w.Header().Set("Last-Modified-Since", tu.Format("Mon, 02 Jan 2006 15:04:05 GMT"))

	if w.Header().Get("Content-Type") == "" {
		mtyp := mime.TypeByExtension(filepath.Ext(filename))
		if mtyp == "" {
			// do not allow Go to sniff the content-type; see
			// https://www.youtube.com/watch?v=8t8JYpt0egE
			// TODO: If we want a Content-Type, consider writing a default of application/octet-stream - this is secure but violates spec
			w.Header()["Content-Type"] = nil
		} else {
			w.Header().Set("Content-Type", mtyp)
		}
	}

	// let the standard library do what it does best; note, however,
	// that errors generated by ServeContent are written immediately
	// to the response, so we cannot handle them (but errors there
	// are rare)
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(file))

	return nil
}

// openFile opens the file at the given filename. If there was an error,
// the response is configured to inform the client how to best handle it
// and a well-described handler error is returned (do not wrap the
// returned error value).
func (adminUI *CaddyAdminUI) openFile(filename string, w http.ResponseWriter) ([]byte, error) {
	file, err := buildFs.ReadFile(filename)
	if err != nil {
		err = mapDirOpenError(err, filename)
		if os.IsNotExist(err) {
			adminUI.logger.Debug("file not found", zap.String("filename", filename), zap.Error(err))
			return nil, caddyhttp.Error(http.StatusNotFound, err)
		}
		return nil, caddyhttp.Error(http.StatusServiceUnavailable, err)
	}
	return file, nil
}

// mapDirOpenError maps the provided non-nil error from opening name
// to a possibly better non-nil error. In particular, it turns OS-specific errors
// about opening files in non-directories into os.ErrNotExist. See golang/go#18984.
// Adapted from the Go standard library; originally written by Nathaniel Caza.
// https://go-review.googlesource.com/c/go/+/36635/
// https://go-review.googlesource.com/c/go/+/36804/
func mapDirOpenError(originalErr error, name string) error {
	if os.IsNotExist(originalErr) {
		return originalErr
	}

	parts := strings.Split(name, separator)
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		fi, err := os.Stat(strings.Join(parts[:i+1], separator))
		if err != nil {
			return originalErr
		}
		if !fi.IsDir() {
			return os.ErrNotExist
		}
	}

	return originalErr
}

// it calls the next handler in the chain.
func (fsrv *CaddyAdminUI) notFound(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(w, r)
}

// parseCaddyfile parses the caddy_admin_ui directive. It enables the static file
// server and configures it with this syntax:
//
//    caddy_admin_ui
//
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var adminUI CaddyAdminUI

	for h.Next() {
		for h.NextBlock(0) {
			switch h.Val() {

			default:
				return nil, h.Errf("unknown subdirective '%s'", h.Val())
			}
		}
	}

	return &adminUI, nil
}

// calculateEtag produces a strong etag by default, although, for
// efficiency reasons, it does not actually consume the contents
// of the file to make a hash of all the bytes. ¯\_(ツ)_/¯
// Prefix the etag with "W/" to convert it into a weak etag.
// See: https://tools.ietf.org/html/rfc7232#section-2.3
func calculateEtag(d os.FileInfo) string {
	ti, _ := strconv.ParseInt(compileUnixTime, 10, 64)
	t := strconv.FormatInt(ti, 36)
	s := strconv.FormatInt(d.Size(), 36)
	return `"` + t + "_" + s + `"`
}

func redirect(w http.ResponseWriter, r *http.Request, to string) error {
	for strings.HasPrefix(to, "//") {
		// prevent path-based open redirects
		to = strings.TrimPrefix(to, "/")
	}
	http.Redirect(w, r, to, http.StatusPermanentRedirect)
	return nil
}

const (
	separator = string(filepath.Separator)
)

// Interface guards
var (
	_ caddy.Provisioner           = (*CaddyAdminUI)(nil)
	_ caddyhttp.MiddlewareHandler = (*CaddyAdminUI)(nil)
)
