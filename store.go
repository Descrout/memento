package main

import (
	"encoding/json"
	"fmt"
	"math"
	"memento/models"
	"sort"
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
	type MovieAverage struct {
		Name    string
		Average float64
	}

	var movieAverages []MovieAverage

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

			movieAverages = append(movieAverages, MovieAverage{
				Name:    string(k),
				Average: avg,
			})

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

	// Sort movieAverages by average score in descending order
	sort.Slice(movieAverages, func(i, j int) bool {
		return movieAverages[i].Average > movieAverages[j].Average
	})

	// Separate sorted movies and averages
	movies := make([]string, len(movieAverages))
	averages := make([]float64, len(movieAverages))

	for i, ma := range movieAverages {
		movies[i] = ma.Name
		averages[i] = ma.Average
	}

	return movies, averages, nil
}

func (s *Store) SearchMovies(search string) ([]string, error) {
	movies := []string{}
	search = strings.ToLower(strings.TrimSpace(search))

	err := s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)

		err := moviesBucket.ForEach(func(k, v []byte) error {
			mv := strings.TrimSpace(string(k))

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

	sort.Slice(reviews, func(i, j int) bool {
		return reviews[i].Score > reviews[j].Score
	})

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
	type ReviewWithMovie struct {
		Review    *models.Review
		MovieName string
	}

	var reviewsWithMovies []ReviewWithMovie

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
				return err
			}

			reviewsWithMovies = append(reviewsWithMovies, ReviewWithMovie{
				Review:    &review,
				MovieName: string(k),
			})

			return nil
		})
	})

	if err != nil {
		return nil, nil, err
	}

	// Sort by review score in descending order
	sort.Slice(reviewsWithMovies, func(i, j int) bool {
		return reviewsWithMovies[i].Review.Score > reviewsWithMovies[j].Review.Score
	})

	// Separate sorted reviews and movie names
	reviews := make([]*models.Review, len(reviewsWithMovies))
	movieNames := make([]string, len(reviewsWithMovies))

	for i, rm := range reviewsWithMovies {
		reviews[i] = rm.Review
		movieNames[i] = rm.MovieName
	}

	return reviews, movieNames, nil
}

func (s *Store) GetUserSimilarities(userID string) ([]*models.UserSimilarity, error) {
	// First, get the current user's reviews
	currentUserReviews, currentUserMovies, err := s.GetReviewsByUser(userID)
	if err != nil {
		return nil, err
	}

	// If the current user has no reviews, return empty list
	if len(currentUserReviews) == 0 {
		return []*models.UserSimilarity{}, nil
	}

	// Map to store other users' reviews
	userReviewMap := make(map[string][]*models.UserMovieReview)

	// Collect reviews from all users
	err = s.db.View(func(tx *bolt.Tx) error {
		moviesBucket := tx.Bucket(s.moviesBucketKey)
		return moviesBucket.ForEach(func(k, v []byte) error {
			movieBucket := moviesBucket.Bucket(k)
			movieName := string(k)

			return movieBucket.ForEach(func(userKey, reviewData []byte) error {
				// Skip the current user
				if string(userKey) == userID {
					return nil
				}

				var review *models.Review
				if err := json.Unmarshal(reviewData, &review); err != nil {
					return err
				}

				userReviewMap[string(userKey)] = append(userReviewMap[string(userKey)], &models.UserMovieReview{
					MovieName: movieName,
					Score:     review.Score,
				})
				return nil
			})
		})
	})

	if err != nil {
		return nil, err
	}

	// Calculate similarities
	similarities := []*models.UserSimilarity{}
	for otherUserID, otherUserReviews := range userReviewMap {
		similarity := calculateImprovedSimilarity(currentUserReviews, currentUserMovies, otherUserReviews)
		if similarity > 0 {
			similarities = append(similarities, &models.UserSimilarity{
				UserID:     otherUserID,
				Similarity: similarity,
			})
		}
	}

	// Sort similarities in descending order
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	return similarities, nil
}

// calculateImprovedSimilarity provides a more nuanced similarity calculation
func calculateImprovedSimilarity(currentUserReviews []*models.Review, currentUserMovies []string, otherUserReviews []*models.UserMovieReview) float64 {
	// Create a map of current user's reviews by movie
	currentUserReviewMap := make(map[string]float64)
	for i, movie := range currentUserMovies {
		currentUserReviewMap[movie] = currentUserReviews[i].Score
	}

	// Prepare vectors for common movies
	var currentUserVector, otherUserVector []float64

	// Track common movies
	for _, otherReview := range otherUserReviews {
		currentUserScore, exists := currentUserReviewMap[otherReview.MovieName]
		if exists {
			currentUserVector = append(currentUserVector, currentUserScore)
			otherUserVector = append(otherUserVector, otherReview.Score)
		}
	}

	// If fewer than 2 common movies, return 0
	if len(currentUserVector) < 2 {
		return 0
	}

	// Calculate Pearson correlation coefficient
	return pearsonCorrelation(currentUserVector, otherUserVector)
}

// pearsonCorrelation calculates the Pearson correlation coefficient
func pearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	// Calculate means
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(len(x))
	meanY := sumY / float64(len(y))

	// Calculate covariance and standard deviations
	var covariance, varX, varY float64
	for i := range x {
		diffX := x[i] - meanX
		diffY := y[i] - meanY
		covariance += diffX * diffY
		varX += diffX * diffX
		varY += diffY * diffY
	}

	// Avoid division by zero
	if varX == 0 || varY == 0 {
		return 0
	}

	// Calculate Pearson correlation
	correlation := covariance / (math.Sqrt(varX) * math.Sqrt(varY))

	// Ensure correlation is between -1 and 1
	if math.IsNaN(correlation) {
		return 0
	}

	return correlation
}
