package api

import (
	"encoding/json"
	"net/http"

	"enterprise/service"
)

type Handler struct {
	UserService         *service.UserService
	InventoryService    *service.InventoryService
	OrderService        *service.OrderService
	WishlistService     *service.WishlistService
	ReviewService       *service.ReviewService
	CouponService       *service.CouponService
	InvoiceService      *service.InvoiceService
	NotificationService *service.NotificationService
	CacheService        *service.CacheService
	Authorization       *service.AuthorizationService
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input service.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.UserService.CreateUser(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(user)
}

func (h *Handler) Inventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input service.InventoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := h.InventoryService.CreateInventory(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(item)
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input service.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.OrderService.CreateOrder(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(order)
}

func (h *Handler) Wishlist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		CustomerID string `json:"customer_id"`
		SKU        string `json:"sku"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	wishlist, err := h.WishlistService.AddProduct(input.CustomerID, input.SKU, input.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(wishlist)
}

func (h *Handler) Reviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		CustomerID string `json:"customer_id"`
		ProductSKU string `json:"product_sku"`
		Rating     int    `json:"rating"`
		Comment    string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	review, err := h.ReviewService.AddReview(input.CustomerID, input.ProductSKU, input.Rating, input.Comment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(review)
}

func (h *Handler) Coupons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Code      string  `json:"code"`
		OrderTotal float64 `json:"order_total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	coupon, err := h.CouponService.ValidateCoupon(input.Code, input.OrderTotal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(coupon)
}
