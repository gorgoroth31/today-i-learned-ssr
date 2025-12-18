serverside rendered application to display the markdown files in github repository. currently are only repositories supported in which the markdown files are not located in directories.

Stack: Golang 1.24.0 with Gin router

## How to run
- clone repository
- create .env file and enter the relevant environment variables (like mentioned below)
- go build -o md-ssr.application
- execute the built file

Webpages are displayed under localhost:{PORT} (port from .env file or 8080 if not specified)

## Environment variables
In .env file are following variables declared:
- `GITHUB_URL`: URL for github repository
- `ENVIRONMENT`: [prod, dev]
- `REFRESH_TIME`: time in seconds; interval in which the website is being updated
- `PORT`: port to run, defaults to :8080; it is important to add the colon as suffix