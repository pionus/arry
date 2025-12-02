# Arry

> A blazingly fast, simple and elegant web framework for Go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.14-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/pionus/arry)](https://goreportcard.com/report/github.com/pionus/arry)

## ⚡ Features

- **Blazingly Fast** - 278x faster template rendering with built-in caching
- **Simple & Elegant** - Clean API design, easy to learn and use
- **Powerful Router** - Pattern-based routing with path parameters
- **Rich Middleware** - Built-in Logger, Gzip, Panic recovery, and Auth
- **Template Engine** - High-performance HTML rendering with caching
- **HTTP/2 & TLS** - Built-in HTTPS support with auto Let's Encrypt
- **Graceful Shutdown** - Safe server shutdown handling
- **Production Ready** - Tested, optimized, and battle-proven

## 🚀 Performance

Latest optimizations (v1.0.0):

| Metric | Without Cache | With Cache | Improvement |
|--------|---------------|------------|-------------|
| Speed | 58,579 ns/op | 210.3 ns/op | **278x faster** |
| Memory | 4,816 B/op | 128 B/op | **37x less** |
| Allocations | 38 allocs/op | 2 allocs/op | **19x less** |

## 📦 Installation

```bash
go get github.com/pionus/arry
```

## 🎯 Quick Start

```go
package main

import (
    "net/http"
    "github.com/pionus/arry"
)

func main() {
    app := arry.New()

    app.Router().Get("/", func(ctx arry.Context) {
        ctx.Text(http.StatusOK, "Hello, World!")
    })

    app.Start(":8080")
}
```

## 📚 Table of Contents

- [Core Concepts](#core-concepts)
  - [Routing](#routing)
  - [Context](#context)
  - [Middleware](#middleware)
  - [Template Engine](#template-engine)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

## 🏗️ Core Concepts

### Routing

#### Basic Routes

```go
app := arry.New()
router := app.Router()

// HTTP methods
router.Get("/users", getUsers)
router.Post("/users", createUser)
router.Put("/users/:id", updateUser)
```

#### Path Parameters

```go
router.Get("/hello/:name", func(ctx arry.Context) {
    name := ctx.Param("name")
    ctx.Text(200, "Hello, " + name)
})

// Wildcard
router.Get("/files/*", func(ctx arry.Context) {
    path := ctx.Param("*")
    ctx.File(path)
})
```

#### Static Files

```go
app.Static("/static", "public")
// Serves files from ./public at /static/*
```

### Context

The Context is the heart of request handling in Arry.

#### Query Parameters

```go
router.Get("/search", func(ctx arry.Context) {
    // Get query parameter
    q := ctx.Query("q")

    // With default value
    page := ctx.QueryDefault("page", "1")

    // Get all parameters
    params := ctx.QueryParams()
})
```

#### Request Headers

```go
router.Get("/", func(ctx arry.Context) {
    userAgent := ctx.Header("User-Agent")
    token := ctx.Header("Authorization")
})
```

#### Request Body

```go
router.Post("/users", func(ctx arry.Context) {
    // Read raw body (can be read multiple times!)
    body := ctx.Body()

    // Or decode JSON directly
    var user User
    if err := ctx.Decode(&user); err != nil {
        ctx.JSON(400, map[string]string{"error": err.Error()})
        return
    }

    ctx.JSON(200, user)
})
```

#### Response Methods

```go
// Text response
ctx.Text(200, "Hello")

// JSON response
ctx.JSON(200, map[string]interface{}{
    "name": "John",
    "age": 30,
})

// Render template
ctx.Render(200, "index.html", data)

// Custom headers
ctx.SetHeader("X-Custom-Header", "value")

// Redirect
ctx.SetHeader("Location", "/new-url")
ctx.Status(302)
```

### Middleware

Middleware wraps handlers to add cross-cutting functionality.

#### Using Built-in Middleware

```go
import "github.com/pionus/arry/middlewares"

app := arry.New()

// Logger - logs all requests
app.Use(middlewares.Logger())

// Gzip - compresses responses
app.Use(middlewares.Gzip)

// Panic recovery - catches panics
app.Use(middlewares.Panic)

// Basic Auth
app.Use(middlewares.Auth("username", "password"))
```

#### Custom Middleware

```go
func CORS() arry.Middleware {
    return func(next arry.Handler) arry.Handler {
        return func(ctx arry.Context) {
            ctx.SetHeader("Access-Control-Allow-Origin", "*")
            ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
            next(ctx)
        }
    }
}

app.Use(CORS())
```

#### Route-specific Middleware

```go
// Apply middleware to specific routes
authMiddleware := middlewares.Auth("admin", "secret")

router.Get("/admin", authMiddleware(adminHandler))
```

### Template Engine

Arry includes a high-performance template engine with caching.

#### Basic Usage

```go
// Set template directory
app.Views("templates")

// In handler
router.Get("/", func(ctx arry.Context) {
    data := map[string]interface{}{
        "Title": "Home",
        "User": "John",
    }
    ctx.Render(200, "index.html", data)
})
```

#### Advanced Configuration

```go
import "text/template"

// Production mode with caching (278x faster!)
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Dir:       "templates",
    Extension: "html",
    Cache:     true,  // Enable caching
    FuncMap: template.FuncMap{
        "upper": strings.ToUpper,
        "lower": strings.ToLower,
    },
})
```

#### Development Mode

```go
// Disable caching for development (hot reload)
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Dir:       "templates",
    Extension: "html",
    Cache:     false,  // Reload on every request
})
```

## 📖 API Reference

### Context Interface

```go
type Context interface {
    // Request
    Request() *http.Request
    Param(key string) string
    Query(key string) string
    QueryDefault(key, defaultValue string) string
    QueryParams() url.Values
    Header(key string) string
    Body() []byte
    Decode(i interface{}) error

    // Response
    Response() *Response
    Status(code int)
    SetHeader(key, value string)
    SetContentType(value string)
    Text(code int, body string)
    JSON(code int, body interface{})
    JSONBlob(code int, body []byte)
    Render(code int, name string, data interface{})
    File(name string)

    // Cookies
    Cookie(name string) *http.Cookie
    Cookies() []*http.Cookie
    SetCookie(cookie *http.Cookie)

    // Store (context values)
    Set(key string, value interface{})
    Get(key string) interface{}

    // HTTP/2 Server Push
    Push(url string) error
}
```

### Router Methods

```go
// HTTP methods
router.Get(pattern string, handler Handler)
router.Post(pattern string, handler Handler)
router.Put(pattern string, handler Handler)

// Custom method
router.Handle(method, pattern string, handler Handler)

// Mount sub-router
router.Graft(pattern string, subrouter *Router)

// Default handler (404)
router.DefaultHandler(handler Handler)
```

### Engine Configuration

```go
type EngineConfig struct {
    Dir       string              // Template directory
    Extension string              // File extension (default: "html")
    FuncMap   template.FuncMap    // Custom template functions
    Cache     bool                // Enable caching (production: true, dev: false)
}
```

## 💡 Examples

### RESTful API

```go
package main

import (
    "net/http"
    "github.com/pionus/arry"
    "github.com/pionus/arry/middlewares"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

var users = []User{
    {ID: 1, Name: "Alice", Age: 25},
    {ID: 2, Name: "Bob", Age: 30},
}

func main() {
    app := arry.New()
    app.Use(middlewares.Logger())
    app.Use(middlewares.Gzip)

    router := app.Router()

    // List users
    router.Get("/api/users", func(ctx arry.Context) {
        ctx.JSON(200, users)
    })

    // Get user by ID
    router.Get("/api/users/:id", func(ctx arry.Context) {
        id := ctx.Param("id")
        for _, user := range users {
            if fmt.Sprint(user.ID) == id {
                ctx.JSON(200, user)
                return
            }
        }
        ctx.JSON(404, map[string]string{"error": "User not found"})
    })

    // Create user
    router.Post("/api/users", func(ctx arry.Context) {
        var user User
        if err := ctx.Decode(&user); err != nil {
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        user.ID = len(users) + 1
        users = append(users, user)
        ctx.JSON(201, user)
    })

    app.Start(":8080")
}
```

### Template Rendering

```go
package main

import (
    "github.com/pionus/arry"
    "github.com/pionus/arry/middlewares"
)

func main() {
    app := arry.New()
    app.Use(middlewares.Logger())

    // Configure template engine with caching
    app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
        Dir:   "templates",
        Cache: true,  // 278x performance boost!
    })

    router := app.Router()

    router.Get("/", func(ctx arry.Context) {
        data := map[string]interface{}{
            "Title": "Welcome",
            "User":  "Guest",
        }
        ctx.Render(200, "index.html", data)
    })

    app.Start(":8080")
}
```

### HTTPS with Auto TLS

```go
package main

import "github.com/pionus/arry"

func main() {
    app := arry.New()

    router := app.Router()
    router.Get("/", func(ctx arry.Context) {
        ctx.Text(200, "Secure Hello!")
    })

    // Automatic Let's Encrypt certificate
    app.StartTLS(":443", "example.com", "www.example.com")
}
```

### Middleware Chain

```go
package main

import (
    "log"
    "github.com/pionus/arry"
    "github.com/pionus/arry/middlewares"
)

// Custom request ID middleware
func RequestID() arry.Middleware {
    return func(next arry.Handler) arry.Handler {
        return func(ctx arry.Context) {
            id := generateID()
            ctx.SetHeader("X-Request-ID", id)
            ctx.Set("request_id", id)
            next(ctx)
        }
    }
}

// Custom timing middleware
func Timer() arry.Middleware {
    return func(next arry.Handler) arry.Handler {
        return func(ctx arry.Context) {
            start := time.Now()
            next(ctx)
            duration := time.Since(start)
            log.Printf("Request took %v", duration)
        }
    }
}

func main() {
    app := arry.New()

    // Middleware stack (executed in order)
    app.Use(middlewares.Logger())
    app.Use(RequestID())
    app.Use(Timer())
    app.Use(middlewares.Gzip)
    app.Use(middlewares.Panic)

    router := app.Router()
    router.Get("/", handler)

    app.Start(":8080")
}
```

## 🔧 Advanced Features

### Graceful Shutdown

```go
app := arry.New()
app.Graceful(true)  // Enabled by default

// Handles SIGINT, SIGTERM
app.Start(":8080")
// Press Ctrl+C for graceful shutdown
```

### HTTP/2 Server Push

```go
router.Get("/", func(ctx arry.Context) {
    // Push CSS before sending HTML
    ctx.Push("/static/style.css")
    ctx.Push("/static/app.js")
    ctx.Render(200, "index.html", nil)
})
```

### Custom Engine

Create a custom template engine for other formats (YAML, XML, etc.):

```go
type YAMLEngine struct {
    *arry.HTMLEngine
}

func (e *YAMLEngine) ContentType() string {
    return "text/plain; charset=utf-8"
}

app.Engine = &YAMLEngine{...}
```

## 📊 Benchmarks

Run benchmarks yourself:

```bash
go test -bench=BenchmarkEngine -benchmem
```

Results on Intel i5-8257U @ 1.40GHz:

```
BenchmarkEngineRenderWithCache-8    4826809    210.3 ns/op    128 B/op    2 allocs/op
BenchmarkEngineRenderNoCache-8        41386   58579 ns/op   4816 B/op   38 allocs/op
```

## 🧪 Testing

Run tests:

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Race detection
go test -race ./...
```

Current coverage: **58.3%** (main), **85.7%** (middlewares)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/pionus/arry.git
cd arry

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem
```

### Guidelines

- Write tests for new features
- Maintain backward compatibility
- Follow Go best practices
- Update documentation

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

## 🙏 Acknowledgments

- Inspired by Express.js and other great web frameworks
- Built with ❤️ for the Go community

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Documentation](https://github.com/pionus/arry)
- [Examples](_example/)
- [Issue Tracker](https://github.com/pionus/arry/issues)

---

**Made with ❤️ and Go**

Star ⭐ this repository if you find it helpful!
