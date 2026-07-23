package service

import (
	"errors"
	"fmt"

	"enterprise/domain"
)

type CouponService struct{}

func NewCouponService() *CouponService { return &CouponService{} }

func (s *CouponService) ValidateCoupon(code string, orderTotal float64) (*domain.Coupon, error) {
	if code == "" {
		return nil, errors.New("coupon code is required")
	}
	coupon := &domain.Coupon{ID: fmt.Sprintf("coupon-%s", code), Code: code, DiscountType: "percentage", DiscountValue: 10, MinAmount: 20, IsActive: true}
	if !coupon.IsActive {
		return nil, errors.New("coupon is inactive")
	}
	if orderTotal < coupon.MinAmount {
		return nil, errors.New("order total does not meet minimum")
	}
	return coupon, nil
}

func (s *CouponService) CalculateDiscount(orderTotal float64, coupon *domain.Coupon) float64 {
	if coupon == nil || !coupon.IsActive {
		return 0
	}
	if coupon.DiscountType == "percentage" {
		return orderTotal * coupon.DiscountValue / 100
	}
	return coupon.DiscountValue
}
