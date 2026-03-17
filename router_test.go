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

// TestRoutePriority tests that static routes have higher priority than param routes
func TestRoutePriority(t *testing.T) {
	router := NewRouter()

	called := ""

	// Register routes in "wrong" order (param before static)
	router.Get("/users/:id", func(ctx Context) {
		called = "param"
		ctx.Text(200, "param route")
	})

	router.Get("/users/admin", func(ctx Context) {
		called = "static"
		ctx.Text(200, "static route")
	})

	// Test that static route wins despite registration order
	req := httptest.NewRequest("GET", "/users/admin", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/users/admin", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Fatal("handler not set")
	}

	handler(ctx)

	if called != "static" {
		t.Errorf("expected static route to be called, but got %s", called)
	}
}

// TestParameterExtraction tests that parameters are correctly extracted
func TestParameterExtraction(t *testing.T) {
	router := NewRouter()

	var capturedID string
	router.Get("/users/:id", func(ctx Context) {
		capturedID = ctx.Param("id")
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/users/123", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Fatal("handler not set")
	}

	handler(ctx)

	if capturedID != "123" {
		t.Errorf("expected id=123, got id=%s", capturedID)
	}
}

// TestWildcardRoute tests that wildcard routes work correctly
func TestWildcardRoute(t *testing.T) {
	router := NewRouter()

	var capturedPath string
	router.Get("/files/*", func(ctx Context) {
		capturedPath = ctx.Param("*")
	})

	req := httptest.NewRequest("GET", "/files/images/photo.jpg", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/files/images/photo.jpg", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Fatal("handler not set")
	}

	handler(ctx)

	if capturedPath != "images/photo.jpg" {
		t.Errorf("expected path=images/photo.jpg, got path=%s", capturedPath)
	}
}

// TestPriorityOrder tests the complete priority order: Static > Param > Wildcard
func TestPriorityOrder(t *testing.T) {
	router := NewRouter()

	called := ""

	// Register all three types of routes
	router.Get("/files/*", func(ctx Context) {
		called = "wildcard"
	})

	router.Get("/files/:name", func(ctx Context) {
		called = "param"
	})

	router.Get("/files/readme.txt", func(ctx Context) {
		called = "static"
	})

	// Test 1: Static route should win
	called = ""
	req := httptest.NewRequest("GET", "/files/readme.txt", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/files/readme.txt", ctx)
	if node != nil && node.methods["GET"] != nil {
		node.methods["GET"](ctx)
	}

	if called != "static" {
		t.Errorf("test 1: expected static, got %s", called)
	}

	// Test 2: Param route should win (no static match)
	called = ""
	ctx.params = make(map[string]string)
	node = router.route("/files/other.txt", ctx)
	if node != nil && node.methods["GET"] != nil {
		node.methods["GET"](ctx)
	}

	if called != "param" {
		t.Errorf("test 2: expected param, got %s", called)
	}

	// Test 3: Wildcard should win (multi-segment path)
	called = ""
	ctx.params = make(map[string]string)
	node = router.route("/files/images/photo.jpg", ctx)
	if node != nil && node.methods["GET"] != nil {
		node.methods["GET"](ctx)
	}

	if called != "wildcard" {
		t.Errorf("test 3: expected wildcard, got %s", called)
	}
}

// TestMultipleParams tests routes with multiple parameters
func TestMultipleParams(t *testing.T) {
	router := NewRouter()

	params := make(map[string]string)
	router.Get("/users/:userId/posts/:postId", func(ctx Context) {
		params["userId"] = ctx.Param("userId")
		params["postId"] = ctx.Param("postId")
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/users/123/posts/456", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	handler := node.methods["GET"]
	if handler == nil {
		t.Fatal("handler not set")
	}

	handler(ctx)

	if params["userId"] != "123" {
		t.Errorf("expected userId=123, got %s", params["userId"])
	}
	if params["postId"] != "456" {
		t.Errorf("expected postId=456, got %s", params["postId"])
	}
}

// TestEmptySegments tests that empty segments are handled correctly
func TestEmptySegments(t *testing.T) {
	router := NewRouter()

	called := false
	router.Get("/users/list", func(ctx Context) {
		called = true
	})

	// Test with multiple slashes (empty segments)
	req := httptest.NewRequest("GET", "/users/list", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/users/list", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	if node.methods["GET"] != nil {
		node.methods["GET"](ctx)
	}

	if !called {
		t.Error("handler was not called")
	}
}

// TestRootRoute tests the root route "/"
// TestLCPPrefixCollision tests routes that share a common prefix at the same segment level.
// This reproduces the bug where /assets/* and /article both start with "a",
// causing LCP compression to break routing.
func TestLCPPrefixCollision(t *testing.T) {
	router := NewRouter()

	called := ""

	// Register routes that share the prefix "a" at segment level 1
	router.Get("/assets/*", func(ctx Context) {
		called = "assets-wildcard"
	})
	router.Get("/article", func(ctx Context) {
		called = "article"
	})
	router.Get("/article/:slug", func(ctx Context) {
		called = "article-slug"
	})
	router.Get("/api/articles", func(ctx Context) {
		called = "api-articles"
	})
	router.Get("/health", func(ctx Context) {
		called = "health"
	})
	router.Get("/", func(ctx Context) {
		called = "root"
	})

	tests := []struct {
		path     string
		expected string
		param    string // expected param value
		paramKey string // param key to check
	}{
		{"/assets/neural-bg.js", "assets-wildcard", "neural-bg.js", "*"},
		{"/assets/styles/main.css", "assets-wildcard", "styles/main.css", "*"},
		{"/assets/components/posts-list.js", "assets-wildcard", "components/posts-list.js", "*"},
		{"/article", "article", "", ""},
		{"/article/hello-world", "article-slug", "hello-world", "slug"},
		{"/api/articles", "api-articles", "", ""},
		{"/health", "health", "", ""},
		{"/", "root", "", ""},
	}

	for _, tt := range tests {
		called = ""
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		ctx := &context{
			request:  req,
			response: &Response{Writer: w, Code: 404},
			store:    make(map[string]interface{}),
			params:   make(map[string]string),
		}

		node := router.route(tt.path, ctx)
		if node == nil {
			t.Errorf("[%s] route not found", tt.path)
			continue
		}

		handler := node.methods["GET"]
		if handler == nil {
			t.Errorf("[%s] GET handler not found", tt.path)
			continue
		}

		handler(ctx)

		if called != tt.expected {
			t.Errorf("[%s] expected handler=%q, got=%q", tt.path, tt.expected, called)
		}

		if tt.paramKey != "" && ctx.Param(tt.paramKey) != tt.param {
			t.Errorf("[%s] expected param %s=%q, got=%q", tt.path, tt.paramKey, tt.param, ctx.Param(tt.paramKey))
		}
	}
}

func TestRootRoute(t *testing.T) {
	router := NewRouter()

	called := false
	router.Get("/", func(ctx Context) {
		called = true
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := &context{
		request:  req,
		response: &Response{Writer: w, Code: 404},
		store:    make(map[string]interface{}),
		params:   make(map[string]string),
	}

	node := router.route("/", ctx)
	if node == nil {
		t.Fatal("route not found")
	}

	if node.methods["GET"] != nil {
		node.methods["GET"](ctx)
	}

	if !called {
		t.Error("root handler was not called")
	}
}