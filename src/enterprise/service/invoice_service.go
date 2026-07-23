package service

import (
	"fmt"
	"time"

	"enterprise/domain"
)

type InvoiceService struct{}

func NewInvoiceService() *InvoiceService { return &InvoiceService{} }

func (s *InvoiceService) CreateInvoice(order *domain.Order) (*domain.Invoice, error) {
	invoice := &domain.Invoice{ID: fmt.Sprintf("invoice-%d", time.Now().UnixNano()), OrderID: order.ID, InvoiceNumber: fmt.Sprintf("INV-%d", time.Now().UnixNano()), Amount: order.TotalAmount, Currency: order.Currency, Status: "issued", Metadata: map[string]string{"customer_id": order.CustomerID}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return invoice, nil
}
