package vgmusic

import "github.com/PuerkitoBio/goquery"

type Parser interface {
	Parse(c *Console, doc *goquery.Document) ([]Song, error)
}
