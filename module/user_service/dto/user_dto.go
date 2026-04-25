package user_dto

import (
	user_model "main_service/module/user_service/model"
	"time"
)

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Phone    string `json:"phone"     validate:"required,min=9,max=20"`
	Password string `json:"password"  validate:"required,min=6"`
	Role     string `json:"role"      validate:"omitempty,oneof=user employer"`
}

type LoginRequest struct {
	Phone    string `json:"phone"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string          `json:"token"`
	User  user_model.User `json:"user"`
}

type CreateUserRequest struct {
	FullName string  `json:"full_name" validate:"required,min=2,max=255"`
	Phone    string  `json:"phone"     validate:"required,min=9,max=20"`
	Photo    *string `json:"photo"     validate:"omitempty,max=255"`
	Password string  `json:"password"  validate:"required,min=6"`
	Role     string  `json:"role"      validate:"omitempty,oneof=user employer admin"`
	IsActive *bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	FullName *string `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone    *string `json:"phone"     validate:"omitempty,min=9,max=20"`
	Photo    *string `json:"photo"     validate:"omitempty,max=255"`
	Password *string `json:"password"  validate:"omitempty,min=6"`
	Role     *string `json:"role"      validate:"omitempty,oneof=user employer admin"`
	IsActive *bool   `json:"is_active"`
}

type UserFilter struct {
	FullName string
	Phone    string
	Role     string
	IsActive *bool
}

type UserResponse struct {
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
