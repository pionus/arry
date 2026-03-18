package arry

import (
	// "fmt"
	"os"
	"io"
	"mime"
	"bytes"
	"strconv"
	"strings"
	"path/filepath"
	"net/http"
	"net/url"
	"encoding/json"
)


type Context interface {
	Response() *Response
	Request() *http.Request
	// set engine
	SetEngine(engine Engine)
	// get param
	Param(key string) string
	// query parameters
	Query(key string) string
	QueryDefault(key string, defaultValue string) string
	QueryInt(key string, defaultValue int) int
	QueryParams() url.Values
	// headers
	Header(key string) string
	SetHeader(key string, value string)
	// get cookie
	Cookie(name string) *http.Cookie
	Cookies() []*http.Cookie
	SetCookie(cookie *http.Cookie)
	Body() []byte
	Decode(i interface{}) error
	// set status
	Status(code int)
	Set(key string, value interface{})
	Get(key string) interface{}
	SetContentType(value string)
	// reply response immidiatly
	Reply(code int)
	Text(code int, body string)
	JSON(code int, body interface{})
	JSONError(code int, msg string)
	JSONBlob(code int, body []byte)
	File(name string)
	Render(code int, name string, data interface{})
	Push(url string) error
	// Utility methods
	Redirect(code int, url string)
	Stream(code int, contentType string, reader io.Reader) error
	Attachment(filename string, reader io.Reader) error
	ClientIP() string
	Bind(i interface{}) error
	// Auth helpers
	IsAuthed() bool
}


type context struct {
	response *Response
	request *http.Request
	params map[string]string
	store map[string]interface{}
	engine Engine
	body []byte  // Cached request body
	bodyRead bool // Whether body has been read
}


func NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request: r,
		response: &Response{Writer: w, Code: http.StatusNotFound},
	}
}


type JSONTemplate struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func (c *context) Response() *Response {
	return c.response
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetEngine(engine Engine) {
	c.engine = engine
}

func (c *context) Param(key string) string {
	return c.params[key]
}

// Query returns the query parameter value for the given key
func (c *context) Query(key string) string {
	return c.Request().URL.Query().Get(key)
}

// QueryDefault returns the query parameter value or default if not found
func (c *context) QueryDefault(key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// QueryInt returns the query parameter as an integer, or defaultValue if missing/invalid
func (c *context) QueryInt(key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultValue
}

// QueryParams returns all query parameters
func (c *context) QueryParams() url.Values {
	return c.Request().URL.Query()
}

// Header returns the request header value for the given key
func (c *context) Header(key string) string {
	return c.Request().Header.Get(key)
}

// SetHeader sets a response header
func (c *context) SetHeader(key string, value string) {
	c.Response().Header().Set(key, value)
}

func (c *context) Cookie(name string) *http.Cookie {
	cookie, err := c.Request().Cookie(name)
	if err != nil {
		return nil
	}

	return cookie
}

func (c *context) Cookies() []*http.Cookie {
	return c.Request().Cookies()
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Body() []byte {
	if !c.bodyRead {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			// Log error but don't fail - return empty body
			c.body = []byte{}
		} else {
			c.body = body
		}
		c.bodyRead = true
	}
	return c.body
}

func (c *context) Decode(i interface{}) error {
	// Use cached body to allow multiple reads
	body := c.Body()
	return json.NewDecoder(bytes.NewReader(body)).Decode(i)
}

func (c *context) Status(code int) {
	c.Response().Code = code
}

func (c *context) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(map[string]interface{})
	}

	c.store[key] = value
}

func (c *context) Get(key string) interface{} {
	return c.store[key]
}

func (c *context) SetContentType(value string) {
	header := c.Response().Header()

	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", value)
	}
}

func (c *context) Reply(code int) {
	message := http.StatusText(code)
	acceptType := c.Request().Header.Get("Accept")

	switch acceptType {
	case "application/json":
		body := JSONTemplate{ code, message }
		c.JSON(code, body)
		return
	default:
		c.Text(code, message)
		return
	}
}


func (c *context) Text(code int, body string) {
	c.SetContentType("text/plain")
	c.Blob(code, []byte(body))
}

func (c *context) JSON(code int, body interface{}) {
	encoder := json.NewEncoder(c.Response())
	c.SetContentType("application/json")
	c.Status(code)
	encoder.Encode(body)
}

// JSONError sends a JSON error response: {"error": msg}
func (c *context) JSONError(code int, msg string) {
	c.JSON(code, map[string]string{"error": msg})
}

func (c *context) JSONBlob(code int, body []byte) {
	c.SetContentType("application/json")
	c.Blob(code, body)
}

func (c *context) Blob(code int, body []byte) {
	c.Status(code)
	c.Response().Write(body)
}

func (c *context) File(name string) {
	f, err := os.Open(name)
	if err != nil {
		c.Reply(404)
		return
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		c.Reply(404)
		return
	}

	mimeType := getMineType(name)
	c.SetContentType(mimeType)

	c.Status(http.StatusOK)
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
}

func (c *context) Render(code int, name string, data interface{}) {
	c.Status(code)

	buf := new(bytes.Buffer)
	c.engine.Render(buf, name, data, c)

	// Use engine's content type if available
	if c.engine != nil {
		c.SetContentType(c.engine.ContentType())
	} else {
		c.SetContentType("text/html; charset=utf-8")
	}

	c.Blob(code, buf.Bytes())
}

func (c *context) Push(url string) error {
	pusher, ok := c.Response().Writer.(http.Pusher)

	if ok {
		return pusher.Push(url, nil)
	} else {
		asType := getPushType(url)
		c.Response().Header().Add("Link", "<"+ url +">; rel=preload; as="+ asType)
		return nil
	}
}


var mimeTypes = map[string]string{
	".js":    "application/javascript",
	".mjs":   "application/javascript",
	".css":   "text/css; charset=utf-8",
	".html":  "text/html; charset=utf-8",
	".json":  "application/json",
	".xml":   "application/xml",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".webp":  "image/webp",
	".ico":   "image/x-icon",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
	".pdf":   "application/pdf",
	".wasm":  "application/wasm",
	".map":   "application/json",
}

func getMineType(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	if t, ok := mimeTypes[ext]; ok {
		return t
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}

func getPushType(file string) string {
	switch ext := filepath.Ext(file); ext {
	case ".js":
		return "script"
	case ".css":
		return "style"
	default:
		return "image"
	}
}

// Redirect sends an HTTP redirect response
func (c *context) Redirect(code int, url string) {
	if code < 300 || code > 308 {
		code = http.StatusFound // Default to 302
	}
	c.SetHeader("Location", url)
	c.Status(code)
	c.Response().WriteHeader(code)
}

// Stream sends a streaming response
func (c *context) Stream(code int, contentType string, reader io.Reader) error {
	c.SetContentType(contentType)
	c.Status(code)
	_, err := io.Copy(c.Response(), reader)
	return err
}

// Attachment sends a file download response
func (c *context) Attachment(filename string, reader io.Reader) error {
	c.SetHeader("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Stream(http.StatusOK, "application/octet-stream", reader)
}

// ClientIP returns the real client IP address
// It checks X-Forwarded-For, X-Real-IP headers first, then falls back to RemoteAddr
func (c *context) ClientIP() string {
	// Check X-Forwarded-For header
	if xff := c.Header("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := c.Header("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if ip, _, ok := strings.Cut(c.Request().RemoteAddr, ":"); ok {
		return ip
	}
	return c.Request().RemoteAddr
}

// Bind intelligently binds request data to the given interface
// It automatically detects the content type and decodes accordingly
func (c *context) Bind(i interface{}) error {
	contentType := c.Header("Content-Type")

	// Check for JSON content type
	if strings.Contains(contentType, "application/json") {
		return c.Decode(i)
	}

	// Check for form data
	if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		if err := c.Request().ParseForm(); err != nil {
			return err
		}
		// For form data, we would need a form decoder library
		// For now, just return nil as basic form parsing is done
		return nil
	}

	// Default to JSON decoding
	return c.Decode(i)
}

// IsAuthed checks if the request has been authenticated by the Auth middleware
func (c *context) IsAuthed() bool {
	v := c.Get("auth")
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}
