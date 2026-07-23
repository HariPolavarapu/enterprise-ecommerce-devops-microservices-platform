package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"enterprise/domain"
)

type OrderRepository interface {
	Save(order *domain.Order) error
	FindByID(id string) (*domain.Order, error)
	FindByIdempotencyKey(key string) (*domain.Order, error)
	ListByCustomer(customerID string) ([]*domain.Order, error)
}

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{orders: make(map[string]*domain.Order)}
}

func (r *InMemoryOrderRepository) Save(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if order.ID == "" {
		order.ID = fmt.Sprintf("order-%d", time.Now().UnixNano())
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	order.UpdatedAt = time.Now()
	order.Version++
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (r *InMemoryOrderRepository) FindByIdempotencyKey(key string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, order := range r.orders {
		if order.IdempotencyKey == key {
			return order, nil
		}
	}
	return nil, errors.New("order not found")
}

func (r *InMemoryOrderRepository) ListByCustomer(customerID string) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Order, 0)
	for _, order := range r.orders {
		if order.CustomerID == customerID {
			result = append(result, order)
		}
	}
	return result, nil
}
