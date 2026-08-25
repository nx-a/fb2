package domain

// Book is the reader's format-independent representation of an FB2 book.
type Book struct {
	Title   string
	Authors []string
	Text    string
}
