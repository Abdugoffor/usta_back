package user_model

import "time"

type User struct {
	ID               int64      `json:"id"`
	FullName         string     `json:"full_name"`
	Photo            *string    `json:"photo,omitempty"`
	Phone            string     `json:"phone"`
	TelegramID       *int64     `json:"telegram_id,omitempty"`
	TelegramUsername *string    `json:"telegram_username,omitempty"`
	Role             string     `json:"role"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}
