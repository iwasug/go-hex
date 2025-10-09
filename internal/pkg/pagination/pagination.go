package pagination

import (
	"math"
)

type Metadata struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalRecords int64 `json:"total_records"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

type PaginatedResponse struct {
	Data     interface{} `json:"data"`
	Metadata Metadata    `json:"metadata"`
}

func NewMetadata(page, pageSize int, totalRecords int64) Metadata {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}

func NewPaginatedResponse(data interface{}, page, pageSize int, totalRecords int64) PaginatedResponse {
	return PaginatedResponse{
		Data:     data,
		Metadata: NewMetadata(page, pageSize, totalRecords),
	}
}

func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

func ValidatePagination(page, pageSize, maxPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
