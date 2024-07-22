package main

import (
	"fmt"
	"strconv"
	"strings"

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

func (s *Store) AddMovie(movie string) (bool, error) {
	movie = strings.ToLower(movie) // Ensure movie name is in lowercase
	alreadyExists := false

	err := s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket, err := tx.CreateBucketIfNotExists([]byte("Movies"))
		if err != nil {
			return err
		}

		movieBucket := moviesBucket.Bucket([]byte(movie))
		if movieBucket != nil {
			alreadyExists = true // The movie bucket already exists
			return nil
		}

		_, err = moviesBucket.CreateBucket([]byte(movie)) // Create new bucket if not exists
		return err
	})

	return alreadyExists, err
}

func (s *Store) AddScore(movie string, score float64, userID string) error {
	movie = strings.ToLower(movie) // Convert movie name to lowercase
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket([]byte("Movies"))
		if moviesBucket == nil {
			return fmt.Errorf("movies bucket does not exist")
		}
		// Create or retrieve the movie bucket
		movieBucket, err := moviesBucket.CreateBucketIfNotExists([]byte(movie))
		if err != nil {
			return fmt.Errorf("failed to create or get movie bucket for '%s': %s", movie, err)
		}
		scoreString := fmt.Sprintf("%.2f", score)
		return movieBucket.Put([]byte(userID), []byte(scoreString))
	})
}

func (s *Store) GetScores(movie string) (map[string]float64, float64, error) {
	movie = strings.ToLower(movie) // Convert movie name to lowercase
	scores := make(map[string]float64)
	var totalScore float64
	var count float64

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket([]byte("Movies"))
		if moviesBucket == nil {
			return fmt.Errorf("movies bucket does not exist")
		}

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

func (s *Store) ListMovies() ([]string, error) {
	var movies []string
	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket([]byte("Movies"))
		if moviesBucket == nil {
			return fmt.Errorf("movies bucket does not exist")
		}

		return moviesBucket.ForEach(func(k, v []byte) error {
			movies = append(movies, string(k))
			return nil
		})
	})

	return movies, err
}

func (s *Store) ClearAllData() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Iterate over all bucket names and delete them
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}
