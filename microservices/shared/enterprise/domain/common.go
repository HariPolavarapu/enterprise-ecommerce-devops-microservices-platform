package domain

import "time"

type Auditable struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

type SoftDeletable struct {
	IsDeleted bool
	DeletedAt *time.Time
}

type Versioned struct {
	Version int
}

type PageRequest struct {
	Page     int
	PageSize int
	SortBy   string
	SortDesc bool
	Search   string
}

type PageResult[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
}

type FilterSpec struct {
	Status string
	SKU    string
	Role   string
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)
