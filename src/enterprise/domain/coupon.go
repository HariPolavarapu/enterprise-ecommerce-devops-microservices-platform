package domain

type Coupon struct {
	ID            string
	Code          string
	DiscountType  string
	DiscountValue float64
	MinAmount     float64
	IsActive      bool
}
