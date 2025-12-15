package arry

import (
	"net/http/httptest"
	"testing"
)

// TestPrefixCompression tests that prefix compression is working
func TestPrefixCompression(t *testing.T) {
	router := NewRouter()

	// Register routes that share prefixes
	router.Get("/users/admin", func(ctx Context) {})
	router.Get("/users/list", func(ctx Context) {})
	router.Get("/users/create", func(ctx Context) {})

	// All routes should be accessible
	tests := []string{
		"/users/admin",
		"/users/list",
		"/users/create",
	}

	for _, path := range tests {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		ctx := &context{
			request:  req,
			response: &Response{Writer: w, Code: 404},
			store:    make(map[string]interface{}),
			params:   make(map[string]string),
		}

		node := router.route(path, ctx)
		if node == nil {
			t.Errorf("route %s not found", path)
			continue
		}

		if node.methods["GET"] == nil {
			t.Errorf("handler not set for %s", path)
		}
	}
}

// TestLongestCommonPrefix tests the LCP function
func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 0},
		{"", "a", 0},
		{"abc", "abc", 3},
		{"abc", "abd", 2},
		{"hello", "help", 3},
		{"users/admin", "users/list", 6}, // "users/"
	}

	for _, tt := range tests {
		got := longestCommonPrefix(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("longestCommonPrefix(%q, %q) = %d, want %d",
				tt.a, tt.b, got, tt.want)
		}
	}
}
