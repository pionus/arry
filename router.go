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
    tree    *node
    handler Handler
}


func (r *Router) Handle(method string, pattern string, handler Handler) {
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

func (r *Router) Graft(pattern string, router *Router) {
    r.tree.addChild(pattern, router.tree)
}


func (r *Router) Route(url string, ctx Context) *node {
    paths := strings.Split(url, "/")[1:]
    n, components := r.tree.traverse(paths, ctx)

    if len(components) == 0 {
        return n
    }

    return nil
}


// set default handler
func (r *Router) DefaultHandler(handler Handler) {
    r.handler = handler
}


func NewRouter() *Router {
    node := node{
        component: "/",
        isParam: false,
        methods: make(map[string]Handler),
    }

    router := &Router{
        tree: &node,
        handler: defaultHandler,
    }

    return router
}
