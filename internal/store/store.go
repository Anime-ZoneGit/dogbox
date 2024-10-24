package store

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"time"
)

const CTX_COPY_BUF_SIZE int64 = 32 * 1024

type ObjectReader interface {
	io.Reader
	io.Seeker
	io.Closer
}

type ObjectWriter struct {
	*io.PipeWriter
}

type Store interface {
	// Returns the base URL where images in the store are served from.
	BaseURL() string
	// Write the contents of the io.Reader into the store at the given
	// location. If a file exists at that location, it should be overwritten.
	// Writes should be atomic- either the file is completely written to the
	// store
	Store(r io.Reader, path string) error

	// Delete the file at the given path.
	Delete(path string) error

	// Returns an object that reads the file at the given string.
	Retrieve(path string) (ObjectReader, error)

	Size(path string) (int64, error)
	ModTime(path string) (time.Time, error)
}

func Copy(s Store, src, dst string) error {
	reader, err := s.Retrieve(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	err = s.Store(reader, dst)

	if err != nil {
		return err
	}

	return nil
}

func Move(s Store, src, dst string) error {
	err := Copy(s, src, dst)
	if err != nil {
		return err
	}
	return s.Delete(src)
}

func FileURL(s Store, path string) string {
	return filepath.Join(s.BaseURL(), path)
}

// Context-aware variant of the io.Copy function: will stop once the given
// Context is canceled.
func ContextCopy(
	ctx context.Context,
	dst io.Writer,
	src io.Reader,
) (written int64, err error) {
	bufSize := CTX_COPY_BUF_SIZE
	written = int64(0)

	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}

		n, err := io.CopyN(dst, src, bufSize)
		written += n
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return written, err
		}
	}
}

func NewWriter(s Store, path string) *ObjectWriter {
	r, w := io.Pipe()
	go func() {
		err := s.Store(r, path)
		if err != nil {
			r.CloseWithError(err)
			return
		}
		r.Close()
	}()

	return &ObjectWriter{
		w,
	}
}
