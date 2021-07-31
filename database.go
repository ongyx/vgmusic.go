package vgmusic

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/PuerkitoBio/goquery"
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

type DatabaseStats struct {
	NEntries  int
	NConsoles int
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

				fu := strings.TrimSuffix(BASE.ResolveReference(u).String(), "/")

				name := selection.Text()

				infoL.Printf("adding system %s with url %s", name, fu)

				db.Consoles = append(db.Consoles, Console{Name: name, Url: fu})
			},
		)

	}
}

// public functions

func NewDatabase() *Database {
	return &Database{
		Entries: make(map[string]Song),
	}
}

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

	// ensure there are no duplicates if refreshed twice
	db.Consoles = nil
	doc.Find("p.menu").Each(db.parseConsoles)

	return nil
}

func (db *Database) Refresh() error {

	oldStats := db.Stats()

	var g errgroup.Group
	var mu sync.Mutex

	for i := range db.Consoles {

		// https://golang.org/doc/faq#closures_and_goroutines
		i := i

		g.Go(
			func() error {
				// so console can be modified in-place
				c := &db.Consoles[i]

				songs, err := c.ParseSongs()
				nSongs := len(songs)

				if nSongs > 0 {
					// make sure only one thread adds the entries at a time
					mu.Lock()
					for _, song := range songs {
						db.Entries[song.Checksum] = song
					}
					mu.Unlock()
				}

				infoL.Printf("parsed %d songs for console %s", nSongs, c.Name)

				return err
			},
		)

	}

	err := g.Wait()

	newStats := db.Stats()

	infoL.Printf(
		"refreshed database (%d new entries)",
		newStats.NEntries-oldStats.NEntries,
	)

	return err

}

func (db *Database) Stats() DatabaseStats {
	return DatabaseStats{
		NEntries:  len(db.Entries),
		NConsoles: len(db.Consoles),
	}
}

func (db *Database) Search(criteria func(s Song) bool) []Song {
	var songs []Song

	for _, song := range db.Entries {
		if criteria(song) {
			songs = append(songs, song)
		}
	}

	return songs
}

func (db *Database) Dump(w io.Writer) (int, error) {
	infoL.Println("dumping database")

	b, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

func (db *Database) DumpC(w io.Writer) (int, error) {
	infoL.Println("dumping database (compressed as gzip)")

	zw := gzip.NewWriter(w)
	defer zw.Close()

	zw.Name = "database.json"
	zw.ModTime = time.Now().UTC()

	n, err := db.Dump(zw)

	if err := zw.Close(); err != nil {
		return 0, err
	}

	return n, err
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

func (db *Database) LoadC(r io.Reader) error {
	infoL.Println("loading database (compressed as gzip)")

	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	err = db.Load(zr)
	if err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}
