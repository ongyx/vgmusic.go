package vgmusic

import (
	"errors"
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

	song := Song{Console: c.Name}

	s.Find("td").Each(
		func(i int, s *goquery.Selection) {
			switch i {

			case 0:
				// url and title
				fname, _ := s.Find("a").Attr("href")
				song.Url = c.Url + "/" + fname
				song.Title = strings.Trim(s.Text(), "\r\n")

			case 1:
				// size
				song.Size, _ = strconv.Atoi(strings.ReplaceAll(s.Text(), " bytes", ""))

			case 2:
				// author
				song.Author = strings.Trim(s.Text(), "\r\n")

			case 3:
				// checksum
				u, _ := s.Find("a").Attr("href")
				song.Checksum = reChecksum.FindStringSubmatch(u)[1]

			}
		},
	)

	return song

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
	} else {
		c.Etag = etag
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return songs, err
	}

	var game string

	// Take only the first table (there might be multiple)
	doc.Find("table").First().Find("tr").Each(
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
