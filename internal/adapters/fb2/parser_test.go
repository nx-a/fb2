package fb2

import (
	"strings"
	"testing"
)

func TestParserParsesBookAndNestedSections(t *testing.T) {
	input := `<FictionBook><description><title-info><book-title>  The Book  </book-title></title-info></description><body><section><title><p>Chapter one</p></title><p>Hello,   reader.</p><section><p>Nested paragraph.</p></section></section></body></FictionBook>`

	book, err := (Parser{}).Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if book.Title != "The Book" {
		t.Fatalf("Title = %q, want %q", book.Title, "The Book")
	}
	want := "    Chapter one\n    Hello, reader.\n    Nested paragraph."
	if book.Text != want {
		t.Fatalf("Text = %q, want %q", book.Text, want)
	}
}

func TestParserRejectsBookWithoutText(t *testing.T) {
	_, err := (Parser{}).Parse(strings.NewReader(`<FictionBook><description><title-info><book-title>Empty</book-title></title-info></description></FictionBook>`))
	if err == nil {
		t.Fatal("Parse() error = nil, want an error")
	}
}
