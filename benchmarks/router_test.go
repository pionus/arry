package benchmarks

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/pionus/arry"
)

// BenchmarkStaticRouting benchmarks static route lookups
func BenchmarkStaticRouting(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register static routes
	router.Get("/users/list", func(ctx arry.Context) {})
	router.Get("/users/create", func(ctx arry.Context) {})
	router.Get("/users/delete", func(ctx arry.Context) {})
	router.Get("/posts/list", func(ctx arry.Context) {})
	router.Get("/posts/create", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/users/list", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkParamRouting benchmarks parameter route lookups
func BenchmarkParamRouting(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register parameter routes
	router.Get("/users/:id", func(ctx arry.Context) {})
	router.Get("/posts/:id", func(ctx arry.Context) {})
	router.Get("/comments/:id", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkDeepNestedRoutes benchmarks deeply nested route lookups
func BenchmarkDeepNestedRoutes(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register deeply nested routes
	router.Get("/api/v1/users/:userId/posts/:postId/comments/:commentId", func(ctx arry.Context) {})
	router.Get("/api/v1/users/:userId/posts/:postId", func(ctx arry.Context) {})
	router.Get("/api/v1/users/:userId", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/users/123/posts/456/comments/789", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkMixedPriority benchmarks routing with mixed route types
func BenchmarkMixedPriority(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register mixed route types
	router.Get("/users/admin", func(ctx arry.Context) {})
	router.Get("/users/:id", func(ctx arry.Context) {})
	router.Get("/users/*", func(ctx arry.Context) {})
	router.Get("/posts/featured", func(ctx arry.Context) {})
	router.Get("/posts/:id", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/users/admin", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkLargeRouteTable100 benchmarks performance with 100 routes
func BenchmarkLargeRouteTable100(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register 100 routes
	for i := 0; i < 100; i++ {
		pattern := fmt.Sprintf("/route%d/:id", i)
		router.Get(pattern, func(ctx arry.Context) {})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/route50/123", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkLargeRouteTable1000 benchmarks performance with 1000 routes
func BenchmarkLargeRouteTable1000(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register 1000 routes
	for i := 0; i < 1000; i++ {
		pattern := fmt.Sprintf("/route%d/:id", i)
		router.Get(pattern, func(ctx arry.Context) {})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/route500/123", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkPrefixCompression benchmarks routes with shared prefixes
func BenchmarkPrefixCompression(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register routes with shared prefixes
	router.Get("/api/v1/users/list", func(ctx arry.Context) {})
	router.Get("/api/v1/users/create", func(ctx arry.Context) {})
	router.Get("/api/v1/users/update", func(ctx arry.Context) {})
	router.Get("/api/v1/users/delete", func(ctx arry.Context) {})
	router.Get("/api/v1/posts/list", func(ctx arry.Context) {})
	router.Get("/api/v1/posts/create", func(ctx arry.Context) {})
	router.Get("/api/v2/users/list", func(ctx arry.Context) {})
	router.Get("/api/v2/posts/list", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkWildcardRouting benchmarks wildcard route lookups
func BenchmarkWildcardRouting(b *testing.B) {
	app := arry.New()
	router := app.Router()

	// Register wildcard route
	router.Get("/files/*", func(ctx arry.Context) {})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/files/images/photos/vacation/2024/summer.jpg", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
}

// BenchmarkRouteInsertion benchmarks route insertion performance
func BenchmarkRouteInsertion(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router := arry.NewRouter()

		// Insert 100 routes
		for j := 0; j < 100; j++ {
			pattern := fmt.Sprintf("/route%d/:id", j)
			router.Get(pattern, func(ctx arry.Context) {})
		}
	}
}

// BenchmarkGraftOperation benchmarks the Graft operation
func BenchmarkGraftOperation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mainRouter := arry.NewRouter()
		subRouter := arry.NewRouter()

		// Add routes to sub-router
		for j := 0; j < 10; j++ {
			pattern := fmt.Sprintf("/resource%d", j)
			subRouter.Get(pattern, func(ctx arry.Context) {})
		}

		// Graft sub-router
		mainRouter.Graft("/api/v1", subRouter)
	}
}
