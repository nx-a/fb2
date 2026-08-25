package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nx-a/fb2/internal/domain"
)

func TestStoreSavesAndLoadsSettings(t *testing.T) {
	store := &Store{dir: t.TempDir()}
	wantConfig := Config{CurrentFile: "/books/a.fb2", CurrentUUID: "book-id", LastDir: "/books"}
	if err := store.SaveConfig(wantConfig); err != nil {
		t.Fatal(err)
	}
	gotConfig, err := store.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if gotConfig != wantConfig {
		t.Fatalf("config = %#v, want %#v", gotConfig, wantConfig)
	}

	book := domain.Book{Title: "A book"}
	if err := store.SaveBook("book-id", book, "/books/a.fb2", 42); err != nil {
		t.Fatal(err)
	}
	id, state, err := store.StateFor("/books/a.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if id != "book-id" || state.Title != book.Title || state.Line != 42 {
		t.Fatalf("state = %q %#v", id, state)
	}
	if _, err := os.Stat(filepath.Join(store.dir, "book-id.yml")); err != nil {
		t.Fatal(err)
	}
}
