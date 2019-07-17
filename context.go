package arry

import (
	// "io"
	// "fmt"
	"os"
	"mime"
	"bytes"
	"path/filepath"
	"net/http"
	"encoding/json"
)


type Context interface {
	Response() *Response
	Request() *http.Request
	// set engine
	SetEngine(engine Engine)
	// get param
	Param(key string) string
	// get cookie
	Cookie(name string) *http.Cookie
	Cookies() []*http.Cookie
	SetCookie(cookie *http.Cookie)
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
	JSONBlob(code int, body []byte)
	File(name string)
	Render(code int, name string, data interface{})
	Push(url string) error
}


type context struct {
	response *Response
	request *http.Request
	params map[string]string
	store map[string]interface{}
	engine Engine
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

func (c *context) Decode(i interface{}) error {
	body := c.Request().Body
	return json.NewDecoder(body).Decode(i)
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
	c.SetContentType("text/html")
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


func getMineType(file string) string {
	ext := filepath.Ext(file)
	t := mime.TypeByExtension(ext)
	if t == "" {
		t = "text/plain"
	}
	return t
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
