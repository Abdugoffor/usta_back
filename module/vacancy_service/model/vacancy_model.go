package vacancy_model

import "time"

type Vacancy struct {
	ID         int64
	Slug       *string
	UserID     *int64
	RegionID   *int64
	DistrictID *int64
	MahallaID  *int64
	Adress     *string
	Name       *string
	Title      *string
	Text       *string
	Contact    *string
	Price      *int64
	ViewsCount *int64
	IsActive   *bool
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time
}

type CategoryVacancy struct {
	ID          int64
	CategoryaID int64
	VacancyID   int64
}
