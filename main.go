package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	stop := make(chan struct{})
	cleanupDone := make(chan struct{})

	setup(stop, cleanupDone)

	r := gin.Default()
	r.Static("/static", "./static")
	renderer := configureHtmlRender(r)
	r.HTMLRender = renderer

	go startUpdateLoop(renderer, r, stop)

	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func setup(stop chan struct{}, cleanupDone chan struct{}) {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	err = resetFiles()

	if err != nil {
		panic(err)
	}

	go setupFileCleanerAtExit(stop, cleanupDone)

	go func() {
		<-cleanupDone
		fmt.Println("Exiting...")
		os.Exit(0)
	}()
}

func configureHtmlRender(engine *gin.Engine) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	r.AddFromFiles("index", "templates/base.gohtml", "templates/pages/index.gohtml")
	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{})
	})
	r.AddFromFiles("about", "templates/base.gohtml", "templates/pages/about.gohtml")
	engine.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about", gin.H{})
	})

	return r
}
