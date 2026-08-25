package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreFindFB2(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "book.FB2"), []byte("book"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("notes"), 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := (Store{}).FindFB2(root)
	if err != nil {
		t.Fatalf("FindFB2() error = %v", err)
	}
	if len(files) != 1 || files[0] != filepath.Join(root, "book.FB2") {
		t.Fatalf("FindFB2() = %#v, want the FB2 file", files)
	}
}
