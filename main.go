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
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	resetFiles()

	r := gin.Default()
	r.Static("/static", "./static")
	r.HTMLRender = configureHtmlRender(r)

	err = r.Run(":8080")
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

	cloneRepository()
}

func handleMarkdownFiles(r multitemplate.Renderer, engine *gin.Engine) {
	for {
		fmt.Println("Processing files...")

		root := "./templates/markdowns"

		files, err := os.ReadDir(root)

		if err != nil {
			panic(err)
		}

		var pages []string

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			cleanedTitle := file.Name()[:len(file.Name())-3]
			pages = append(pages, cleanedTitle)

			newPath := filepath.Join("templates/generated/", cleanedTitle+".gohtml")

			md, err := os.ReadFile(filepath.Join(root, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			newHtml := mdToHTML(md)

			htmlTemplate := getGoHtmlContent(string(newHtml), cleanedTitle)

			fileAlreadyExists := false

			if _, err := os.Stat(newPath); err == nil {
				fileAlreadyExists = true
			}

			if err := os.WriteFile(newPath, []byte(htmlTemplate), 0666); err != nil {
				log.Fatal(err)
			}

			if fileAlreadyExists {
				continue
			}

			r.AddFromFiles(cleanedTitle, "templates/base.gohtml", newPath)
			engine.GET("/"+cleanedTitle, func(c *gin.Context) {
				c.HTML(http.StatusOK, cleanedTitle, gin.H{})
			})
		}
		updateIndexPage(pages)

		fmt.Println("Processing files done")

		if os.Getenv("ENVIRONMENT") == "prod" {
			time.Sleep(time.Hour * 3)
		} else {
			time.Sleep(time.Second * 10)
		}

		cloneRepository()
	}
}
func updateIndexPage(pages []string) {
	pagesAsList := "<ul>"

	for _, page := range pages {
		pagesAsList += "<li><a href=\"/" + page + "\">" + page + "</a></li>"
	}

	pagesAsList += "</ul>"

	goHtml := getGoHtmlContent(pagesAsList, "index")

	if err := os.WriteFile("templates/pages/index.gohtml", []byte(goHtml), 0666); err != nil {
		log.Fatal(err)
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

func cloneRepository() {
	err := os.RemoveAll("./templates/markdowns")

	_, err = git.PlainClone("./templates/markdowns", &git.CloneOptions{
		URL:      os.Getenv("GITHUB_URL"),
		Progress: os.Stdout,
	})

	if err != nil {
		panic(err)
	}
}
