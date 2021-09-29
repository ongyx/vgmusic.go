package vgmusic

import "github.com/PuerkitoBio/goquery"

var NewFile = newFile{}

type newFile struct{}

func (newFile) Parse(c *Console, doc *goquery.Document) ([]Song, error) {
	var songs []Song

	doc.Find("table").First().Find("tr").Each(
		func(i int, row *goquery.Selection) {
			if i >= 1 {
				song := Song{}
				song.parseNewFile(row)

				songs = append(songs, song)
			}
		},
	)

	return songs, nil
}
