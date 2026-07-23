package service

import (
	"errors"
	"fmt"
	"time"

	"enterprise/domain"
	"enterprise/events"
	"enterprise/repository"
)

type CreateOrderInput struct {
	CustomerID     string
	Items          []OrderItemInput
	IdempotencyKey string
	OrderDate      time.Time
}

type OrderItemInput struct {
	SKU      string
	Name     string
	Quantity int
	UnitPrice float64
}

type OrderService struct {
	orderRepo repository.OrderRepository
	inventory *InventoryService
	publisher events.EventPublisher
}

func NewOrderService(orderRepo repository.OrderRepository, inventory *InventoryService, publisher events.EventPublisher) *OrderService {
	return &OrderService{orderRepo: orderRepo, inventory: inventory, publisher: publisher}
}

func (s *OrderService) CreateOrder(input CreateOrderInput) (*domain.Order, error) {
	if input.CustomerID == "" {
		return nil, errors.New("customer id is required")
	}
	if existing, err := s.orderRepo.FindByIdempotencyKey(input.IdempotencyKey); err == nil && existing != nil {
		return existing, nil
	}
	order := &domain.Order{ID: fmt.Sprintf("order-%d", time.Now().UnixNano()), CustomerID: input.CustomerID, IdempotencyKey: input.IdempotencyKey, Status: domain.OrderStatusCreated, Currency: "USD", OrderDate: input.OrderDate, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	var total float64
	for _, item := range input.Items {
		if _, err := s.inventory.ReserveStock(item.SKU, item.Quantity); err != nil {
			return nil, err
		}
		lineTotal := float64(item.Quantity) * item.UnitPrice
		total += lineTotal
		order.Items = append(order.Items, domain.OrderItem{SKU: item.SKU, Name: item.Name, Quantity: item.Quantity, UnitPrice: item.UnitPrice, TotalPrice: lineTotal})
	}
	order.TotalAmount = total
	if err := s.orderRepo.Save(order); err != nil {
		return nil, err
	}
	_ = s.publisher.Publish("order-created", order)
	return order, nil
}

func (s *OrderService) GetOrderHistory(customerID string) ([]*domain.Order, error) {
	return s.orderRepo.ListByCustomer(customerID)
}

func (s *OrderService) GetOrder(orderID string) (*domain.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

func (s *OrderService) AddPaymentRecord(orderID string, record domain.PaymentRecord) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	order.PaymentHistory = append(order.PaymentHistory, record)
	order.UpdatedAt = time.Now()
	return order, s.orderRepo.Save(order)
}

func (s *OrderService) TrackOrder(orderID string) (string, string, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return "", "pending", err
	}
	if order.TrackingNumber == "" {
		order.TrackingNumber = fmt.Sprintf("TRK-%d", time.Now().UnixNano())
		_ = s.orderRepo.Save(order)
	}
	return order.TrackingNumber, string(order.Status), nil
}

func (s *OrderService) CancelOrder(orderID string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now()
	_ = s.publisher.Publish("order-cancelled", order)
	return order, s.orderRepo.Save(order)
}

func (s *OrderService) UpdateOrderStatus(orderID string, status domain.OrderStatus) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	order.Status = status
	return order, s.orderRepo.Save(order)
}
