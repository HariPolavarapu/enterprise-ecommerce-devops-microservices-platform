package service

import (
	"fmt"
	"time"

	"enterprise/domain"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService { return &NotificationService{} }

func (s *NotificationService) SendOrderNotification(order *domain.Order) string {
	return fmt.Sprintf("order %s notification queued", order.ID)
}

func (s *NotificationService) SendPaymentNotification(order *domain.Order) string {
	return fmt.Sprintf("payment for order %s processed", order.ID)
}

func (s *NotificationService) SendEmailNotification(user *domain.User, subject string) string {
	return fmt.Sprintf("email sent to %s about %s at %s", user.Email, subject, time.Now().Format(time.RFC3339))
}
