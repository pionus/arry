package arry

import (
	// "io"
	// "fmt"
	"os"
	"mime"
	"path/filepath"
	"net/http"
	"encoding/json"
	"html/template"
)


type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	// get param
	Param(key string) string
	// get cookie
	Cookie(name string) *http.Cookie
	SetCookie(cookie *http.Cookie)
	Cookies() []*http.Cookie
	Decode(i interface{}) error
	// set status
	Status(code int)
	GetStatus() int
	Set(key string, value interface{})
	Get(key string) interface{}
	SetContentType(value string)
	// reply response immidiatly
	Reply()
	Text(code int, body string)
	JSON(code int, body interface{})
	File(name string)
	Render(code int, name string, data interface{})
	Push(url string) error
}


type context struct {
	response http.ResponseWriter
	request *http.Request
	code int
	params map[string]string
	store map[string]interface{}
	template *template.Template
}

func (c *context) Response() http.ResponseWriter {
	return c.response
}

func (c *context) Request() *http.Request {
	return c.request
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

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Cookies() []*http.Cookie {
	return c.Request().Cookies()
}

func (c *context) Decode(i interface{}) error {
	body := c.Request().Body
	return json.NewDecoder(body).Decode(i)
}

func (c *context) Status(code int) {
	c.code = code
	c.Response().WriteHeader(c.code)
}

func (c *context) GetStatus() int {
	return c.code
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

func (c *context) Reply() {
	body := http.StatusText(c.code)
	c.Response().Write([]byte(body))
}


func (c *context) Text(code int, body string) {
	c.Status(code)
	c.SetContentType("text/plain")
	c.Response().Write([]byte(body))
}

func (c *context) JSON(code int, body interface{}) {
	encoder := json.NewEncoder(c.Response())
	c.SetContentType("application/json")
	c.Status(code)
	encoder.Encode(body)
}

func (c *context) File(name string) {
	f, err := os.Open(name)
	if err != nil {
		c.Reply()
		return
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		c.Reply()
		return
	}

	mimeType := getMineType(name)
	c.SetContentType(mimeType)

	c.code = http.StatusOK
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
}

func (c *context) Render(code int, name string, data interface{}) {
	c.Status(code)

	c.template.ExecuteTemplate(c.Response(), name, data)
}

func (c *context) Push(url string) error {
	pusher, ok := c.Response().(http.Pusher)

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
