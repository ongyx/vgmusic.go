package vgmusic

type Song struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Size     int    `json:"size"`
	Author   string `json:"author"`
	Game     string `json:"game"`
	Console  string `json:"console"` // The name of the console.
	Checksum string `json:"checksum"`
}
