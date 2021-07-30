package vgmusic

import (
	"errors"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	reChecksum = regexp.MustCompile(`/file/(.*?)\.html`)
)

type Console struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Etag string `json:"etag"`
}

func (c *Console) parseSong(s *goquery.Selection) Song {
	// first <td> tag has a <a href...> inside
	nameSel := s.Eq(0).First()
	title := nameSel.Text()
	fname, _ := nameSel.Attr("href")
	u := path.Join(c.Url, fname)

	size, _ := strconv.Atoi(strings.ReplaceAll(s.Eq(1).Text(), " bytes", ""))

	author := s.Eq(2).Text()

	// checksum is stored as a url
	// extract the checksum out of the url
	infoL.Println(s.Eq(3).First())
	infoUrl, _ := s.Eq(3).First().Attr("href")
	infoL.Println(u, title, size, author, s)
	checksum := reChecksum.FindStringSubmatch(infoUrl)[1]

	return Song{
		Url:      u,
		Title:    title,
		Size:     size,
		Author:   author,
		Console:  c.Name,
		Checksum: checksum,
	}
}

// Parse songs from this console into a slice.
// If the songs have already been parsed, the slice will be empty.
func (c *Console) ParseSongs() ([]Song, error) {
	var songs []Song

	resp, err := client.Get(c.Url)
	if err != nil {
		return songs, err
	}

	defer resp.Body.Close()

	if !okay(resp) {
		return songs, errors.New("response not ok: " + strconv.Itoa(resp.StatusCode))
	}

	// check ETag if the content has changed.
	// If the console has not been parsed before, the ETag field will be empty.
	etag := resp.Header.Get("ETag")
	if etag == c.Etag {
		return songs, nil // bail: console has been cached in the database already
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return songs, err
	}

	var game string

	doc.Find("tr").Each(
		func(i int, s *goquery.Selection) {
			// skip first two rows
			if i >= 2 {
				text := strings.TrimSpace(s.Text())

				if text == "" {
					return // probably a spacer
				} else if s.HasClass("header") {
					game = text
					return
				}

				// row is a song, parse
				song := c.parseSong(s)
				song.Game = game

				songs = append(songs, song)
			}
		},
	)

	return songs, nil

}
