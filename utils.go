package vgmusic

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	infoL, warnL, errorL *log.Logger

	client *http.Client
)

func init() {
	infoL = createLog("info")
	warnL = createLog("warning")
	errorL = createLog("error")

	// custom	client is used, in case the server takes too long to respond
	client = &http.Client{
		Timeout: time.Second * 10,
	}
}

func createLog(level string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("%s ", strings.ToUpper(level)), 0)
}

func okay(resp *http.Response) bool {
	return (resp.StatusCode >= 200 && resp.StatusCode <= 200)
}

func toString(s *goquery.Selection) string {
	return strings.Trim(s.Text(), "\r\n")
}

func download(u string) (*http.Response, error) {
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}

	if !okay(resp) {
		return nil, errors.New("response not ok: " + strconv.Itoa(resp.StatusCode))
	}

	return resp, err
}
