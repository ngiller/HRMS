package models

import (
	"time"

	"github.com/google/uuid"
)

// Login Request/Response
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID             uuid.UUID                  `json:"id"`
	EmployeeID     string                     `json:"employee_id"`
	FullName       string                     `json:"full_name"`
	Email          string                     `json:"email"`
	RoleID         uuid.UUID                  `json:"role_id"`
	RoleSlug       string                     `json:"role_slug"`
	RoleName       string                     `json:"role_name"`
	PositionName   string                     `json:"position_name"`
	DepartmentName string                     `json:"department_name"`
	AvatarInitials string                     `json:"avatar_initials"`
	Permissions    map[string]map[string]bool `json:"permissions"`
}

// Forgot Password Request/Response
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// Reset Password Request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// Change Password Request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// Token Claims
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	EmployeeID string    `json:"employee_id"`
	Email      string    `json:"email"`
	RoleSlug   string    `json:"role_slug"`
}

// Session
type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	IsActive     bool      `json:"is_active"`
	ExpiresAt    time.Time `json:"expires_at"`
}
