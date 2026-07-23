package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"enterprise/domain"
)

type WishlistRepository interface {
	Save(wishlist *domain.Wishlist) error
	FindByCustomer(customerID string) (*domain.Wishlist, error)
}

type InMemoryWishlistRepository struct {
	mu         sync.RWMutex
	wishlists  map[string]*domain.Wishlist
}

func NewInMemoryWishlistRepository() *InMemoryWishlistRepository {
	return &InMemoryWishlistRepository{wishlists: make(map[string]*domain.Wishlist)}
}

func (r *InMemoryWishlistRepository) Save(wishlist *domain.Wishlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if wishlist.ID == "" {
		wishlist.ID = fmt.Sprintf("wishlist-%d", time.Now().UnixNano())
	}
	wishlist.UpdatedAt = time.Now()
	r.wishlists[wishlist.CustomerID] = wishlist
	return nil
}

func (r *InMemoryWishlistRepository) FindByCustomer(customerID string) (*domain.Wishlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wishlist, ok := r.wishlists[customerID]
	if !ok {
		return nil, errors.New("wishlist not found")
	}
	return wishlist, nil
}
