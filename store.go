package main

import (
	"encoding/json"
	"fmt"
	"memento/models"
	"strings"

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

func (s *Store) AddReview(movieName string, review *models.Review) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket, err := moviesBucket.CreateBucketIfNotExists([]byte(movieName))
		if err != nil {
			return fmt.Errorf("failed to create or get movie bucket for '%s': %s", movieName, err)
		}

		rawReview, err := json.Marshal(review)
		if err != nil {
			return err
		}

		return movieBucket.Put([]byte(review.AuthorID), rawReview)
	})
}

func (s *Store) GetMovies() ([]string, []float64, error) {
	movies := []string{}
	averages := []float64{}
	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		err := moviesBucket.ForEach(func(k, v []byte) error {
			movieBucket := moviesBucket.Bucket(k)
			totalScore := float64(0)
			count := 0
			err := movieBucket.ForEach(func(k, v []byte) error {
				review := &models.Review{}
				err := json.Unmarshal(v, review)
				if err != nil {
					return err
				}

				totalScore += review.Score
				count++
				return nil
			})
			if err != nil {
				return err
			}
			var avg float64 = 0
			if count > 0 {
				avg = totalScore / float64(count)
			}
			averages = append(averages, avg)
			movies = append(movies, string(k))

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return movies, averages, nil
}

func (s *Store) SearchMovies(search string) ([]string, error) {
	movies := []string{}
	search = strings.ToLower(strings.TrimSpace(search))

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		err := moviesBucket.ForEach(func(k, v []byte) error {
			mv := string(k)

			if strings.Contains(strings.ToLower(mv), search) && len(movies) < 8 {
				movies = append(movies, mv)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *Store) GetReviews(movie string) ([]*models.Review, float64, error) {
	var totalScore float64
	reviews := []*models.Review{}

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket := moviesBucket.Bucket([]byte(movie))

		err := movieBucket.ForEach(func(k, v []byte) error {
			review := &models.Review{}
			err := json.Unmarshal(v, review)
			if err != nil {
				return err
			}
			reviews = append(reviews, review)

			totalScore += review.Score

			return nil
		})
		return err
	})

	if err != nil {
		return nil, 0, err
	}

	average := float64(0)
	count := float64(len(reviews))
	if count > 0 {
		average = totalScore / count
	}

	return reviews, average, nil
}

func (s *Store) GetReviewCount(movie string) (int, error) {
	var count int = 0

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)
		movieBucket := moviesBucket.Bucket([]byte(movie))
		err := movieBucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
		return err
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) DeleteMovie(movie string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket := moviesBucket.Bucket([]byte(movie))
		if movieBucket == nil {
			return fmt.Errorf("this movie did not exists")
		}

		// Delete the movie bucket
		if err := moviesBucket.DeleteBucket([]byte(movie)); err != nil {
			return err
		}

		return nil
	})
}

func (s *Store) DeleteReview(movie string, authorID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		movieBucket := moviesBucket.Bucket([]byte(movie))
		if movieBucket == nil {
			return fmt.Errorf("this movie did not exists")
		}

		deleted := false

		err := movieBucket.ForEach(func(k, v []byte) error {
			if string(k) == authorID {
				deleted = true
			}
			return nil
		})
		if err != nil {
			return nil
		}

		if !deleted {
			return fmt.Errorf("you did not review this movie before")
		}

		return movieBucket.Delete([]byte(authorID))

	})
}

func (s *Store) ClearAllData() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Iterate over all bucket names and delete them
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}

func (s *Store) GetReviewsByUser(userID string) ([]*models.Review, []string, error) {
	reviews := []*models.Review{}
	movieNames := []string{}

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)
		return moviesBucket.ForEach(func(k, v []byte) error {
			movieBucket := moviesBucket.Bucket(k)
			if movieBucket == nil {
				return nil
			}

			reviewValue := movieBucket.Get([]byte(userID))
			if reviewValue == nil {
				return nil
			}

			var review models.Review
			if err := json.Unmarshal(reviewValue, &review); err != nil {
				return nil
			}

			reviews = append(reviews, &review)
			movieNames = append(movieNames, string(k))

			return nil
		})
	})

	if err != nil {
		return nil, nil, err
	}

	return reviews, movieNames, nil
}

// func (s *Store) GetMovieNameByReview(review *Review) (string, error) {
// 	var movieName string

// 	err := s.db.View(func(tx *bolt.Tx) error {
// 		moviesBucket := tx.Bucket(s.moviesBucketKey)

// 		return moviesBucket.ForEach(func(k, v []byte) error {
// 			movieBucket := moviesBucket.Bucket(k)
// 			if movieBucket == nil {
// 				return nil
// 			}

// 			return movieBucket.ForEach(func(reviewKey, reviewValue []byte) error {
// 				var r Review
// 				if err := json.Unmarshal(reviewValue, &r); err != nil {
// 					return err
// 				}

// 				if r.AuthorID == review.AuthorID && r.Comment == review.Comment && r.Score == review.Score {
// 					movieName = string(k)
// 					return nil
// 				}
// 				return nil
// 			})
// 		})
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	return movieName, nil
// }
