package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-git/go-git/v6"
)

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

func getRefreshTimeFromEnvironment() int {
	variable := os.Getenv("REFRESH_TIME")

	i, err := strconv.Atoi(variable)

	if err != nil {
		panic(err)
	}

	return i
}

func getPortFromEnvironment() string {
	variable := os.Getenv("PORT")

	if variable == "" {
		fmt.Println("empty")
		return ":8080"
	}

	return variable
}
