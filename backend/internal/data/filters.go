package data

import (
	"math"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// sortColumn returns the column name for sorting without the "-" prefix if it exists.
// If the sort parameter is not in the SortSafeList, it panics with an error message.
// This function should only be called after validating user input in the handler.
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// panic here because we should already be validating user input in the handler
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection returns a string representing the SQL sort direction based o pn the provided sort parameter.
// If the sort parameter starts with a '-', it returns "DESC", otherwise it returns "ASC".
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// limit returns the page size as an integer. This value is used in the LIMIT clause of an SQL query.
func (f Filters) limit() int {
	return f.PageSize
}

// offset returns the SQL offset value based on the provided page and page size.
// The calculation (f.Page - 1) * f.PageSize may result in an integer overflow if the validation rule for pageSize is changed.
func (f Filters) offset() int {
	// integer overflow possibility if change pagasize validation rule
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty" example:"1"`
	PageSize     int `json:"page_page,omitempty" example:"20"`
	FirstPage    int `json:"first_page,omitempty" example:"1"`
	LastPage     int `json:"last_page,omitempty" example:"5"`
	TotalRecords int `json:"total_records,omitempty" example:"100"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
