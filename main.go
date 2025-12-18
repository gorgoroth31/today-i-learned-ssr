package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	renderer := configureHtmlRender(r)
	r.HTMLRender = renderer

	stop := make(chan struct{})

	go setupFileCleanerAtExit(stop)

	go handleMarkdownFiles(renderer, r, stop)

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

	if err := os.WriteFile("templates/pages/index.gohtml", []byte(""), 0666); err != nil {
		log.Fatal(err)
	}
}

func handleMarkdownFiles(r multitemplate.Renderer, engine *gin.Engine, stop chan struct{}) {
	for {
		select {
		case <-stop:
			fmt.Println("stopping")
			return
		default:
		}
		fmt.Println("Processing files...")
		cloneRepository()

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

			linkToStartpage := []byte("<a href=\"/\">Overview</a>")

			md = append(linkToStartpage, md...)

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

		refreshTime := getRefreshTimeFromEnvironment()

		fmt.Printf("Refresh in %d seconds\n", refreshTime)

		time.Sleep(time.Second * time.Duration(refreshTime))
	}
}

func getRefreshTimeFromEnvironment() int {
	variable := os.Getenv("REFRESH_TIME")

	i, err := strconv.Atoi(variable)

	if err != nil {
		panic(err)
	}

	return i
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

func setupFileCleanerAtExit(stop chan struct{}) {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		close(stop)
		fmt.Println("\nReceived an interrupt, deleting generated files...\n")
		err := os.RemoveAll("./templates/generated")
		if err != nil {
			panic(err)
		}

		err = os.Remove("./templates/pages/index.gohtml")

		fmt.Println("Files are deleted and process will finish")
		close(cleanupDone)
		os.Exit(0)
	}()
	<-cleanupDone
}
