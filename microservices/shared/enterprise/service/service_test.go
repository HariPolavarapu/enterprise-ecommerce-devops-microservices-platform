package service

import (
	"testing"
	"time"

	"enterprise/domain"
	"enterprise/events"
	"enterprise/repository"
)

func TestUserServiceCreateAndAddress(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	service := NewUserService(userRepo)

	user, err := service.CreateUser(CreateUserInput{
		Email:       "jane@example.com",
		FirstName:   "Jane",
		LastName:    "Doe",
		PhoneNumber: "+123456789",
		Role:        domain.RoleCustomer,
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	_, err = service.AddAddress(user.ID, AddressInput{
		Line1:       "10 Market St",
		City:        "Seattle",
		State:       "WA",
		PostalCode:  "98101",
		Country:     "US",
		AddressType: domain.AddressTypeShipping,
		IsDefault:   true,
	})
	if err != nil {
		t.Fatalf("AddAddress failed: %v", err)
	}

	loaded, err := service.GetUser(user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if len(loaded.Addresses) != 1 {
		t.Fatalf("expected one address, got %d", len(loaded.Addresses))
	}
}

func TestInventoryAndOrderFlow(t *testing.T) {
	inventoryRepo := repository.NewInMemoryInventoryRepository()
	orderRepo := repository.NewInMemoryOrderRepository()
	inventoryService := NewInventoryService(inventoryRepo)
	publisher := events.NewInMemoryEventPublisher()
	orderService := NewOrderService(orderRepo, inventoryService, publisher)

	_, err := inventoryService.CreateInventory(InventoryInput{
		SKU:          "SKU-100",
		Name:         "Wireless Mouse",
		AvailableQty: 10,
		ReservedQty:  0,
		ReorderLevel: 2,
	})
	if err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	reserved, err := inventoryService.ReserveStock("SKU-100", 3)
	if err != nil {
		t.Fatalf("ReserveStock failed: %v", err)
	}
	if reserved.AvailableQty != 7 {
		t.Fatalf("expected available stock 7, got %d", reserved.AvailableQty)
	}

	order, err := orderService.CreateOrder(CreateOrderInput{
		CustomerID: "cust-1",
		Items: []OrderItemInput{{
			SKU:      "SKU-100",
			Name:     "Wireless Mouse",
			Quantity: 2,
			UnitPrice: 30,
		}},
		IdempotencyKey: "order-1",
		OrderDate:      time.Now(),
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.Status != domain.OrderStatusCreated {
		t.Fatalf("expected created status, got %s", order.Status)
	}
}
