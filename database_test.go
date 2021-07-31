package vgmusic

import (
	"os"
	"testing"
)

const (
	DEBUGFILE = "database.json.gz"
)

var (
	db *Database = NewDatabase()
)

func TestLoad(t *testing.T) {
	f, err := os.OpenFile(DEBUGFILE, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic("failed to open debug.json")
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.Size() == 0 {
		// debug file does not exist, refresh
		err := db.ParseConsoles()
		if err != nil {
			t.Fatalf(`error while parsing: %v`, err)
		}

	} else {
		err = db.LoadC(f)
		if err != nil {
			t.Fatalf(`error while loading: %v`, err)
		}
	}
}

func TestRefresh(t *testing.T) {
	err := db.Refresh()

	if err != nil {
		t.Fatalf(`error while refreshing: %v`, err)
	}

}

func TestDump(t *testing.T) {
	f, err := os.OpenFile(DEBUGFILE, os.O_WRONLY, 0644)
	if err != nil {
		panic("failed to open debug.json")
	}
	defer f.Close()

	_, err = db.DumpC(f)
	if err != nil {
		t.Fatalf(`error while dumping: %v`, err)
	}
}
