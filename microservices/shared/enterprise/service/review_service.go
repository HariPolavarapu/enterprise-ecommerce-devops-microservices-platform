package service

import (
	"errors"
	"fmt"
	"time"

	"enterprise/domain"
	"enterprise/repository"
)

type ReviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) AddReview(customerID, productSKU string, rating int, comment string) (*domain.Review, error) {
	if customerID == "" || productSKU == "" {
		return nil, errors.New("customer id and product sku are required")
	}
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	review := &domain.Review{ID: fmt.Sprintf("review-%d", time.Now().UnixNano()), CustomerID: customerID, ProductSKU: productSKU, Rating: rating, Comment: comment, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.repo.Save(review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *ReviewService) GetReviews(productSKU string) ([]*domain.Review, error) {
	return s.repo.FindByProduct(productSKU)
}

func (s *ReviewService) AverageRating(productSKU string) (float64, error) {
	reviews, err := s.repo.FindByProduct(productSKU)
	if err != nil {
		return 0, err
	}
	if len(reviews) == 0 {
		return 0, nil
	}
	sum := 0
	for _, review := range reviews {
		sum += review.Rating
	}
	return float64(sum) / float64(len(reviews)), nil
}
