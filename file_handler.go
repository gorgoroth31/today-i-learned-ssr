package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func resetFiles() error {
	err := os.RemoveAll("./templates/generated")

	if err != nil {
		return err
	}

	err = os.Mkdir("./templates/generated", 0755)

	if err != nil {
		return err
	}

	return os.WriteFile("templates/pages/index.gohtml", []byte(""), 0666)
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

func startMarkdownLoop(r multitemplate.Renderer, engine *gin.Engine, stop chan struct{}) {
	for {
		select {
		case <-stop:
			fmt.Println("stopping")
			break
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

func setupFileCleanerAtExit(stop chan struct{}, cleanupDone chan struct{}) {
	signalChan := make(chan os.Signal, 1)
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
	}()
}
