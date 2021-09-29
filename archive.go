package vgmusic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var Archive = archive{}

type archive struct{}

func (archive) Parse(c *Console, doc *goquery.Document) ([]Song, error) {
	var (
		songs []Song
		game  string
	)

	// Take only the first table (there might be multiple)
	doc.Find("table").First().Find("tr").Each(
		func(i int, row *goquery.Selection) {
			// skip first two rows
			if i >= 2 {
				text := strings.TrimSpace(row.Text())

				if text == "" {
					return // probably a spacer
				} else if row.HasClass("header") {
					game = text
					return
				}

				// row is a song, parse
				song := Song{}
				song.parse(row)
				song.Game = game
				song.Console = c.Name

				songs = append(songs, song)
			}
		},
	)

	return songs, nil
}
