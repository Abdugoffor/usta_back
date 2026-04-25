package helper

import (
	"math"
	"net/http"
	"strconv"
	"strings"
)

type PageQuery struct {
	Page      int
	Limit     int
	SortCol   string
	SortOrder string
}

func (p PageQuery) Offset() int { return (p.Page - 1) * p.Limit }

type PageMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int64 `json:"total_pages"`
}

func ParsePage(r *http.Request) PageQuery {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	{
		if page < 1 {
			page = 1
		}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	{
		if limit < 1 || limit > 100 {
			limit = 10
		}
	}

	sortOrder := strings.ToUpper(r.URL.Query().Get("sort_order"))
	{
		if sortOrder != "DESC" {
			sortOrder = "ASC"
		}
	}

	return PageQuery{Page: page, Limit: limit, SortCol: r.URL.Query().Get("sort_by"), SortOrder: sortOrder}
}

func NewPageMeta(total int64, page, limit int) PageMeta {
	tp := int64(math.Ceil(float64(total) / float64(limit)))
	{
		if total == 0 {
			tp = 0
		}
	}

	return PageMeta{Total: total, Page: page, Limit: limit, TotalPages: tp}
}
