package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"enterprise/domain"
)

type ReviewRepository interface {
	Save(review *domain.Review) error
	FindByProduct(productSKU string) ([]*domain.Review, error)
}

type InMemoryReviewRepository struct {
	mu      sync.RWMutex
	reviews []domain.Review
}

func NewInMemoryReviewRepository() *InMemoryReviewRepository {
	return &InMemoryReviewRepository{}
}

func (r *InMemoryReviewRepository) Save(review *domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if review.ID == "" {
		review.ID = fmt.Sprintf("review-%d", time.Now().UnixNano())
	}
	if review.CreatedAt.IsZero() {
		review.CreatedAt = time.Now()
	}
	review.UpdatedAt = time.Now()
	r.reviews = append(r.reviews, *review)
	return nil
}

func (r *InMemoryReviewRepository) FindByProduct(productSKU string) ([]*domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Review, 0)
	for i := range r.reviews {
		if r.reviews[i].ProductSKU == productSKU && !r.reviews[i].IsDeleted {
			copy := r.reviews[i]
			result = append(result, &copy)
		}
	}
	return result, nil
}
