package country_model

import "time"

type Country struct {
	ID        int64
	ParentID  *int64
	Name      *string
	IsActive  *bool
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}
