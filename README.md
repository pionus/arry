# Arry

> A blazingly fast, simple and elegant web framework for Go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/pionus/arry)](https://goreportcard.com/report/github.com/pionus/arry)

## ⚡ Features

- **Blazingly Fast** - 278x faster template rendering with built-in caching
- **Simple & Elegant** - Clean API design, easy to learn and use
- **Powerful Router** - Pattern-based routing with path parameters and middleware support
- **Complete HTTP Methods** - GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, and more
- **Multi-Engine Support** - HTML, JSON, XML, YAML, and Plain Text templates
- **Rich Middleware** - Built-in Logger, Gzip, Panic recovery, and Auth
- **Context Utilities** - Redirect, Stream, ClientIP, Bind, and more
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
router.Delete("/users/:id", deleteUser)
router.Patch("/users/:id", patchUser)
router.Options("/users", optionsHandler)
router.Head("/users", headHandler)

// Match all HTTP methods
router.Any("/health", healthCheck)

// Match specific methods
router.Match([]string{"GET", "POST"}, "/webhook", webhookHandler)
```

#### Router with Middleware

```go
// Create router with middleware
authMiddleware := func(next arry.Handler) arry.Handler {
    return func(ctx arry.Context) {
        // Authentication logic
        if !isAuthenticated(ctx) {
            ctx.JSON(401, map[string]string{"error": "Unauthorized"})
            return
        }
        next(ctx)
    }
}

// Apply middleware to router
adminRouter := arry.NewRouter(authMiddleware)
adminRouter.Get("/dashboard", dashboardHandler)
adminRouter.Get("/settings", settingsHandler)

// Mount to main router
app.Router().Graft("/admin", adminRouter)
// All routes under /admin/* will require authentication
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
ctx.Redirect(302, "/new-url")

// Stream response
file, _ := os.Open("video.mp4")
ctx.Stream(200, "video/mp4", file)

// File download
data := bytes.NewReader([]byte("file content"))
ctx.Attachment("document.pdf", data)
```

#### Utilities

```go
// Get client IP (handles X-Forwarded-For, X-Real-IP)
clientIP := ctx.ClientIP()

// Bind request data (auto-detect JSON/Form)
var user User
if err := ctx.Bind(&user); err != nil {
    ctx.JSON(400, map[string]string{"error": err.Error()})
    return
}
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

Arry includes a high-performance template engine with caching and multiple format support.

#### Basic Usage

```go
// Set template directory (HTML by default)
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

#### Multiple Engine Types

Arry supports multiple template engines for different output formats:

```go
// HTML Engine (default)
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Type:      arry.EngineHTML,
    Dir:       "templates",
    Extension: "html",
    Cache:     true,
})

// JSON Engine - serialize data to JSON
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Type:   arry.EngineJSON,
    Indent: "  ",  // Pretty print
})

// XML Engine - serialize data to XML
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Type:   arry.EngineXML,
    Indent: "  ",
})

// Plain Text Engine - for YAML, config files, etc.
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Type:      arry.EnginePlain,
    Dir:       "configs",
    Extension: "yaml",
    Cache:     true,
})

// YAML Engine
app.Engine = arry.NewEngineWithConfig(arry.EngineConfig{
    Type: arry.EngineYAML,
    Dir:  "configs",
})
```

#### Auto-Detection

The engine type can be automatically detected from file extension:

```go
// Auto-detect based on extension
app.Engine = arry.NewEngine("templates", "html")  // HTML engine
app.Engine = arry.NewEngine("templates", "json")  // JSON engine
app.Engine = arry.NewEngine("templates", "xml")   // XML engine
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
    Bind(i interface{}) error

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
    Redirect(code int, url string)
    Stream(code int, contentType string, reader io.Reader) error
    Attachment(filename string, reader io.Reader) error

    // Cookies
    Cookie(name string) *http.Cookie
    Cookies() []*http.Cookie
    SetCookie(cookie *http.Cookie)

    // Store (context values)
    Set(key string, value interface{})
    Get(key string) interface{}

    // Utilities
    ClientIP() string

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
router.Delete(pattern string, handler Handler)
router.Patch(pattern string, handler Handler)
router.Options(pattern string, handler Handler)
router.Head(pattern string, handler Handler)

// Multiple methods
router.Any(pattern string, handler Handler)
router.Match(methods []string, pattern string, handler Handler)

// Custom method
router.Handle(method, pattern string, handler Handler)

// Mount sub-router with middleware inheritance
router.Graft(pattern string, subrouter *Router)

// Create router with middleware
NewRouter(middlewares ...Middleware) *Router
```

### Engine Configuration

```go
type EngineType string

const (
    EngineHTML  EngineType = "html"
    EngineJSON  EngineType = "json"
    EngineXML   EngineType = "xml"
    EnginePlain EngineType = "plain"
    EngineYAML  EngineType = "yaml"
)

type EngineConfig struct {
    Type      EngineType          // Engine type (HTML, JSON, XML, Plain, YAML)
    Dir       string              // Template directory
    Extension string              // File extension (auto-detected if not specified)
    FuncMap   template.FuncMap    // Custom template functions
    Cache     bool                // Enable caching (production: true, dev: false)
    Indent    string              // Indentation for JSON/XML (e.g., "  " or "\t")
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
    "time"
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

### Multi-Engine Example

```go
package main

import (
    "github.com/pionus/arry"
)

type Config struct {
    Server   string `json:"server" xml:"server" yaml:"server"`
    Port     int    `json:"port" xml:"port" yaml:"port"`
    Features []string `json:"features" xml:"features" yaml:"features"`
}

func main() {
    app := arry.New()
    router := app.Router()

    config := Config{
        Server:   "example.com",
        Port:     8080,
        Features: []string{"fast", "simple", "elegant"},
    }

    // JSON endpoint
    router.Get("/config.json", func(ctx arry.Context) {
        ctx.SetEngine(arry.NewEngineWithConfig(arry.EngineConfig{
            Type: arry.EngineJSON,
            Indent: "  ",
        }))
        ctx.Render(200, "config", config)
    })

    // XML endpoint
    router.Get("/config.xml", func(ctx arry.Context) {
        ctx.SetEngine(arry.NewEngineWithConfig(arry.EngineConfig{
            Type: arry.EngineXML,
            Indent: "  ",
        }))
        ctx.Render(200, "config", config)
    })

    // YAML endpoint
    router.Get("/config.yaml", func(ctx arry.Context) {
        ctx.SetEngine(arry.NewEngineWithConfig(arry.EngineConfig{
            Type: arry.EngineYAML,
        }))
        ctx.Render(200, "config", config)
    })

    app.Start(":8080")
}
```

### Router Composition with Middleware

```go
package main

import (
    "github.com/pionus/arry"
    "github.com/pionus/arry/middlewares"
)

func main() {
    app := arry.New()

    // Global middleware
    app.Use(middlewares.Logger())
    app.Use(middlewares.Panic)

    // Public API router (no auth)
    publicAPI := arry.NewRouter()
    publicAPI.Get("/health", healthHandler)
    publicAPI.Get("/version", versionHandler)

    // Admin API router (with auth middleware)
    authMiddleware := middlewares.Auth("admin", "secret")
    adminAPI := arry.NewRouter(authMiddleware)
    adminAPI.Get("/dashboard", dashboardHandler)
    adminAPI.Get("/users", usersHandler)
    adminAPI.Delete("/users/:id", deleteUserHandler)

    // API v1 router
    apiV1 := arry.NewRouter()
    apiV1.Get("/posts", getPostsHandler)
    apiV1.Post("/posts", createPostHandler)

    // Mount all routers
    app.Router().Graft("/api/public", publicAPI)
    app.Router().Graft("/api/admin", adminAPI)
    app.Router().Graft("/api/v1", apiV1)

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
