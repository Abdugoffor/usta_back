package resume_dto

import "time"

// ─── Request DTOs ────────────────────────────────────────────────────────────

type CreateResumeRequest struct {
	RegionID       *int64  `json:"region_id"       validate:"omitempty,min=1"`
	DistrictID     *int64  `json:"district_id"     validate:"omitempty,min=1"`
	MahallaID      *int64  `json:"mahalla_id"      validate:"omitempty,min=1"`
	Adress         string  `json:"adress"          validate:"required,min=3,max=500"`
	Name           string  `json:"name"            validate:"required,min=2,max=255"`
	Photo          string  `json:"photo"           validate:"required"`
	Title          string  `json:"title"           validate:"required,min=2,max=500"`
	Text           string  `json:"text"            validate:"required,min=10"`
	Contact        string  `json:"contact"         validate:"required,min=5,max=255"`
	Price          *int64  `json:"price"`
	ExperienceYear *int    `json:"experience_year"`
	Skills         string  `json:"skills"          validate:"required,min=2"`
	IsActive       *bool   `json:"is_active"`
	CategoryIDs    []int64 `json:"category_ids"    validate:"omitempty,dive,min=1"`
}

type UpdateResumeRequest struct {
	RegionID       *int64  `json:"region_id"       validate:"omitempty,min=1"`
	DistrictID     *int64  `json:"district_id"     validate:"omitempty,min=1"`
	MahallaID      *int64  `json:"mahalla_id"      validate:"omitempty,min=1"`
	Adress         *string `json:"adress"          validate:"omitempty,min=3,max=500"`
	Name           *string `json:"name"            validate:"omitempty,min=2,max=255"`
	Photo          *string `json:"photo"           validate:"omitempty"`
	Title          *string `json:"title"           validate:"omitempty,min=2,max=500"`
	Text           *string `json:"text"            validate:"omitempty,min=10"`
	Contact        *string `json:"contact"         validate:"omitempty,min=5,max=255"`
	Price          *int64  `json:"price"`
	ExperienceYear *int    `json:"experience_year"`
	Skills         *string `json:"skills"          validate:"omitempty,min=2"`
	IsActive       *bool   `json:"is_active"`
	CategoryIDs    []int64 `json:"category_ids"    validate:"omitempty,dive,min=1"`
}

// ─── Filter ──────────────────────────────────────────────────────────────────

type ResumeFilter struct {
	UserID        *int64
	RegionID      *int64
	DistrictID    *int64
	MahallaID     *int64
	Name          string
	Title         string
	Search        string
	IsActive      *bool
	MinPrice      *int64
	MaxPrice      *int64
	CategoryID    *int64
	CategoryIDs   []int64
	MinExperience *int
	SortBy        string
	SortOrder     string
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

type CategoryShort struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type ResumeResponse struct {
	ID             int64           `json:"id"`
	Slug           *string         `json:"slug"`
	UserID         *int64          `json:"user_id"`
	RegionID       *int64          `json:"region_id"`
	RegionName     string          `json:"region_name"`
	DistrictID     *int64          `json:"district_id"`
	DistrictName   string          `json:"district_name"`
	MahallaID      *int64          `json:"mahalla_id"`
	MahallaName    string          `json:"mahalla_name"`
	Adress         *string         `json:"adress"`
	Name           *string         `json:"name"`
	Photo          *string         `json:"photo"`
	Title          *string         `json:"title"`
	Text           *string         `json:"text"`
	Contact        *string         `json:"contact"`
	Price          *int64          `json:"price"`
	ExperienceYear *int            `json:"experience_year"`
	Skills         *string         `json:"skills"`
	ViewsCount     *int64          `json:"views_count"`
	IsActive       *bool           `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
	Categories     []CategoryShort `json:"categories"`
}
