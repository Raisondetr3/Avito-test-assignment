package service

import (
	"math/rand"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
)

type ReviewerAssigner struct {
}

func NewReviewerAssigner() *ReviewerAssigner {
	return &ReviewerAssigner{}
}

func (ra *ReviewerAssigner) SelectReviewers(candidates []*domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	count := min(maxCount, len(candidates))

	selected := make([]string, 0, count)
	indices := rand.Perm(len(candidates))

	for i := 0; i < count; i++ {
		selected = append(selected, candidates[indices[i]].UserID)
	}

	return selected
}

func (ra *ReviewerAssigner) SelectRandomReviewer(candidates []*domain.User) (string, bool) {
	if len(candidates) == 0 {
		return "", false
	}

	idx := rand.Intn(len(candidates))
	return candidates[idx].UserID, true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
