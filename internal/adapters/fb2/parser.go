package fb2

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/nx-a/fb2/internal/domain"
)

type Parser struct{}

type document struct {
	Description description `xml:"description"`
	Bodies      []body      `xml:"body"`
}
type description struct {
	Titles []string `xml:"title-info>book-title"`
}
type body struct {
	Sections   []section `xml:"section"`
	Paragraphs []string  `xml:"p"`
}
type section struct {
	Title      title     `xml:"title"`
	Paragraphs []string  `xml:"p"`
	Sections   []section `xml:"section"`
}
type title struct {
	Paragraphs []string `xml:"p"`
}

func (Parser) Parse(r io.Reader) (domain.Book, error) {
	var doc document
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return domain.Book{}, fmt.Errorf("parse FB2: %w", err)
	}
	book := domain.Book{}
	if len(doc.Description.Titles) > 0 {
		book.Title = clean(doc.Description.Titles[0])
	}
	var parts []string
	for _, b := range doc.Bodies {
		parts = append(parts, b.Paragraphs...)
		for _, s := range b.Sections {
			collect(s, &parts)
		}
	}
	book.Text = strings.Join(nonEmpty(parts), "\n")
	if book.Title == "" {
		book.Title = "Без названия"
	}
	if book.Text == "" {
		return domain.Book{}, fmt.Errorf("FB2 does not contain readable text")
	}
	return book, nil
}

func collect(s section, out *[]string) {
	*out = append(*out, s.Title.Paragraphs...)
	*out = append(*out, s.Paragraphs...)
	for _, child := range s.Sections {
		collect(child, out)
	}
}
func clean(s string) string { return strings.Join(strings.Fields(s), " ") }
func nonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		if p = clean(p); p != "" {
			if !strings.HasPrefix(p, "-") {
				p = "    " + p
			}
			out = append(out, p)
		}
	}
	return out
}
