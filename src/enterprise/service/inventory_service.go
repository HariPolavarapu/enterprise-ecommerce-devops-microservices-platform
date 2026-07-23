package service

import (
	"errors"
	"fmt"
	"time"

	"enterprise/domain"
	"enterprise/repository"
)

type InventoryInput struct {
	SKU          string
	Name         string
	AvailableQty int
	ReservedQty  int
	ReorderLevel int
}

type InventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CreateInventory(input InventoryInput) (*domain.Inventory, error) {
	if input.SKU == "" || input.Name == "" {
		return nil, errors.New("sku and name are required")
	}
	item := &domain.Inventory{ID: fmt.Sprintf("inv-%d", time.Now().UnixNano()), SKU: input.SKU, Name: input.Name, AvailableQty: input.AvailableQty, ReservedQty: input.ReservedQty, ReorderLevel: input.ReorderLevel}
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *InventoryService) ListInventory() ([]*domain.Inventory, error) {
	return s.repo.List()
}

func (s *InventoryService) GetInventory(sku string) (*domain.Inventory, error) {
	return s.repo.FindBySKU(sku)
}

func (s *InventoryService) ValidateStock(sku string, quantity int) error {
	item, err := s.repo.FindBySKU(sku)
	if err != nil {
		return err
	}
	if quantity < 0 {
		return errors.New("quantity must be positive")
	}
	if item.AvailableQty-item.ReservedQty < quantity {
		return errors.New("insufficient stock")
	}
	return nil
}

func (s *InventoryService) ReserveStock(sku string, quantity int) (*domain.Inventory, error) {
	item, err := s.repo.FindBySKU(sku)
	if err != nil {
		return nil, err
	}
	if quantity < 0 {
		return nil, errors.New("quantity must be positive")
	}
	if item.AvailableQty < quantity {
		return nil, errors.New("insufficient stock")
	}
	item.AvailableQty -= quantity
	item.ReservedQty += quantity
	return item, s.repo.Save(item)
}

func (s *InventoryService) ReleaseStock(sku string, quantity int) (*domain.Inventory, error) {
	item, err := s.repo.FindBySKU(sku)
	if err != nil {
		return nil, err
	}
	if item.ReservedQty < quantity {
		return nil, errors.New("cannot release more than reserved")
	}
	item.AvailableQty += quantity
	item.ReservedQty -= quantity
	return item, s.repo.Save(item)
}

func (s *InventoryService) UpdateStock(sku string, quantity int) (*domain.Inventory, error) {
	item, err := s.repo.FindBySKU(sku)
	if err != nil {
		return nil, err
	}
	item.AvailableQty = quantity
	return item, s.repo.Save(item)
}

func (s *InventoryService) IsLowStock(sku string) bool {
	item, err := s.repo.FindBySKU(sku)
	if err != nil {
		return false
	}
	return item.AvailableQty-item.ReservedQty <= item.ReorderLevel
}

func (s *InventoryService) LowStockItems() ([]*domain.Inventory, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Inventory, 0)
	for _, item := range items {
		if s.IsLowStock(item.SKU) {
			result = append(result, item)
		}
	}
	return result, nil
}
