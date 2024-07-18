package main

import (
	"github.com/boltdb/bolt"
)

type Store struct {
	db *bolt.DB
}

func NewStore() (*Store, error) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) AddMovie(movie string) error {
	return nil
}
