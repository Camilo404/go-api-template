package models

import "strconv"

// Page is the metadata returned alongside paginated list responses.
type Page struct {
	Page    int `json:"page" example:"1"`
	PerPage int `json:"per_page" example:"20"`
	Total   int `json:"total" example:"42"`
}

// PaginatedResponse wraps a slice with pagination metadata so clients
// always know how to fetch the next page.
type PaginatedResponse struct {
	Data any  `json:"data"`
	Page Page `json:"page"`
}

// ParsePagination parses ?page=&per_page= query string values, applying
// safe defaults and clamping the per-page count to maxPerPage.
func ParsePagination(pageStr, perPageStr string, defaultPerPage, maxPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	perPage, _ = strconv.Atoi(perPageStr)
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}
