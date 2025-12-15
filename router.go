package arry

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
	tree        *radixTree
	handler     Handler
	middlewares []Middleware
}

func (r *Router) Handle(method string, pattern string, handler Handler) {
	// Apply current layer middlewares
	if len(r.middlewares) > 0 {
		handler = applyMiddlewares(handler, r.middlewares)
	}
	r.tree.insert(method, pattern, handler)
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
		r.applyMiddlewaresToSubtree(subRouter.tree.root, r.middlewares)
	}

	r.graftSubtree(pattern, subRouter.tree.root)
}

// applyMiddlewaresToSubtree recursively applies middlewares to all handlers in the radix tree
func (r *Router) applyMiddlewaresToSubtree(n *radixNode, middlewares []Middleware) {
	if n == nil {
		return
	}

	// Apply middlewares to all HTTP methods in current node
	for method, handler := range n.methods {
		n.methods[method] = applyMiddlewares(handler, middlewares)
	}

	// Recursively apply to child nodes
	for _, child := range n.children {
		r.applyMiddlewaresToSubtree(child, middlewares)
	}
	r.applyMiddlewaresToSubtree(n.paramChild, middlewares)
	r.applyMiddlewaresToSubtree(n.catchAllChild, middlewares)
}

// graftSubtree grafts a subtree at the given pattern
func (r *Router) graftSubtree(pattern string, subtree *radixNode) {
	if subtree == nil {
		return
	}

	// For each method in the subtree, register it with the combined pattern
	r.graftNode(pattern, subtree, "")
}

// graftNode recursively grafts nodes from the subtree
func (r *Router) graftNode(prefix string, n *radixNode, currentPath string) {
	if n == nil {
		return
	}

	// Build the full path for this node
	var fullPath string
	if n.nodeType == ntRoot {
		fullPath = prefix
	} else if n.nodeType == ntParam {
		fullPath = currentPath + "/:" + n.paramName
	} else if n.nodeType == ntCatchAll {
		fullPath = currentPath + "/*"
	} else {
		fullPath = currentPath + "/" + n.prefix
	}

	// If this node has methods, register them
	for method, handler := range n.methods {
		if n.nodeType == ntRoot && prefix != "" {
			r.tree.insert(method, prefix, handler)
		} else if fullPath != "" {
			r.tree.insert(method, fullPath, handler)
		}
	}

	// Recursively graft children
	for _, child := range n.children {
		r.graftNode(prefix, child, fullPath)
	}
	if n.paramChild != nil {
		r.graftNode(prefix, n.paramChild, fullPath)
	}
	if n.catchAllChild != nil {
		r.graftNode(prefix, n.catchAllChild, fullPath)
	}
}

func (r *Router) route(url string, ctx Context) *radixNode {
	return r.tree.search(url, ctx)
}

// Handler returns the default handler
func (r *Router) Handler() Handler {
	return r.handler
}

func NewRouter(middlewares ...Middleware) *Router {
	router := &Router{
		tree:        newRadixTree(),
		handler:     defaultHandler,
		middlewares: middlewares,
	}

	return router
}
