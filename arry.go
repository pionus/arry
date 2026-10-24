package arry

import (
	"crypto/tls"
	stdContext "context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type Handler func(Context)

type Middleware func(Handler) Handler


type Arry struct {
	router *Router
	middlewares []Middleware
	graceful bool
	Engine Engine

	Server *http.Server
	// DefaultRoute Handler
}

func New() *Arry {
	r := NewRouter()

	arry := &Arry{
		router: r,
		graceful: true,
	}

	return arry
}

func (a *Arry) Router() *Router {
	return a.router
}


func (a *Arry) Use(middleware Middleware) {
	a.middlewares = append(a.middlewares, middleware)
}

func (a *Arry) Views(dir string) {
	a.Engine = NewEngine(dir, "html")
}

// StaticConfig configures a static file route.
type StaticConfig struct {
	// Dir is the directory served by the route.
	Dir string
	// MaxAge sets the Cache-Control max-age for served files. Zero disables it.
	MaxAge time.Duration
	// Index is the file served for directory requests (e.g. "index.html").
	// Empty means directory requests return 404.
	Index string
}

// Static serves files from dir under url. Path traversal outside dir is rejected.
func (a *Arry) Static(url string, dir string) {
	a.StaticWithConfig(url, StaticConfig{Dir: dir})
}

// StaticWithConfig serves files from cfg.Dir under url with optional settings.
func (a *Arry) StaticWithConfig(url string, cfg StaticConfig) {
	base, err := filepath.Abs(cfg.Dir)
	if err != nil {
		panic(fmt.Sprintf("arry: static dir %q: %v", cfg.Dir, err))
	}

	h := func(c Context) {
		rel := path.Clean("/" + c.Param("*"))
		name := filepath.Join(base, filepath.FromSlash(rel))

		if !isSubPath(base, name) {
			c.Reply(http.StatusNotFound)
			return
		}

		fi, err := os.Stat(name)
		if err != nil {
			c.Reply(http.StatusNotFound)
			return
		}
		if fi.IsDir() {
			if cfg.Index == "" {
				c.Reply(http.StatusNotFound)
				return
			}
			name = filepath.Join(name, cfg.Index)
			if !isSubPath(base, name) {
				c.Reply(http.StatusNotFound)
				return
			}
			if _, err := os.Stat(name); err != nil {
				c.Reply(http.StatusNotFound)
				return
			}
		}

		if cfg.MaxAge > 0 {
			c.SetHeader("Cache-Control", fmt.Sprintf("public, max-age=%d", int64(cfg.MaxAge/time.Second)))
		}

		c.File(name)
	}

	a.router.Get(url+"/*", h)
	a.router.Head(url+"/*", h)
}

// StaticFS serves the given filesystem under url. It is useful for assets
// embedded with go:embed.
func (a *Arry) StaticFS(url string, fsys fs.FS) {
	if fsys == nil {
		panic("arry: StaticFS called with nil filesystem")
	}

	prefix := strings.TrimSuffix(url, "/")
	if prefix == "" {
		prefix = "/"
	}

	h := http.StripPrefix(prefix, http.FileServerFS(fsys))
	handler := func(c Context) {
		h.ServeHTTP(c.Response(), c.Request())
	}

	a.router.Get(prefix+"/*", handler)
	a.router.Head(prefix+"/*", handler)
}

func (a *Arry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)
	ctx.SetEngine(a.Engine)

	n := a.router.route(ctx.Request().URL.Path, ctx)

	h := a.router.Handler()
	headFallback := false

	if n != nil {
		if hd, ok := n.methods[r.Method]; ok {
			h = hd
		} else if r.Method == http.MethodHead {
			if hd, ok := n.methods[http.MethodGet]; ok {
				h = hd
				headFallback = true
			}
		} else if r.Method == http.MethodOptions && len(n.methods) > 0 {
			allow := allowHeader(n.methods)
			h = func(c Context) {
				c.SetHeader("Allow", allow)
				c.NoContent(http.StatusNoContent)
			}
		} else if len(n.methods) > 0 {
			allow := allowHeader(n.methods)
			h = func(c Context) {
				c.SetHeader("Allow", allow)
				c.Reply(http.StatusMethodNotAllowed)
			}
		}
	}

	if headFallback {
		ctx.Response().Writer = headResponseWriter{ResponseWriter: ctx.Response().Writer}
	}

	h = applyMiddlewares(h, a.middlewares)
	h(ctx)
}

// isSubPath reports whether target is base or inside base.
func isSubPath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func allowHeader(methods map[string]Handler) string {
	seen := make(map[string]bool, len(methods)+2)
	for m := range methods {
		seen[m] = true
	}
	seen[http.MethodHead] = true
	seen[http.MethodOptions] = true

	ordered := []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
	}
	var allowed []string
	for _, m := range ordered {
		if seen[m] {
			allowed = append(allowed, m)
		}
	}
	sort.Strings(allowed)
	return strings.Join(allowed, ", ")
}

// headResponseWriter discards the body while preserving headers and status.
type headResponseWriter struct {
	http.ResponseWriter
}

func (w headResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w headResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (a *Arry) Graceful(enable bool) {
	a.graceful = enable
}

func (a *Arry) Start(addr string) error {
	a.Server = &http.Server{
		Addr: addr,
		Handler: a,
	}
	return a.StartServer(a.Server)
}

func (a *Arry) StartTLS(addr string, domains... string) error {
	certManager := autocert.Manager{
        Prompt: autocert.AcceptTOS,
        HostPolicy: autocert.HostWhitelist(domains...),
		Cache: autocert.DirCache("certs"),
    }

	a.Server = &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
            GetCertificate: certManager.GetCertificate,
        },
		Handler: a,
	}

	a.Server.TLSConfig.NextProtos = append(a.Server.TLSConfig.NextProtos, acme.ALPNProto, "h2")

	return a.StartServer(a.Server)
}

func (a *Arry) StartServer(s *http.Server) error {
	if !a.graceful {
		return a.serve(s)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	errs := make(chan error, 1)
	
	go func() {
		err := a.serve(s)
		if err != nil {
			errs <- err
		}
	}()
	
	select {
	case err := <-errs:
		log.Println(err)
		return err
	case <-quit:
		log.Println("Graceful shutting down...")
		err := a.Shutdown(stdContext.Background())
		return err
	}
}

func (a *Arry) serve(s *http.Server) error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	log.Printf("Listening at %s", s.Addr)

	if s.TLSConfig != nil {
		l = tls.NewListener(l, s.TLSConfig)
	}

	return s.Serve(l)
}

func (a *Arry) Close() error {
	return a.Server.Close()
}

func (a *Arry) Shutdown(ctx stdContext.Context) error {
	return a.Server.Shutdown(ctx)
}


func applyMiddlewares(h Handler, middlewares []Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}


func defaultHandler(ctx Context) {
	ctx.Reply(404)
}
