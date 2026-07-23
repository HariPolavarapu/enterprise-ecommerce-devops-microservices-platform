package domain

type Inventory struct {
	ID            string
	SKU           string
	Name          string
	AvailableQty  int
	ReservedQty   int
	ReorderLevel  int
	IsDeleted     bool
	Version       int
}

func (i Inventory) AvailableAfterReservation() int {
	return i.AvailableQty - i.ReservedQty
}
