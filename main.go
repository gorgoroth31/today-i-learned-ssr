package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v6"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func main() {
	resetFiles()

	r := gin.Default()
	r.HTMLRender = configureHtmlRender(r)

	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
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

	go handleMarkdownFiles(r, engine)

	return r
}

func resetFiles() {
	err := os.RemoveAll("./templates/generated")

	if err != nil {
		panic(err)
	}

	err = os.Mkdir("./templates/generated", 0755)

	if err != nil {
		panic(err)
	}

	err = os.RemoveAll("./templates/markdowns")

	_, err = git.PlainClone("./templates/markdowns", &git.CloneOptions{
		URL:      "https://github.com/gorgoroth31/today-i-learned",
		Progress: os.Stdout,
	})

	if err != nil {
		panic(err)
	}
}

func handleMarkdownFiles(r multitemplate.Renderer, engine *gin.Engine) {
	for {
		fmt.Println("Processing files...")

		root := "./templates/markdowns"

		files, err := os.ReadDir(root)

		if err != nil {
			panic(err)
		}

		for _, file := range files {
			cleanedTitle := file.Name()[:len(file.Name())-3]
			newPath := filepath.Join("templates/generated/", cleanedTitle+".gohtml")

			if _, err := os.Stat(newPath); err == nil {
				continue
			}

			md, err := os.ReadFile(filepath.Join(root, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			newHtml := mdToHTML(md)

			htmlTemplate := getGoHtmlContent(string(newHtml), cleanedTitle)

			if err := os.WriteFile(newPath, []byte(htmlTemplate), 0666); err != nil {
				log.Fatal(err)
			}

			r.AddFromFiles(cleanedTitle, "templates/base.gohtml", newPath)
			engine.GET("/"+cleanedTitle, func(c *gin.Context) {
				c.HTML(http.StatusOK, cleanedTitle, gin.H{})
			})
		}

		time.Sleep(time.Second * 60)
	}
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func getGoHtmlContent(htmlTemplate string, title string) string {
	return `{{/*` + title + `.gohtml*/}}

{{ define "title" }}` + title + `{{ end }}

{{ define "body" }}
` + htmlTemplate + `
{{ end }}
`
}
