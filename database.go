package vgmusic

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

var (
	BASE *url.URL
)

func init() {
	BASE, _ = url.Parse("https://vgmusic.com")
}

type Database struct {
	// Map of song's md5 checksum to the Song itself.
	Entries map[string]Song `json:"entries"`
	// Map of console name to the Console struct.
	Consoles []Console `json:"consoles"`
}

func NewDatabase() *Database {
	return &Database{
		Entries: make(map[string]Song),
	}
}

// protected functions

func (db *Database) parseConsoles(i int, s *goquery.Selection) {
	// skip the first section
	if i != 0 {

		s.Find("a").Each(
			func(_ int, selection *goquery.Selection) {
				ru, exists := selection.Attr("href")

				if !exists {
					// somehow, an <a> tag without a href.
					return
				}

				// Parse the raw relative url.
				u, err := url.Parse(ru)
				if err != nil {
					warnL.Printf("parsing section url %s failed", ru)
				}

				fu := BASE.ResolveReference(u).String()

				name := selection.Text()

				infoL.Printf("adding system %s with url %s", name, fu)

				db.Consoles = append(db.Consoles, Console{Name: name, Url: fu})
			},
		)

	}
}

// public functions

func (db *Database) ParseConsoles() error {
	resp, err := client.Get(BASE.String())
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

func (db *Database) Refresh() error {
	var mu sync.Mutex
	g := new(errgroup.Group)

	for _, console := range db.Consoles {

		g.Go(func() error {

			mu.Lock()
			defer mu.Unlock()

			songs, err := console.ParseSongs()

			if len(songs) > 0 {
				for _, song := range songs {
					db.Entries[song.Checksum] = song
				}

				infoL.Printf("parsed %d songs for console %s", len(songs), console.Name)

			} else if err != nil {
				infoL.Printf("skipping console %s: no new songs found", console.Name)
			}

			return err
		})

	}

	// block main thread until all songs have been parsed or there was an error.
	return g.Wait()

}

func (db *Database) Dump(w io.Writer) (int, error) {
	infoL.Println("dumping database")
	b, err := json.Marshal(db)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

func (db *Database) Load(r io.Reader) error {
	infoL.Println("loading database")
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &db); err != nil {
		return err
	}

	return nil
}
