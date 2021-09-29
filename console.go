package vgmusic

import "net/http"

type Console struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Etag string `json:"etag"`
}

// check the response's ETag if the content has changed.
// If the console has not been parsed before, the ETag field will be empty.
func (c *Console) MustParse(resp *http.Response) bool {
	return resp.Header.Get("ETag") != c.Etag
}
