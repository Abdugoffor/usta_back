package translation_dto

import "time"

type CreateTranslationRequest struct {
	Slug     string            `json:"slug" validate:"required,min=2,max=150"`
	Name     map[string]string `json:"name" validate:"required"`
	IsActive *bool             `json:"is_active"`
}

type UpdateTranslationRequest struct {
	Slug     *string            `json:"slug" validate:"omitempty,min=2,max=150"`
	Name     *map[string]string `json:"name"`
	IsActive *bool              `json:"is_active"`
}

type TranslationResponse struct {
	ID        int64             `json:"id"`
	Slug      string            `json:"slug"`
	Name      map[string]string `json:"name"`
	IsActive  bool              `json:"is_active"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type TranslationFilter struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	IsActive *bool  `json:"is_active"`
}

// TranslationKeyResponse — frontend uchun bitta kalit bo'yicha tarjima.
type TranslationKeyResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Lang  string `json:"lang"`
}
