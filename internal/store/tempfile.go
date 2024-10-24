package store

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type TempFile struct {
	*os.File
	Path   string
	closed bool
}

func CreateTempFile(path string) (*TempFile, error) {
	dir, name := filepath.Split(path)
	newPath := filepath.Join(dir, name+"-"+uuid.NewString()+".tmp")

	if err := os.MkdirAll(dir, DIR_PERMISSIONS); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return &TempFile{
		f,
		newPath,
		false,
	}, nil
}

func (t *TempFile) Cleanup() error {
	if t.closed {
		return nil
	}
	t.File.Close()
	t.closed = true
	return os.Remove(t.Path)
}

func (t *TempFile) Save(path string) error {
	if t.closed {
		return nil
	}
	t.File.Close()
	t.closed = true
	return os.Rename(t.Path, path)
}
