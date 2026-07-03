package service

import (
	"testing"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGetInitials(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", "NA"},
		{"single char", "A", "A"},
		{"two chars", "AB", "AB"},
		{"full name", "John Doe", "JD"},
		{"indonesian name", "Budi Santoso", "BS"},
		{"single rune", "\u00c3", "\u00c3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitials(tt.input)
			if got != tt.want {
				t.Errorf("getInitials() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetStringClaim(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id":   "123",
		"email":     "test@test.com",
		"role_slug": "super_admin",
	}

	tests := []struct {
		name   string
		claims jwt.MapClaims
		key    string
		want   string
	}{
		{"existing key", claims, "email", "test@test.com"},
		{"existing key 2", claims, "role_slug", "super_admin"},
		{"missing key", claims, "nonexistent", ""},
		{"nil claims", nil, "key", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringClaim(tt.claims, tt.key)
			if got != tt.want {
				t.Errorf("getStringClaim() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuthService_GenerateRefreshToken(t *testing.T) {
	cfg := &config.Config{}
	service := NewAuthService(cfg)

	token, err := service.generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken() error = %v", err)
	}
	if len(token) == 0 {
		t.Error("generateRefreshToken() returned empty token")
	}
	// 32 bytes = 64 hex chars
	if len(token) != 64 {
		t.Errorf("generateRefreshToken() length = %d, want 64", len(token))
	}
}

func TestAuthService_GenerateAccessToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:       "test-secret-key-12345",
		JWTAccessExpiry: 3600000000000, // 1 hour in nanoseconds
	}
	service := NewAuthService(cfg)

	employee := &models.Employee{
		ID:         uuid.New(),
		EmployeeID: "EMP001",
		Email:      "test@test.com",
		RoleSlug:   "super_admin",
	}

	token, expiresIn, err := service.generateAccessToken(employee)
	if err != nil {
		t.Fatalf("generateAccessToken() error = %v", err)
	}
	if len(token) == 0 {
		t.Error("generateAccessToken() returned empty token")
	}
	if expiresIn <= 0 {
		t.Errorf("expiresIn = %d, want > 0", expiresIn)
	}
}

func TestValidateAccessToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:       "test-secret-key-12345",
		JWTAccessExpiry: 3600000000000,
	}
	service := NewAuthService(cfg)

	employee := &models.Employee{
		ID:         uuid.New(),
		EmployeeID: "EMP001",
		Email:      "test@test.com",
		RoleSlug:   "super_admin",
	}

	token, _, err := service.generateAccessToken(employee)
	if err != nil {
		t.Fatalf("generateAccessToken() error = %v", err)
	}

	// Test valid token
	claims, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != employee.ID {
		t.Errorf("UserID = %v, want %v", claims.UserID, employee.ID)
	}
	if claims.Email != employee.Email {
		t.Errorf("Email = %s, want %s", claims.Email, employee.Email)
	}
	if claims.RoleSlug != employee.RoleSlug {
		t.Errorf("RoleSlug = %s, want %s", claims.RoleSlug, employee.RoleSlug)
	}

	// Test invalid token
	_, err = service.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("ValidateAccessToken() with invalid token should return error")
	}
}
