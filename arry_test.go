package arry

import (
    "testing"
    "net/http/httptest"
)


type user struct {
    Name string `json:"name"`
    Age int `json:"age"`
}


func TestArry(t *testing.T) {
    arry := New()
    request("GET", "/", arry)
}


func TestGet(t *testing.T) {
    arry := New()
    router := arry.Router()

    router.Get("/", func(ctx Context) {
		ctx.Text(200, "OK")
	})

    code, _ := request("GET", "/", arry)
    if code != 200 {
        t.Errorf("response is not ok")
    }
}

func TestURLParam(t *testing.T) {
    arry := New()
    router := arry.Router()

    router.Get("/:name", func(ctx Context) {
        name := ctx.Param("name")
		ctx.Text(200, name)
	})

    code, body := request("GET", "/jim", arry)
    if code != 200 || body != "jim" {
        t.Errorf("response is not correct")
    }
}

func TestJSON(t *testing.T) {
    arry := New()
    router := arry.Router()

    router.Get("/", func(ctx Context) {
        jim := &user{
            Name: "jim",
            Age: 26,
        }
		ctx.JSON(200, jim)
	})

    code, body := request("GET", "/", arry)
    if code != 200 {
        t.Errorf("response is not correct, %s", body)
    }
}



func request(method string, path string, a *Arry) (int, string) {
    req := httptest.NewRequest(method, path, nil)
    rec := httptest.NewRecorder()
    a.ServeHTTP(rec, req)
    return rec.Code, rec.Body.String()
}
