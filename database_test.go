package vgmusic

import (
	"bytes"
	"fmt"
	"testing"
)

var db *Database = NewDatabase()

func TestDatabaseParseConsoles(t *testing.T) {
	err := db.ParseConsoles()
	if err != nil {
		t.Fatalf(`error while parsing: %v`, err)
	}
}

func TestDatabaseDumpAndLoad(t *testing.T) {
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
