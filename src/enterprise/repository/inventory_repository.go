package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"enterprise/domain"
)

type InventoryRepository interface {
	Save(item *domain.Inventory) error
	FindBySKU(sku string) (*domain.Inventory, error)
	List() ([]*domain.Inventory, error)
}

type InMemoryInventoryRepository struct {
	mu         sync.RWMutex
	inventory  map[string]*domain.Inventory
}

func NewInMemoryInventoryRepository() *InMemoryInventoryRepository {
	return &InMemoryInventoryRepository{inventory: make(map[string]*domain.Inventory)}
}

func (r *InMemoryInventoryRepository) Save(item *domain.Inventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item.ID == "" {
		item.ID = fmt.Sprintf("inventory-%d", time.Now().UnixNano())
	}
	item.Version++
	r.inventory[item.SKU] = item
	return nil
}

func (r *InMemoryInventoryRepository) FindBySKU(sku string) (*domain.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.inventory[sku]
	if !ok {
		return nil, errors.New("inventory not found")
	}
	return item, nil
}

func (r *InMemoryInventoryRepository) List() ([]*domain.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]*domain.Inventory, 0, len(r.inventory))
	for _, item := range r.inventory {
		items = append(items, item)
	}
	return items, nil
}
