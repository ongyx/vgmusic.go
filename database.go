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
	BASE     *url.URL
	NEWFILES string = "https://www.vgmusic.com/new-files/index.php?&s1=date"
)

func init() {
	BASE, _ = url.Parse("https://vgmusic.com")
}

// DatabaseOptions affects the behaviour of database operations.
type DatabaseOptions struct {
	// Whether or not new-files (not in the archive yet) are to be added to the database.
	// This may cause refreshes to take a long time.
	RefreshNewFiles bool
}

type Database struct {
	// Map of song's md5 checksum to the Song itself.
	Entries map[string]Song `json:"entries"`
	// Map of console name to the Console struct.
	Consoles []Console `json:"consoles"`

	options DatabaseOptions
	mu      sync.RWMutex
}

type stats struct {
	Entries  int
	Consoles int
}

// protected functions

func (db *Database) add(songs ...Song) {
	// make sure only one thread adds the entries at a time
	db.mu.Lock()
	for _, song := range songs {
		db.Entries[song.Checksum] = song
	}
	db.mu.Unlock()
}

func (db *Database) sync(i int, s *goquery.Selection) {
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

				db.Consoles = append(db.Consoles, Console{Name: name, URL: fu})
			},
		)

	}
}

func (db *Database) refresh(c *Console) error {
	var parser Parser

	if strings.Contains(c.URL, "new-files") {
		if !db.options.RefreshNewFiles {
			return nil // skip non-archive files
		}

		parser = NewFile
	} else {
		parser = Archive
	}

	resp, err := download(c.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !c.MustParse(resp) {
		infoL.Printf("skipping console %s, already parsed", c.Name)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	songs, err := parser.Parse(c, doc)
	if err != nil {
		return err
	}

	db.add(songs...)

	infoL.Printf("parsed %d songs for console %s", len(songs), c.Name)

	return nil
}

// public functions

func NewDatabase() *Database {
	return &Database{
		Entries: make(map[string]Song),
	}
}

// Sync the list of consoles from the server.
func (db *Database) Sync() error {
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
	doc.Find("p.menu").Each(db.sync)

	// new files - special console that doesn't exist but can still be parsed
	//db.Consoles = append(db.Consoles, Console{Name: "New Files", URL: NEWFILES})

	return nil
}

// Refresh the database with new songs.
func (db *Database) Refresh() error {

	oldStats := db.Stats()

	var g errgroup.Group

	for i := range db.Consoles {

		// https://golang.org/doc/faq#closures_and_goroutines
		i := i

		g.Go(
			func() error {
				// so console can be modified in-place
				c := &db.Consoles[i]
				return db.refresh(c)
			},
		)

	}

	err := g.Wait()

	newStats := db.Stats()

	infoL.Printf(
		"refreshed database (%d new entries)",
		newStats.Entries-oldStats.Entries,
	)

	return err

}

func (db *Database) Stats() stats {
	return stats{
		Entries:  len(db.Entries),
		Consoles: len(db.Consoles),
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

// Set the database options.
func (db *Database) SetOptions(o DatabaseOptions) {
	db.options = o
}

// Get the database options.
func (db *Database) Options() DatabaseOptions {
	return db.options
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
