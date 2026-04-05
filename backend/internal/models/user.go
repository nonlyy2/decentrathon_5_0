package models

import "time"

// допустимые роли
var ValidRoles = []string{"superadmin", "tech-admin", "auditor", "manager", "admin", "committee"}

type User struct {
	ID           int        `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     *string    `json:"full_name"`
	Role         string     `json:"role"`
	AvatarURL    *string    `json:"avatar_url"`
	CreatedAt    time.Time  `json:"created_at"`
}

type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required,oneof=superadmin tech-admin auditor manager admin committee"`
	FullName *string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty"`
}

type UpdateUserRequest struct {
	Role     *string `json:"role" binding:"omitempty,oneof=superadmin tech-admin auditor manager admin committee"`
	FullName *string `json:"full_name"`
	Email    *string `json:"email" binding:"omitempty,email"`
}
