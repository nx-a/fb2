package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nx-a/fb2/internal/domain"
	"github.com/nx-a/fb2/internal/ports"
)

type Reader struct {
	files     ports.FileStore
	downloads ports.Downloader
	parser    ports.BookParser
}

func NewReader(files ports.FileStore, downloads ports.Downloader, parser ports.BookParser) *Reader {
	return &Reader{files: files, downloads: downloads, parser: parser}
}

func (r *Reader) OpenFile(path string) (domain.Book, error) {
	f, err := r.files.Open(path)
	if err != nil {
		return domain.Book{}, err
	}
	defer f.Close()
	return r.parser.Parse(f)
}

func (r *Reader) Download(ctx context.Context, url string) (domain.Book, string, error) {
	f, err := r.downloads.Download(ctx, url)
	if err != nil {
		return domain.Book{}, "", err
	}
	defer f.Close()
	book, err := r.parser.Parse(f)
	return book, filepath.Base(url), err
}

func (r *Reader) Search(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	return r.files.FindFB2(path)
}
