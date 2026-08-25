package filesystem

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Store struct{}

func (Store) Open(path string) (io.ReadCloser, error) { return os.Open(path) }
func (Store) FindFB2(root string) ([]string, error) {
	var result []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".fb2") {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}
