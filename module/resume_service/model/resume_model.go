package resume_model

import "time"

type Resume struct {
	ID             int64
	Slug           *string
	UserID         *int64
	RegionID       *int64
	DistrictID     *int64
	MahallaID      *int64
	Adress         *string
	Name           *string
	Photo          *string
	Title          *string
	Text           *string
	Contact        *string
	Price          *int64
	ExperienceYear *int
	Skills         *string
	ViewsCount     *int64
	IsActive       *bool
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

type CategoryResume struct {
	ID          int64
	CategoryaID int64
	ResumeID    int64
}
