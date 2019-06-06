package arry

import (
    "testing"
    "net/http/httptest"
)


func TestResponse(t *testing.T) {
    rec := httptest.NewRecorder()
    res := &Response{
        Writer: rec,
        Code: 200,
    }

    res.Write([]byte("test"))

    if rec.Code != 200 {
        t.Errorf("response is not correct, %s", rec.Body.String())
    }
}
