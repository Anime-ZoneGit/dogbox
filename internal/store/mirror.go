package store

import (
	"errors"
	"io"
	"sync"
	"time"
)

// A Mirror uses multiple backing stores to retrieve data.
// Store operations will write the given file to all of the backing stores,
// and returns true if at least one of the uploads is successful.
// Retrieval operations (retrieval, exists, information, etc.) will try all of
// the backing stores in sequential order given in the stores list.
type Mirror struct {
	stores []Store
	url    string
}

var _ Store = (*Mirror)(nil)

func (m *Mirror) BaseURL() string {
	return m.url
}

func (m *Mirror) Store(r io.Reader, path string) error {
	mirrorReaders := make([]io.Reader, len(m.stores))
	mirrorWriters := make([]io.Writer, len(m.stores))

	for i := 0; i < len(m.stores); i++ {
		mirrorReaders[i], mirrorWriters[i] = io.Pipe()
	}

	joinWriter := io.MultiWriter(mirrorWriters...)

	readHeadErr := make(chan error)
	writeErrs := make([]error, len(m.stores))
	finish := make(chan struct{})

	go func() {
		_, err := io.Copy(joinWriter, r)
		if err != nil {
			readHeadErr <- err
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < len(m.stores); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.stores[i].Store(mirrorReaders[i], path)
			if err != nil {
				writeErrs[i] = err
			}
		}()
	}

	go func() {
		wg.Wait()
		finish <- struct{}{}
	}()

	allErrs := errors.Join(writeErrs...)

	select {
	case err := <-readHeadErr:
		allErrs = errors.Join(allErrs, err)
	case <-finish:
	}

	if allErrs != nil {
		// Undo all stores that have succeeded if one of them has failed
		for i := 0; i < len(m.stores); i++ {
			if writeErrs[i] == nil {
				dErr := m.stores[i].Delete(path)
				if dErr != nil {
					allErrs = errors.Join(allErrs, dErr)
					return allErrs
				}
			}
		}
	}

	return nil
}

func (m *Mirror) Delete(path string) error {
	return nil
}

func (m *Mirror) Retrieve(path string) (ObjectReader, error) {
	for _, st := range m.stores {
		r, err := st.Retrieve(path)
		if err != nil {
			continue
		}

		return r, nil
	}

	return nil, errors.New("Could not find file in any backing store")
}

func (m *Mirror) Size(path string) (int64, error) {
	return 0, nil
}

func (m *Mirror) ModTime(path string) (time.Time, error) {
	return time.Now(), nil
}
