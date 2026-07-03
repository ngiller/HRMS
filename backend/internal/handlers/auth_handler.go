package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles user authentication
// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Email harus diisi"))
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Password harus diisi"))
	}

	ipAddress := c.IP()

	resp, err := h.authService.Login(c.Context(), req, ipAddress)
	if err != nil {
		log.Warnf("Login failed for %s: %v", req.Email, err)
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp, "Login berhasil"))
}

// ForgotPassword sends password reset email
// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Email harus diisi"))
	}

	resp, err := h.authService.ForgotPassword(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp, "Jika email terdaftar, Anda akan menerima tautan reset password"))
}

// ResetPassword handles password reset with token
// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Token harus diisi"))
	}
	if req.NewPassword == "" || len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Password baru minimal 6 karakter"))
	}

	resp, err := h.authService.ResetPassword(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp, "Password berhasil direset, silakan login dengan password baru"))
}

// Me returns current user info
// GET /api/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(user, "Berhasil memuat data pengguna"))
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Refresh token harus diisi"))
	}

	resp, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(resp, "Token berhasil diperbarui"))
}

// ChangePassword handles password change for authenticated user
// PUT /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req models.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.CurrentPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Password saat ini harus diisi"))
	}
	if req.NewPassword == "" || len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Password baru minimal 6 karakter"))
	}
	if req.CurrentPassword == req.NewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Password baru harus berbeda dari password saat ini"))
	}

	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	if err := h.authService.ChangePassword(c.Context(), userID, req); err != nil {
		status := fiber.StatusInternalServerError
		switch err.Error() {
		case "password saat ini salah":
			status = fiber.StatusBadRequest
		case "karyawan tidak ditemukan":
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Password berhasil diubah"))
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.BodyParser(&req) // ignore parse error, optional field

	var uid uuid.UUID
	if id, ok := userID.([16]byte); ok {
		uid = uuid.UUID(id)
	} else if idStr, ok := userID.(string); ok {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("ID pengguna tidak valid"))
		}
		uid = parsed
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("ID pengguna tidak ditemukan"))
	}

	if err := h.authService.Logout(c.Context(), uid, req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{}, "Berhasil logout"))
}
