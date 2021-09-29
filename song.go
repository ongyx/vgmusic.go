package vgmusic

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	reChecksum = regexp.MustCompile(`/file/(.*?)\.html`)
)

// Song is a midi file hosted on VGMusic.
type Song struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Size     int    `json:"size"`
	Author   string `json:"author"`
	Game     string `json:"game"`
	Console  string `json:"console"` // The name of the console.
	Checksum string `json:"checksum"`
}

func (s *Song) setURL(cell *goquery.Selection) {
	u, _ := cell.Find("a").Attr("href")
	s.URL = u
}

func (s *Song) setSize(cell *goquery.Selection) {
	size, _ := strconv.Atoi(strings.ReplaceAll(cell.Text(), " bytes", ""))
	s.Size = size
}

func (s *Song) setChecksum(cell *goquery.Selection) {
	u, _ := cell.Find("a").Attr("href")
	s.Checksum = reChecksum.FindStringSubmatch(u)[1]
}

func (s *Song) parse(row *goquery.Selection) {
	row.Find("td").Each(
		func(i int, cell *goquery.Selection) {
			switch i {

			// delegate based on data cell index.
			case 0: // url and title
				s.setURL(cell)
				s.Title = toString(cell)

			case 1: // size
				s.setSize(cell)

			case 2: // author
				s.Author = toString(cell)

			case 3: // checksum
				s.setChecksum(cell)

			}
		},
	)
}

func (s *Song) parseNewFile(row *goquery.Selection) {
	row.Find("td").Each(
		func(i int, cell *goquery.Selection) {
			switch i {

			case 0: // upload time (ignored)

			case 1: // console name
				s.Console = toString(cell)

			case 2: // game name
				s.Game = toString(cell)

			case 3: // url and title
				s.setURL(cell)

			case 4: // author
				s.Author = toString(cell)

			case 5: // size
				s.setSize(cell)

			case 6: // checksum
				s.setChecksum(cell)

			}
		},
	)
}
