package domain

import "time"

type Wishlist struct {
	ID         string
	CustomerID string
	Items      []WishlistItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WishlistItem struct {
	ID         string
	WishlistID string
	SKU        string
	Name       string
	AddedAt   time.Time
}
