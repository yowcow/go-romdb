package jsonstorage

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/yowcow/goromdb/storage"
)

var (
	_ storage.Storage = (*Storage)(nil)
)

// Data represents a data
type Data map[string]interface{}

// Storage represents a JSON storage
type Storage struct {
	gzipped bool
	data    *atomic.Value
	mux     *sync.RWMutex
}

// New creates and returns a storage
func New(gzipped bool) *Storage {
	return &Storage{gzipped, new(atomic.Value), new(sync.RWMutex)}
}

// Load loads data into storage
func (s *Storage) Load(file string) error {
	data, err := s.openFile(file)
	if err != nil {
		return err
	}

	// Lock, switch, and unlock
	s.mux.Lock()
	defer s.mux.Unlock()

	s.data.Store(data)
	return nil
}

func (s Storage) openFile(file string) (Data, error) {
	fi, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	r, err := s.newReader(fi)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(r)
	var data Data
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s Storage) newReader(rdr io.Reader) (io.Reader, error) {
	if s.gzipped {
		r, err := gzip.NewReader(rdr)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return rdr, nil
}

// Get finds a given key in data, and returns its value
func (s Storage) Get(key []byte) ([]byte, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	ptr := s.data.Load()
	if ptr == nil {
		return nil, storage.KeyNotFoundError(key)
	}
	data := ptr.(Data)
	k := string(key)
	if v, ok := data[k]; ok {
		return []byte(v.(string)), nil
	}
	return nil, storage.KeyNotFoundError(key)
}

func iterate(data Data, fn storage.IterationFunc) error {
	for k, v := range data {
		if err := fn([]byte(k), []byte(v.(string))); err != nil {
			return err
		}
	}
	return nil
}
