package main

import (
    "net/http"
    "fmt"
    "os"
    "os/signal"
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
    a.Use(middlewares.Gzip)
    a.Use(middlewares.Logger())
    a.Use(middlewares.Panic)

    a.Static("/static", "assets")
    a.Views("assets/")

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

    router.Get("/render", func(ctx arry.Context) {
        ctx.Render(200, "static.html", nil)
    })

    router.Get("/render/1", func(ctx arry.Context) {
        ctx.Render(200, "page1.html", nil)
    })

    router.Get("/render/2", func(ctx arry.Context) {
        ctx.Render(200, "page2.html", nil)
    })
    
    go func() {
        err := a.Start(":8087")

        if err != nil {
            log.Fatalf("Could not start server: %s\n", err.Error())
        }
    }()

    quit := make(chan os.Signal)
    signal.Notify(quit, os.Interrupt)

    <-quit
    log.Printf("shutdown")

}
