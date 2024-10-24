package store

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const DIR_PERMISSIONS fs.FileMode = 0744
const DEFAULT_PERMISSIONS fs.FileMode = 0644

type LocalStore struct {
	mu   sync.RWMutex
	url  string
	root string
}

var _ Store = (*LocalStore)(nil)

func MakeLocalStore(root string) *LocalStore {
	return &LocalStore{
		root: root,
	}
}

func (l *LocalStore) BaseURL() string {
	return l.url
}

func (l *LocalStore) Store(r io.Reader, path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	err := os.MkdirAll(filepath.Dir(l.getPath(path)), DIR_PERMISSIONS)
	if err != nil {
		return err
	}

	tf, err := CreateTempFile(l.getPath(path))
	if err != nil {
		return err
	}
	defer tf.Cleanup()

	_, err = io.Copy(tf, r)
	if err != nil {
		return err
	}

	err = tf.Save(l.getPath(path))
	if err != nil {
		return err
	}

	err = os.Chmod(l.getPath(path), DEFAULT_PERMISSIONS)
	if err != nil {
		return err
	}

	return nil
}

func (l *LocalStore) Delete(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return nil
}

func (l *LocalStore) Retrieve(path string) (ObjectReader, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return os.Open(l.getPath(path))
}

func (l *LocalStore) Size(path string) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	st, err := os.Stat(l.getPath(path))
	if err != nil {
		return 0, err
	}

	return st.Size(), nil
}

func (l *LocalStore) ModTime(path string) (time.Time, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if st, err := os.Stat(l.getPath(path)); err == nil {
		return st.ModTime(), nil
	} else {
		return time.Time{}, err
	}
}

func (l *LocalStore) getPath(path string) string {
	return filepath.Join(l.root, path)
}
