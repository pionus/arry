package arry

import (
	"testing"
	"net/http/httptest"
)


func TestRouter(t *testing.T) {
	router := NewRouter()
	
	node := router.route("/", &context{})

	if node == nil {
		t.Error("default router is not correct")
	}
}

func TestRouterPath(t *testing.T) {
	router := NewRouter()
	url := "/path/to/route"

	router.Get(url, defaultHandler)
	node := router.route(url, nil)

	if node == nil {
		t.Error("router path is not correct")
	}
}

func TestGraft(t *testing.T) {
	router := NewRouter()
	sub := NewRouter()

	sub.Get("/path/sub", defaultHandler)
	router.Graft("/s", sub)
	node := router.route("/s/path/sub", nil)

	if node == nil {
		t.Error("router graft is not correct")
	}
}

// Test Router with middlewares
func TestRouterWithMiddlewares(t *testing.T) {
	middlewareCalled := false
	testMiddleware := func(next Handler) Handler {
		return func(ctx Context) {
			middlewareCalled = true
			ctx.SetHeader("X-Test", "middleware-applied")
			next(ctx)
		}
	}

	router := NewRouter(testMiddleware)
	router.Get("/test", func(ctx Context) {
		ctx.Text(200, "ok")
	})

	// Simulate request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
	}

	node := router.route("/test", ctx)
	if node == nil {
		t.Error("route not found")
		return
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Error("handler not set")
		return
	}

	handler(ctx)

	if !middlewareCalled {
		t.Error("middleware was not called")
	}

	if w.Header().Get("X-Test") != "middleware-applied" {
		t.Error("middleware did not set header")
	}
}

// Test Graft inherits parent middlewares
func TestGraftInheritsMiddlewares(t *testing.T) {
	order := []string{}

	parentMiddleware := func(next Handler) Handler {
		return func(ctx Context) {
			order = append(order, "parent-before")
			next(ctx)
			order = append(order, "parent-after")
		}
	}

	childMiddleware := func(next Handler) Handler {
		return func(ctx Context) {
			order = append(order, "child-before")
			next(ctx)
			order = append(order, "child-after")
		}
	}

	// Create child router with its own middleware
	child := NewRouter(childMiddleware)
	child.Get("/test", func(ctx Context) {
		order = append(order, "handler")
	})

	// Create parent router and graft child
	parent := NewRouter(parentMiddleware)
	parent.Graft("/child", child)

	// Simulate request
	req := httptest.NewRequest("GET", "/child/test", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
	}

	node := parent.route("/child/test", ctx)
	if node == nil {
		t.Error("route not found")
		return
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Error("handler not set")
		return
	}

	handler(ctx)

	// Verify execution order: parent -> child -> handler -> child -> parent
	expected := []string{"parent-before", "child-before", "handler", "child-after", "parent-after"}
	if len(order) != len(expected) {
		t.Errorf("execution order length mismatch: got %d, want %d", len(order), len(expected))
		t.Logf("Got: %v", order)
		return
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("execution order[%d]: got %s, want %s", i, order[i], v)
		}
	}
}

// Test nested Graft with multiple layers
func TestNestedGraftMiddlewares(t *testing.T) {
	order := []string{}

	m1 := func(next Handler) Handler {
		return func(ctx Context) {
			order = append(order, "m1")
			next(ctx)
		}
	}

	m2 := func(next Handler) Handler {
		return func(ctx Context) {
			order = append(order, "m2")
			next(ctx)
		}
	}

	m3 := func(next Handler) Handler {
		return func(ctx Context) {
			order = append(order, "m3")
			next(ctx)
		}
	}

	// Create three-level nesting
	users := NewRouter(m3)
	users.Get("/list", func(ctx Context) {
		order = append(order, "handler")
	})

	api := NewRouter(m2)
	api.Graft("/users", users)

	main := NewRouter(m1)
	main.Graft("/api", api)

	// Simulate request to /api/users/list
	req := httptest.NewRequest("GET", "/api/users/list", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
	}

	node := main.route("/api/users/list", ctx)
	if node == nil {
		t.Error("route not found")
		return
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Error("handler not set")
		return
	}

	handler(ctx)

	// Verify execution order: m1 -> m2 -> m3 -> handler
	expected := []string{"m1", "m2", "m3", "handler"}
	if len(order) != len(expected) {
		t.Errorf("execution order length mismatch: got %d, want %d", len(order), len(expected))
		t.Logf("Got: %v", order)
		return
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("execution order[%d]: got %s, want %s", i, order[i], v)
		}
	}
}

// Test Graft without middlewares
func TestGraftWithoutMiddlewares(t *testing.T) {
	middlewareCalled := false
	childMiddleware := func(next Handler) Handler {
		return func(ctx Context) {
			middlewareCalled = true
			next(ctx)
		}
	}

	// Parent has no middleware
	parent := NewRouter()

	// Child has middleware
	child := NewRouter(childMiddleware)
	child.Get("/test", func(ctx Context) {})

	parent.Graft("/child", child)

	// Simulate request
	req := httptest.NewRequest("GET", "/child/test", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
	}

	node := parent.route("/child/test", ctx)
	if node == nil {
		t.Error("route not found")
		return
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Error("handler not set")
		return
	}

	handler(ctx)

	// Only child middleware should be called
	if !middlewareCalled {
		t.Error("child middleware was not called")
	}
}

// Test Router.Any method
func TestRouterAny(t *testing.T) {
	router := NewRouter()
	router.Any("/health", defaultHandler)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		node := router.route("/health", nil)
		if node == nil {
			t.Errorf("Any: route /health not found for method %s", method)
			continue
		}

		if node.methods[method] == nil {
			t.Errorf("Any: handler not set for method %s", method)
		}
	}
}

// Test Router.Match method
func TestRouterMatch(t *testing.T) {
	router := NewRouter()
	router.Match([]string{"GET", "POST"}, "/webhook", defaultHandler)

	// Test that matched methods work
	node := router.route("/webhook", nil)
	if node == nil {
		t.Error("Match: route /webhook not found")
		return
	}

	if node.methods["GET"] == nil {
		t.Error("Match: GET handler not set")
	}

	if node.methods["POST"] == nil {
		t.Error("Match: POST handler not set")
	}

	// Test that non-matched methods don't work
	if node.methods["PUT"] != nil {
		t.Error("Match: PUT handler should not be set")
	}
}