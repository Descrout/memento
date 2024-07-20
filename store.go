package main

import (
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

type Store struct {
	db              *bolt.DB
	moviesBucketKey []byte
}

func NewStore() (*Store, error) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	moviesBucketKey := []byte("Movies")

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(moviesBucketKey)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Store{
		db:              db,
		moviesBucketKey: moviesBucketKey,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) AddMovie(movie string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket := moviesBucket.Bucket([]byte(movie))
		if movieBucket != nil {
			return fmt.Errorf("this movie already exists")
		}

		_, err := moviesBucket.CreateBucket([]byte(movie))
		return err
	})
}

func (s *Store) AddScore(movie string, score float64, userID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket, err := moviesBucket.CreateBucketIfNotExists([]byte(movie))
		if err != nil {
			return fmt.Errorf("failed to create or get movie bucket for '%s': %s", movie, err)
		}

		scoreString := fmt.Sprintf("%.1f", score)
		return movieBucket.Put([]byte(userID), []byte(scoreString))
	})
}

func (s *Store) GetScores(movie string) (map[string]float64, float64, error) {
	scores := make(map[string]float64)
	var totalScore float64
	var count float64

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)
		movieBucket := moviesBucket.Bucket([]byte(movie))
		if movieBucket == nil {
			return fmt.Errorf("no scores available for '%s'", movie)
		}

		err := movieBucket.ForEach(func(k, v []byte) error {
			score, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return err
			}
			scores[string(k)] = score
			totalScore += score
			count++
			return nil
		})
		return err
	})

	if err != nil {
		return nil, 0, err
	}

	var average float64
	if count > 0 {
		average = totalScore / count
	}

	return scores, average, nil
}

func (s *Store) ClearAllData() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Iterate over all bucket names and delete them
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}
