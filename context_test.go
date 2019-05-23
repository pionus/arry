package arry

import (
    "testing"
    "strings"
    "net/http/httptest"
)


func TestJSON(t *testing.T) {
    rec := httptest.NewRecorder()
    ctx := &context{
        response: rec,
    }

    ctx.JSON(200, jim)

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
