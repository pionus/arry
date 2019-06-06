package main

import (
    "net/http"
    "fmt"
    "log"
    "github.com/pionus/arry"
    "github.com/pionus/arry/middlewares"
)



type user struct {
    Name string `json:"name"`
    Age int `json:"age"`
}



func main() {
	a := arry.New()
    a.Use(middlewares.Logger)
    a.Use(middlewares.Panic)
    a.Use(middlewares.Gzip)

    a.Static("/static", "_example/assets")

    router := a.Router()

    router.Get("/", func(ctx arry.Context) {
		ctx.Text(http.StatusOK, "index")
	})

	router.Get(`/hello`, func(ctx arry.Context) {
		ctx.Text(http.StatusOK, "Hello world")
	})

	router.Get(`/hello/:name`, func(ctx arry.Context) {
        fmt.Printf("hello %s", ctx.Param("name"))
		ctx.Text(http.StatusOK, fmt.Sprintf("Hello %s", ctx.Param("name")))
	})

    router.Get("/panic", func(ctx arry.Context) {
        panic(123)
    })

    router.Get("/push", func(ctx arry.Context) {
        ctx.Push("/static/css/test.css")
        ctx.Text(http.StatusOK, "pushed~")
    })

    router.Get("/json", func(ctx arry.Context) {
        jim := user{
            Name: "Jim",
            Age: 26,
        }
        ctx.JSON(http.StatusOK, jim)
    })

	err := a.Start(":8087")

    // err := app.StartTLS(":8087", "_example/server.crt", "_example/server.key")

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
