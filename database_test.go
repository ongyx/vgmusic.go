package vgmusic

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

var db *Database = NewDatabase()

func TestParseConsoles(t *testing.T) {
	err := db.ParseConsoles()
	if err != nil {
		t.Fatalf(`error while parsing: %v`, err)
	}
}

func TestDumpAndLoad(t *testing.T) {
	var buf bytes.Buffer

	_, err := db.Dump(&buf)
	if err != nil {
		t.Fatalf(`error while dumping: %v`, err)
	}

	fmt.Println(buf.String())

	// reset db and reload
	db = NewDatabase()
	err = db.Load(&buf)

	if err != nil {
		t.Fatalf(`error while loading: %v`, err)
	}
}

func TestRefresh(t *testing.T) {
	err := db.Refresh()

	if err != nil {
		t.Fatalf(`error while refreshing: %v`, err)
	}

	// save to file for debugging purposes
	f, err := os.OpenFile("debug.json", os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf(`error while saving to disk: %v`, err)
	}

	_, err = db.Dump(f)
	if err != nil {
		t.Fatalf(`error while saving to disk: %v`, err)
	}

}
