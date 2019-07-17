package arry

import (
	"net"
	"path"
	"net/http"
	"crypto/tls"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type Handler func(Context)

type Middleware func(Handler) Handler


type Arry struct {
	router *Router
	middlewares []Middleware
	Engine Engine

	Server *http.Server
	// DefaultRoute Handler
}

func New() *Arry {
	r := NewRouter()
	r.DefaultHandler(defaultHandler)

	arry := &Arry{
		router: r,
	}

	return arry
}

func (a *Arry) Router() *Router {
	return a.router
}


func (a *Arry) Use(middleware Middleware) {
	a.middlewares = append(a.middlewares, middleware)
}

func (a *Arry) Views(path string) {
	a.Engine = NewEngine(path, "html")
}

func (a *Arry) Static(url string, dir string) {
	h := func(c Context) {
		p := path.Join(dir, c.Param("*"))
		c.File(p)
	}
	a.router.Get(url + "/*", h)
}

func (a *Arry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)
	ctx.SetEngine(a.Engine)

	n := a.router.Route(ctx.Request().URL.Path, ctx)

	h := a.router.handler

	if n != nil {
		if hd, ok := n.methods[r.Method]; ok {
			h = hd
		}
	}

	h = applyMiddlewares(h, a.middlewares)
	h(ctx)
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
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	if s.TLSConfig != nil {
		l = tls.NewListener(l, s.TLSConfig)
	}

	return s.Serve(l)
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
