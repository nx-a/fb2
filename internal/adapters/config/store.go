package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nx-a/fb2/internal/domain"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentFile string `yaml:"current_file"`
	CurrentUUID string `yaml:"current_uuid"`
	LastDir     string `yaml:"last_dir"`
}

type BookState struct {
	FileName string `yaml:"file_name"`
	Path     string `yaml:"path"`
	Title    string `yaml:"title"`
	Line     int    `yaml:"line"`
}

type Store struct{ dir string }

func (s *Store) ListBooks() ([]BookState, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	books := make([]BookState, 0)
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "config.yml" || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		var state BookState
		if yaml.Unmarshal(data, &state) == nil && state.Path != "" {
			books = append(books, state)
		}
	}
	return books, nil
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "fb2")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) LoadConfig() (Config, error) {
	var value Config
	data, err := os.ReadFile(filepath.Join(s.dir, "config.yml"))
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	if err := yaml.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("read config.yml: %w", err)
	}
	return value, nil
}

func (s *Store) SaveConfig(value Config) error { return s.writeYAML("config.yml", value) }

func (s *Store) StateFor(path string) (string, BookState, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return "", BookState{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "config.yml" || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		var state BookState
		data, readErr := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if readErr != nil {
			return "", BookState{}, readErr
		}
		if unmarshalErr := yaml.Unmarshal(data, &state); unmarshalErr != nil {
			continue
		}
		if state.Path == path {
			return strings.TrimSuffix(entry.Name(), ".yml"), state, nil
		}
	}
	id, err := newUUID()
	if err != nil {
		return "", BookState{}, err
	}
	return id, BookState{FileName: filepath.Base(path), Path: path}, nil
}

func (s *Store) SaveBook(id string, book domain.Book, path string, line int) error {
	return s.writeYAML(id+".yml", BookState{FileName: filepath.Base(path), Path: path, Title: book.Title, Line: line})
}

func (s *Store) writeYAML(name string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, ".tmp-*.yml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0o600); err == nil {
		_, err = tmp.Write(data)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmpName, filepath.Join(s.dir, name))
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
