package arry

import (
    // "net/http"
    "strings"
    "path"
    // "fmt"
)

type node struct {
    children    []*node
    component   string
    isParam     bool
    methods     map[string]Handler
}

func (n *node) traverse(components []string, ctx Context) (*node, []string) {
    if len(components) == 0 {
        return n, components
    }

    component := components[0]

    if len(component) == 0 {
        return n, components[1:]
    }

    if len(n.children) == 0 {
        return n, components
    }

    next := components[1:]

    for _, child := range n.children {
        if child.component[0] == ':' {
            setParams(child.component[1:], component, ctx)
            return child.traverse(next, ctx)
        }

        if child.component[0] == '*' {
            setParams("*", path.Join(components...), ctx)
            return child, next[0:0]
        }

        if child.component == component {
            return child.traverse(next, ctx)
        }
    }

    // default return
    return n, components
}


func (n *node) add(method string, pattern string, handler Handler) {
    paths := strings.Split(pattern, "/")[1:]

    child, components := n.traverse(paths, nil)
    child = child.pave(components)
    child.methods[strings.ToUpper(method)] = handler
}


func (n *node) addChild(pattern string, child *node) {
    paths := strings.Split(pattern, "/")[1:]

    end, components := n.traverse(paths, nil)
    end = end.pave(components)
    
    end.isParam = child.isParam
    end.children = child.children
    end.methods = child.methods
}


func (n *node) pave(components []string) *node {
    if len(components) == 0 {
        return n
    }

    component := components[0]

    child := node{
        component: component,
        isParam: false,
        methods: make(map[string]Handler),
    }

    if len(component) > 0 && component[0] == ':' {
        child.isParam = true
    }

    n.children = append(n.children, &child)

    return child.pave(components[1:])
}

func setParams(key string, value string, ctx Context) {
    if ctx == nil {
        return
    }

    c, ok := ctx.(*context)
    if !ok {
        return
    }

    if c.params == nil {
        c.params = make(map[string]string)
    }

    c.params[key] = value
}

type Router struct {
    tree        *node
    handler     Handler
    middlewares []Middleware
}


func (r *Router) Handle(method string, pattern string, handler Handler) {
    // Apply current layer middlewares
    if len(r.middlewares) > 0 {
        handler = applyMiddlewares(handler, r.middlewares)
    }
    r.tree.add(method, pattern, handler)
}

func (r *Router) Get(pattern string, handler Handler) {
    r.Handle("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler Handler) {
    r.Handle("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler Handler) {
    r.Handle("PUT", pattern, handler)
}

// Delete registers a DELETE route on the router
func (r *Router) Delete(pattern string, handler Handler) {
    r.Handle("DELETE", pattern, handler)
}

// Patch registers a PATCH route on the router
func (r *Router) Patch(pattern string, handler Handler) {
    r.Handle("PATCH", pattern, handler)
}

// Options registers an OPTIONS route on the router
func (r *Router) Options(pattern string, handler Handler) {
    r.Handle("OPTIONS", pattern, handler)
}

// Head registers a HEAD route on the router
func (r *Router) Head(pattern string, handler Handler) {
    r.Handle("HEAD", pattern, handler)
}

// Any registers a route that matches all HTTP methods on the router
func (r *Router) Any(pattern string, handler Handler) {
    methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
    for _, method := range methods {
        r.Handle(method, pattern, handler)
    }
}

// Match registers a route that matches the specified HTTP methods on the router
func (r *Router) Match(methods []string, pattern string, handler Handler) {
    for _, method := range methods {
        r.Handle(method, pattern, handler)
    }
}

func (r *Router) Graft(pattern string, subRouter *Router) {
    // If parent router has middlewares, apply them to all handlers in the sub-router's tree
    if len(r.middlewares) > 0 {
        r.applyMiddlewaresToSubtree(subRouter.tree, r.middlewares)
    }

    r.tree.addChild(pattern, subRouter.tree)
}

// applyMiddlewaresToSubtree recursively applies middlewares to all handlers in the node tree
func (r *Router) applyMiddlewaresToSubtree(n *node, middlewares []Middleware) {
    // Apply middlewares to all HTTP methods in current node
    for method, handler := range n.methods {
        n.methods[method] = applyMiddlewares(handler, middlewares)
    }

    // Recursively apply to child nodes
    for _, child := range n.children {
        r.applyMiddlewaresToSubtree(child, middlewares)
    }
}


func (r *Router) route(url string, ctx Context) *node {
    paths := strings.Split(url, "/")[1:]
    n, components := r.tree.traverse(paths, ctx)

    if len(components) == 0 {
        return n
    }

    return nil
}

// Handler returns the default handler
func (r *Router) Handler() Handler {
    return r.handler
}


func NewRouter(middlewares ...Middleware) *Router {
    node := node{
        component: "/",
        isParam: false,
        methods: make(map[string]Handler),
    }

    router := &Router{
        tree:        &node,
        handler:     defaultHandler,
        middlewares: middlewares,
    }

    return router
}
