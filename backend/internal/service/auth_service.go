package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrAccountLocked      = errors.New("akun terkunci, coba lagi nanti")
	ErrInvalidToken       = errors.New("token tidak valid atau sudah kadaluarsa")
	ErrEmailNotFound      = errors.New("email tidak ditemukan")
)

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest, ipAddress string) (*models.LoginResponse, error) {
	// Get employee by email
	employee, err := repository.GetEmployeeByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("terjadi kesalahan sistem")
	}
	if employee == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if employee.IsLocked {
		if employee.LockedUntil != nil && employee.LockedUntil.After(time.Now()) {
			return nil, ErrAccountLocked
		}
	}

	// Check if password_hash exists
	if employee.PasswordHash == "" || employee.PasswordHash == "PLACEHOLDER_HASH_RESET_ON_FIRST_LOGIN" {
		repository.RecordLoginAttempt(ctx, employee.ID, ipAddress, false)
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(employee.PasswordHash), []byte(req.Password))
	if err != nil {
		// Record failed attempt
		repository.RecordLoginAttempt(ctx, employee.ID, ipAddress, false)

		// Check if too many failed attempts (5 within 15 minutes)
		failedCount, _ := repository.GetRecentFailedAttempts(ctx, employee.ID, 15*time.Minute)
		if failedCount >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			repository.LockEmployee(ctx, employee.ID, lockUntil)
		}

		return nil, ErrInvalidCredentials
	}

	// Record successful login
	repository.RecordLoginAttempt(ctx, employee.ID, ipAddress, true)
	repository.UpdateLastLogin(ctx, employee.ID)

	// Generate tokens
	accessToken, expiresIn, err := s.generateAccessToken(employee)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, errors.New("gagal membuat refresh token")
	}

	// Get initials for avatar
	initials := getInitials(employee.FullName)

	roleID := uuid.Nil
	if employee.RoleID != nil {
		roleID = *employee.RoleID
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: models.UserResponse{
			ID:             employee.ID,
			EmployeeID:     employee.EmployeeID,
			FullName:       employee.FullName,
			Email:          employee.Email,
			RoleID:         roleID,
			RoleSlug:       employee.RoleSlug,
			RoleName:       employee.RoleName,
			PositionName:   employee.PositionName,
			DepartmentName: employee.DepartmentName,
			AvatarInitials: initials,
		},
	}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error) {
	// Find employee by email
	employee, err := repository.GetEmployeeByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("terjadi kesalahan sistem")
	}
	if employee == nil {
		// Don't reveal if email exists
		return &models.ForgotPasswordResponse{
			Message: "Jika email terdaftar, Anda akan menerima tautan reset password",
		}, nil
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("gagal membuat token reset")
	}
	token := hex.EncodeToString(tokenBytes)

	// Save token to database
	err = repository.CreatePasswordResetToken(ctx, employee.ID, token, time.Now().Add(s.cfg.ResetTokenExpiry))
	if err != nil {
		return nil, errors.New("gagal menyimpan token reset")
	}

	return &models.ForgotPasswordResponse{
		Message: "Jika email terdaftar, Anda akan menerima tautan reset password",
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req models.ResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	// Validate token
	employeeID, err := repository.ValidateResetToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("terjadi kesalahan sistem")
	}
	if employeeID == nil {
		return nil, ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	// Update password
	err = repository.UpdatePassword(ctx, *employeeID, string(hashedPassword))
	if err != nil {
		return nil, errors.New("gagal mengupdate password")
	}

	// Mark token as used
	repository.MarkResetTokenUsed(ctx, req.Token)

	return &models.ResetPasswordResponse{
		Message: "Password berhasil direset, silakan login dengan password baru",
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.UserResponse, error) {
	// Try to parse as UUID first, then as employee_id
	parsedID, err := uuid.Parse(id)
	if err != nil {
		// Try searching by email
		employee, err := repository.GetEmployeeByEmail(ctx, id)
		if err != nil {
			return nil, errors.New("terjadi kesalahan sistem")
		}
		if employee == nil {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		return employeeToUserResponse(employee), nil
	}

	employee, err := repository.GetEmployeeByID(ctx, parsedID)
	if err != nil {
		return nil, errors.New("terjadi kesalahan sistem")
	}
	if employee == nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	return employeeToUserResponse(employee), nil
}

func (s *AuthService) generateAccessToken(employee *models.Employee) (string, int64, error) {
	claims := jwt.MapClaims{
		"user_id":     employee.ID.String(),
		"employee_id": employee.EmployeeID,
		"email":       employee.Email,
		"role_slug":   employee.RoleSlug,
		"exp":         time.Now().Add(s.cfg.JWTAccessExpiry).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, int64(s.cfg.JWTAccessExpiry.Seconds()), nil
}

func (s *AuthService) generateRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id tidak valid dalam token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("user_id tidak valid dalam token")
	}

	return &models.Claims{
		UserID:     userID,
		EmployeeID: getStringClaim(claims, "employee_id"),
		Email:      getStringClaim(claims, "email"),
		RoleSlug:   getStringClaim(claims, "role_slug"),
	}, nil
}

func employeeToUserResponse(employee *models.Employee) *models.UserResponse {
	return &models.UserResponse{
		ID:              employee.ID,
		EmployeeID:      employee.EmployeeID,
		FullName:        employee.FullName,
		Email:           employee.Email,
		RoleSlug:        employee.RoleSlug,
		RoleName:        employee.RoleName,
		PositionName:    employee.PositionName,
		DepartmentName:  employee.DepartmentName,
		AvatarInitials:  getInitials(employee.FullName),
	}
}

func getInitials(name string) string {
	if len(name) == 0 {
		return "NA"
	}
	runes := []rune(name)
	if len(runes) >= 2 {
		return string(runes[0]) + string(runes[1])
	}
	return string(runes[0])
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}
