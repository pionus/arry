package arry

import (
    "testing"
    "strings"
    "bytes"
    "encoding/json"
    "net/http/httptest"
)


func TestJSON(t *testing.T) {
    rec := httptest.NewRecorder()
    ctx := &context{
        response: &Response{Writer: rec},
    }

    ctx.JSON(200, jim)

    if rec.Code != 200 {
        t.Errorf("response is not correct, %s", rec.Body.String())
    }
}

func TestJSONBlob(t *testing.T) {
    rec := httptest.NewRecorder()
    ctx := &context{
        response: &Response{Writer: rec},
    }

    data, _ := json.Marshal(jim)
    ctx.JSONBlob(200, data)

    if rec.Code != 200 {
        t.Errorf("response is not correct, %s", rec.Body.String())
    }
}


func TestDecode(t *testing.T) {
    req := httptest.NewRequest("POST", "/", strings.NewReader(jimJSON))
    ctx := &context{
        request: req,
    }

    var u user
    ctx.Decode(&u)

    if u.Age != 26 {
        t.Errorf("json decode failed")
    }
}

// TestContextBodyMultipleReads tests that Body() can be read multiple times
func TestContextBodyMultipleReads(t *testing.T) {
    body := []byte(`{"name":"test","value":123}`)
    req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // First read
    body1 := ctx.Body()
    if string(body1) != string(body) {
        t.Errorf("first Body() read = %q, want %q", string(body1), string(body))
    }

    // Second read should return the same data
    body2 := ctx.Body()
    if string(body2) != string(body) {
        t.Errorf("second Body() read = %q, want %q", string(body2), string(body))
    }

    // Should be identical
    if string(body1) != string(body2) {
        t.Error("multiple Body() reads returned different data")
    }
}

// TestContextBodyAndDecode tests Body() and Decode() compatibility
func TestContextBodyAndDecode(t *testing.T) {
    jsonData := `{"name":"John","age":30}`
    req := httptest.NewRequest("POST", "/", strings.NewReader(jsonData))
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Read body first
    body := ctx.Body()
    if len(body) == 0 {
        t.Error("Body() returned empty")
    }

    // Then decode should still work
    var result map[string]interface{}
    err := ctx.Decode(&result)
    if err != nil {
        t.Errorf("Decode() after Body() failed: %v", err)
    }

    if result["name"] != "John" {
        t.Errorf("Decode() name = %v, want John", result["name"])
    }
}

// TestContextQuery tests Query methods
func TestContextQuery(t *testing.T) {
    req := httptest.NewRequest("GET", "/?foo=bar&page=1", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Test Query
    if got := ctx.Query("foo"); got != "bar" {
        t.Errorf("Query(foo) = %q, want bar", got)
    }

    if got := ctx.Query("page"); got != "1" {
        t.Errorf("Query(page) = %q, want 1", got)
    }

    if got := ctx.Query("notexist"); got != "" {
        t.Errorf("Query(notexist) = %q, want empty string", got)
    }
}

// TestContextQueryDefault tests QueryDefault method
func TestContextQueryDefault(t *testing.T) {
    req := httptest.NewRequest("GET", "/?foo=bar", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Test with existing key
    if got := ctx.QueryDefault("foo", "default"); got != "bar" {
        t.Errorf("QueryDefault(foo) = %q, want bar", got)
    }

    // Test with non-existing key
    if got := ctx.QueryDefault("missing", "default"); got != "default" {
        t.Errorf("QueryDefault(missing) = %q, want default", got)
    }
}

// TestContextQueryParams tests QueryParams method
func TestContextQueryParams(t *testing.T) {
    req := httptest.NewRequest("GET", "/?foo=bar&page=1&tags=a&tags=b", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    params := ctx.QueryParams()

    if params.Get("foo") != "bar" {
        t.Errorf("QueryParams().Get(foo) = %q, want bar", params.Get("foo"))
    }

    tags := params["tags"]
    if len(tags) != 2 {
        t.Errorf("QueryParams()[tags] length = %d, want 2", len(tags))
    }
}

// TestContextHeader tests Header methods
func TestContextHeader(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("User-Agent", "test-agent")
    req.Header.Set("X-Custom", "custom-value")
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Test Header
    if got := ctx.Header("User-Agent"); got != "test-agent" {
        t.Errorf("Header(User-Agent) = %q, want test-agent", got)
    }

    if got := ctx.Header("X-Custom"); got != "custom-value" {
        t.Errorf("Header(X-Custom) = %q, want custom-value", got)
    }

    if got := ctx.Header("NotExist"); got != "" {
        t.Errorf("Header(NotExist) = %q, want empty string", got)
    }
}

// TestContextSetHeader tests SetHeader method
func TestContextSetHeader(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Set header
    ctx.SetHeader("X-Test", "test-value")

    // Check response header
    if got := rec.Header().Get("X-Test"); got != "test-value" {
        t.Errorf("SetHeader(X-Test) set %q, want test-value", got)
    }
}

// TestContextRedirect tests Redirect method
func TestContextRedirect(t *testing.T) {
    req := httptest.NewRequest("GET", "/old", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    // Test redirect
    ctx.Redirect(302, "/new")

    if rec.Code != 302 {
        t.Errorf("Redirect status = %d, want 302", rec.Code)
    }

    if got := rec.Header().Get("Location"); got != "/new" {
        t.Errorf("Redirect Location = %q, want /new", got)
    }
}

// TestContextStream tests Stream method
func TestContextStream(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    data := strings.NewReader("streaming data content")
    err := ctx.Stream(200, "text/plain", data)

    if err != nil {
        t.Errorf("Stream() error = %v", err)
    }

    if rec.Code != 200 {
        t.Errorf("Stream status = %d, want 200", rec.Code)
    }

    if got := rec.Header().Get("Content-Type"); got != "text/plain" {
        t.Errorf("Stream Content-Type = %q, want text/plain", got)
    }

    if got := rec.Body.String(); got != "streaming data content" {
        t.Errorf("Stream body = %q, want streaming data content", got)
    }
}

// TestContextAttachment tests Attachment method
func TestContextAttachment(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    ctx := NewContext(req, rec).(*context)

    data := strings.NewReader("file content here")
    err := ctx.Attachment("test.txt", data)

    if err != nil {
        t.Errorf("Attachment() error = %v", err)
    }

    if rec.Code != 200 {
        t.Errorf("Attachment status = %d, want 200", rec.Code)
    }

    if got := rec.Header().Get("Content-Type"); got != "application/octet-stream" {
        t.Errorf("Attachment Content-Type = %q, want application/octet-stream", got)
    }

    if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, "attachment") || !strings.Contains(got, "test.txt") {
        t.Errorf("Attachment Content-Disposition = %q, want attachment with filename", got)
    }

    if got := rec.Body.String(); got != "file content here" {
        t.Errorf("Attachment body = %q, want file content here", got)
    }
}

// TestContextClientIP tests ClientIP method
func TestContextClientIP(t *testing.T) {
    tests := []struct {
        name           string
        remoteAddr     string
        xff            string
        xri            string
        expectedIP     string
    }{
        {
            name:       "X-Forwarded-For single IP",
            remoteAddr: "192.168.1.1:8080",
            xff:        "203.0.113.1",
            expectedIP: "203.0.113.1",
        },
        {
            name:       "X-Forwarded-For multiple IPs",
            remoteAddr: "192.168.1.1:8080",
            xff:        "203.0.113.1, 198.51.100.1, 192.168.1.1",
            expectedIP: "203.0.113.1",
        },
        {
            name:       "X-Real-IP",
            remoteAddr: "192.168.1.1:8080",
            xri:        "203.0.113.2",
            expectedIP: "203.0.113.2",
        },
        {
            name:       "RemoteAddr fallback",
            remoteAddr: "192.168.1.1:8080",
            expectedIP: "192.168.1.1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/", nil)
            req.RemoteAddr = tt.remoteAddr
            if tt.xff != "" {
                req.Header.Set("X-Forwarded-For", tt.xff)
            }
            if tt.xri != "" {
                req.Header.Set("X-Real-IP", tt.xri)
            }

            rec := httptest.NewRecorder()
            ctx := NewContext(req, rec).(*context)

            if got := ctx.ClientIP(); got != tt.expectedIP {
                t.Errorf("ClientIP() = %q, want %q", got, tt.expectedIP)
            }
        })
    }
}

// TestContextBind tests Bind method
func TestContextBind(t *testing.T) {
    type TestData struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }

    t.Run("JSON binding", func(t *testing.T) {
        jsonData := `{"name":"Alice","age":25}`
        req := httptest.NewRequest("POST", "/", strings.NewReader(jsonData))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        ctx := NewContext(req, rec).(*context)

        var data TestData
        err := ctx.Bind(&data)

        if err != nil {
            t.Errorf("Bind() error = %v", err)
        }

        if data.Name != "Alice" {
            t.Errorf("Bind() Name = %q, want Alice", data.Name)
        }

        if data.Age != 25 {
            t.Errorf("Bind() Age = %d, want 25", data.Age)
        }
    })

    t.Run("Form binding", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/", strings.NewReader("name=Bob&age=30"))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        rec := httptest.NewRecorder()
        ctx := NewContext(req, rec).(*context)

        var data TestData
        err := ctx.Bind(&data)

        // Form binding doesn't fully populate struct yet, but shouldn't error
        if err != nil {
            t.Errorf("Bind() form error = %v", err)
        }
    })
}
