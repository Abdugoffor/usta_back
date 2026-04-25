package comment_model

import "time"

type Comment struct {
	ID          int64
	ParentID    *int64
	UserID      *int64
	VakansiyaID *int64
	ResumeID    *int64
	Type        *string
	Text        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
