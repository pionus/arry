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
