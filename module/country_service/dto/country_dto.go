package country_dto

import "time"

type CreateCountryRequest struct {
	ParentID *int64            `json:"parent_id" validate:"omitempty,min=1"`
	Name     map[string]string `json:"name"      validate:"required"`
	IsActive *bool             `json:"is_active"`
}

type UpdateCountryRequest struct {
	ParentID *int64             `json:"parent_id" validate:"omitempty,min=1"`
	Name     *map[string]string `json:"name"`
	IsActive *bool              `json:"is_active"`
}

type CountryFilter struct {
	ParentID  *int64
	HasParent *bool
	Name      string
	IsActive  *bool
}

type CountryResponse struct {
	ID        int64             `json:"id"`
	ParentID  *int64            `json:"parent_id"`
	Name      map[string]string `json:"name"`
	IsActive  bool              `json:"is_active"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

type CountryActiveResponse struct {
	ID       int64  `json:"id"`
	ParentID *int64 `json:"parent_id"`
	Name     string `json:"name"`
}
