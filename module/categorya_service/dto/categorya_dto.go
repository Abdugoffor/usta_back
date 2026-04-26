package categorya_dto

import "time"

type CreateCategoryRequest struct {
	Name     map[string]string `json:"name" validate:"required"`
	IsActive *bool             `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name     *map[string]string `json:"name"`
	IsActive *bool              `json:"is_active"`
}

type CategoryFilter struct {
	Name     string
	IsActive *bool
}

type CategoryResponse struct {
	ID        int64             `json:"id"`
	Name      map[string]string `json:"name"`
	IsActive  bool              `json:"is_active"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

type CategoryActiveResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
