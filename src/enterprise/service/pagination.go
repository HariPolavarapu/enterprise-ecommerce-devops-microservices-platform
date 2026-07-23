package service

import "enterprise/domain"

func Paginate[T any](items []T, page, pageSize int) domain.PageResult[T] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return domain.PageResult[T]{
		Items:      items[start:end],
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: (total + pageSize - 1) / pageSize,
	}
}
