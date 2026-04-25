package user_model

import "time"

type User struct {
	ID        int64      `json:"id"`
	FullName  string     `json:"full_name"`
	Photo     *string    `json:"photo,omitempty"`
	Phone     string     `json:"phone"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
