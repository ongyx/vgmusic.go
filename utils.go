package vgmusic

import (
	"log"
	"net/http"
	"os"
)

func createLog(level string) *log.Logger {
	return log.New(
		os.Stdout,
		level+" ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
}

func okay(resp *http.Response) bool {
	return (resp.StatusCode >= 200 && resp.StatusCode <= 200)
}
