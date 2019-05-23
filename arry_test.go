package arry

import (
    "testing"
    "net/http/httptest"
)


type user struct {
    Name string `json:"name"`
    Age int `json:"age"`
}


var jim = &user{
    Name: "jim",
    Age: 26,
}

const jimJSON = `{"name":"jim","age":26}`


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


func TestPost(t *testing.T) {
    arry := New()
    router := arry.Router()

    router.Post("/post", func (ctx Context) {
        ctx.Text(200, "OK")
    })

    code, _ := request("POST", "/post", arry)
    if code != 200 {
        t.Errorf("post is failed")
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



func request(method string, path string, a *Arry) (int, string) {
    req := httptest.NewRequest(method, path, nil)
    rec := httptest.NewRecorder()
    a.ServeHTTP(rec, req)
    return rec.Code, rec.Body.String()
}
