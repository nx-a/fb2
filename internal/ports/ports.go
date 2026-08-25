package ports

import (
	"context"
	"io"

	"github.com/nx-a/fb2/internal/domain"
)

type BookParser interface {
	Parse(io.Reader) (domain.Book, error)
}

type FileStore interface {
	Open(string) (io.ReadCloser, error)
	FindFB2(string) ([]string, error)
}

type Downloader interface {
	Download(context.Context, string) (io.ReadCloser, error)
}
