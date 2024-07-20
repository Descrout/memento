package main

import (
	"fmt"
	"strconv"
	"strings"
)

func SanitiseCommand(content string, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(content, prefix))
}

func SanitiseMovieName(movieName string) string {
	return strings.ToLower(strings.TrimSpace(movieName))
}

func ValidateScore(scoreStr string) (float64, error) {
	parts := strings.Split(scoreStr, ".")
	if len(parts) == 2 && len(parts[1]) > 1 {
		return 0, fmt.Errorf("Score must be integer(eg 3) or 1 scale decimal(eg 7.5)")
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid score format: %s", err.Error())
	}

	if score < 0 || score > 10 {
		return 0, fmt.Errorf("Score must be between 0 and 10.")
	}

	return score, nil
}
