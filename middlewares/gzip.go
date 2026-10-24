package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pionus/arry"
)

// GzipConfig configures the gzip middleware.
type GzipConfig struct {
	// Level is the gzip compression level. Defaults to gzip.DefaultCompression.
	Level int
	// MinSize is the minimum response size in bytes before compression kicks
	// in. Defaults to 1024. Set to 0 to compress every compressible response.
	MinSize int
}

// DefaultGzipConfig is the default gzip configuration.
var DefaultGzipConfig = GzipConfig{
	Level:   5,
	MinSize: 1024,
}

// Gzip compresses responses with gzip when the client supports it and the
// response is worth compressing.
func Gzip(next arry.Handler) arry.Handler {
	return GzipWithConfig(DefaultGzipConfig)(next)
}

// GzipWithConfig returns a gzip middleware with the given configuration.
func GzipWithConfig(config GzipConfig) arry.Middleware {
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}
	if config.MinSize < 0 {
		config.MinSize = DefaultGzipConfig.MinSize
	}

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			if ctx.Request().Method == http.MethodHead ||
				!acceptsGzip(ctx.Header("Accept-Encoding")) ||
				ctx.Header("Content-Encoding") != "" {
				next(ctx)
				return
			}

			gw := &gzipResponseWriter{
				ResponseWriter: ctx.Response().Writer,
				level:          config.Level,
				minSize:        config.MinSize,
			}
			ctx.Response().Writer = gw

			// Do not write anything if the handler panicked: the panic
			// middleware outside this one is responsible for that response.
			defer func() {
				if p := recover(); p != nil {
					gw.aborted = true
					panic(p)
				}
				gw.Close()
			}()

			next(ctx)
		}
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	level   int
	minSize int
	status  int
	buf     []byte
	wrote   bool
	aborted bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	// Delay the actual header until the body size is known. Status codes
	// without a body are flushed immediately.
	if code == http.StatusNoContent || code == http.StatusNotModified {
		w.writeHeader(code, false)
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.aborted {
		return 0, io.ErrClosedPipe
	}
	if !w.wrote {
		if w.status == 0 {
			w.status = http.StatusOK
		}

		if !w.shouldCompress() {
			w.writeHeader(w.status, false)
		} else if len(w.buf)+len(b) < w.minSize {
			w.buf = append(w.buf, b...)
			return len(b), nil
		} else {
			w.writeHeader(w.status, true)
			if len(w.buf) > 0 {
				if _, err := w.gz.Write(w.buf); err != nil {
					return 0, err
				}
				w.buf = nil
			}
		}
	}

	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// Flush disables compression when it is called before the response is
// committed, then flushes the underlying writer. This keeps streaming
// responses usable.
func (w *gzipResponseWriter) Flush() {
	if !w.wrote {
		if w.status == 0 {
			w.status = http.StatusOK
		}
		w.writeHeader(w.status, false)
		if len(w.buf) > 0 {
			w.ResponseWriter.Write(w.buf)
			w.buf = nil
		}
	}
	if w.gz != nil {
		w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Close commits a buffered response and finishes the gzip stream.
func (w *gzipResponseWriter) Close() error {
	if w.aborted {
		return nil
	}
	if !w.wrote {
		if w.status == 0 {
			w.status = http.StatusOK
		}
		compress := w.shouldCompress() && len(w.buf) >= w.minSize && len(w.buf) > 0
		w.writeHeader(w.status, compress)
	}

	if w.gz != nil {
		if len(w.buf) > 0 {
			w.gz.Write(w.buf)
			w.buf = nil
		}
		err := w.gz.Close()
		w.gz = nil
		return err
	}

	if len(w.buf) > 0 {
		_, err := w.ResponseWriter.Write(w.buf)
		w.buf = nil
		return err
	}
	return nil
}

func (w *gzipResponseWriter) writeHeader(code int, compress bool) {
	if w.wrote {
		return
	}
	w.wrote = true

	header := w.Header()
	if compress {
		header.Del("Content-Length")
		header.Set("Content-Encoding", "gzip")
		addVary(header, "Accept-Encoding")

		gz, err := gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err != nil {
			// Invalid level: fall back to an uncompressed response.
			header.Del("Content-Encoding")
			w.ResponseWriter.WriteHeader(code)
			return
		}
		w.gz = gz
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) shouldCompress() bool {
	switch {
	case w.status == http.StatusNoContent,
		w.status == http.StatusNotModified,
		w.status == http.StatusPartialContent,
		w.status < 200:
		return false
	}

	ct := w.Header().Get("Content-Type")
	if ct == "" {
		sample := w.buf
		if len(sample) > 512 {
			sample = sample[:512]
		}
		if len(sample) > 0 {
			ct = http.DetectContentType(sample)
		}
	}
	if ct == "" {
		return false
	}
	return compressibleType(ct)
}

func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(part, ";")
		if !strings.EqualFold(strings.TrimSpace(fields[0]), "gzip") {
			continue
		}
		q := 1.0
		for _, f := range fields[1:] {
			f = strings.TrimSpace(f)
			if strings.HasPrefix(f, "q=") {
				if v, err := strconv.ParseFloat(f[2:], 64); err == nil {
					q = v
				}
			}
		}
		return q > 0
	}
	return false
}

func compressibleType(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))

	if strings.HasPrefix(ct, "text/") {
		return true
	}
	if strings.HasSuffix(ct, "+json") || strings.HasSuffix(ct, "+xml") {
		return true
	}

	switch ct {
	case "application/json",
		"application/javascript",
		"application/x-javascript",
		"application/ecmascript",
		"application/xml",
		"application/xhtml+xml",
		"application/rss+xml",
		"application/atom+xml",
		"application/manifest+json",
		"application/wasm",
		"image/svg+xml":
		return true
	}
	return false
}

func addVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
