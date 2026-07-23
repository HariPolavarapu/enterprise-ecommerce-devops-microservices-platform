package service

import (
	"errors"
	"fmt"
	"time"

	"enterprise/domain"
	"enterprise/repository"
)

type WishlistService struct {
	repo repository.WishlistRepository
}

func NewWishlistService(repo repository.WishlistRepository) *WishlistService {
	return &WishlistService{repo: repo}
}

func (s *WishlistService) AddProduct(customerID, sku, name string) (*domain.Wishlist, error) {
	if customerID == "" || sku == "" || name == "" {
		return nil, errors.New("customer id, sku and name are required")
	}
	wishlist, err := s.repo.FindByCustomer(customerID)
	if err != nil {
		wishlist = &domain.Wishlist{CustomerID: customerID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	}
	wishlist.Items = append(wishlist.Items, domain.WishlistItem{ID: fmt.Sprintf("wli-%d", time.Now().UnixNano()), WishlistID: wishlist.ID, SKU: sku, Name: name, AddedAt: time.Now()})
	return wishlist, s.repo.Save(wishlist)
}

func (s *WishlistService) RemoveProduct(customerID, sku string) (*domain.Wishlist, error) {
	wishlist, err := s.repo.FindByCustomer(customerID)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.WishlistItem, 0, len(wishlist.Items))
	for _, item := range wishlist.Items {
		if item.SKU != sku {
			filtered = append(filtered, item)
		}
	}
	wishlist.Items = filtered
	return wishlist, s.repo.Save(wishlist)
}

func (s *WishlistService) GetWishlist(customerID string) (*domain.Wishlist, error) {
	return s.repo.FindByCustomer(customerID)
}
