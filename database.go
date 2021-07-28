package vgmusic

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	URL = "https://vgmusic.com"
)

var (
	debug *log.Logger = createLog("debug")
)

type Database struct {
	// Map of song's md5 checksum to the Song itself.
	Entries map[string]Song `json:"entries"`
	// Map of console name to the Console struct.
	Consoles map[string]Console `json:"consoles"`
}

func NewDatabase() *Database {
	return &Database{
		Entries:  make(map[string]Song),
		Consoles: make(map[string]Console),
	}
}

func (db *Database) parseConsoles(i int, s *goquery.Selection) {
	// skip the first section
	if i != 0 {

		s.Find("a").Each(
			func(_ int, selection *goquery.Selection) {
				url, _ := selection.Attr("href")
				name := selection.Text()

				debug.Printf("adding system %s with url %s", name, url)

				db.Consoles[name] = Console{Url: url}
			},
		)

	}
}

func (db *Database) ParseConsoles() error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !okay(resp) {
		return errors.New("response not ok: " + strconv.Itoa(resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	doc.Find("p.menu").Each(db.parseConsoles)

	return nil
}

func (db *Database) Dump(w io.Writer) (int, error) {
	b, err := json.Marshal(db)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

func (db *Database) Load(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &db); err != nil {
		return err
	}

	return nil
}
