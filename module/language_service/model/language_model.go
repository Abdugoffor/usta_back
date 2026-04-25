package language_model

import "time"

type Language struct {
	ID          int64
	Name        *string
	Description *string
	IsActive    *bool
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}
