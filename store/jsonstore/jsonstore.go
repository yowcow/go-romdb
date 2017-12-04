package jsonstore

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/yowcow/goromdb/store"
)

type Data map[string]string

type Store struct {
	data    Data
	filein  <-chan string
	gzipped bool
	mux     *sync.RWMutex
	logger  *log.Logger
}

func New(filein <-chan string, gzipped bool, logger *log.Logger) (store.Store, error) {
	var data Data
	return &Store{
		data,
		filein,
		gzipped,
		new(sync.RWMutex),
		logger,
	}, nil
}

func (s *Store) Start() <-chan bool {
	done := make(chan bool)
	go s.start(done)
	return done
}

func (s *Store) start(done chan<- bool) {
	defer func() {
		s.logger.Println("jsonstore finished")
		close(done)
	}()
	s.logger.Println("jsonstore started")
	for file := range s.filein {
		if err := s.Load(file); err != nil {
			s.logger.Printf("jsonstore failed loading data from '%s': %s", file, err.Error())
		} else {
			s.logger.Printf("jsonstore succeeded loading data from '%s'", file)
		}
	}
}

func (s *Store) Load(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := store.NewReader(f, s.gzipped)
	if err != nil {
		return err
	}

	var data Data
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.data = data
	return nil
}

func (s Store) Get(key []byte) ([]byte, []byte, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if val, ok := s.data[string(key)]; ok {
		return key, []byte(val), nil
	}
	return nil, nil, store.KeyNotFoundError(key)
}
