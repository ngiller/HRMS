package middleware

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// CSP directives
	CSPDefaultSrc  string
	CSPScriptSrc   string
	CSPStyleSrc    string
	CSPImgSrc      string
	CSPConnectSrc  string
	CSPFontSrc     string
	CSPStyleSrcUnsafe string

	// Upload validation
	MaxUploadSize    int64
	AllowedMimeTypes []string
	AllowedExtensions []string
}

// DefaultSecurityConfig returns sensible defaults
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		CSPDefaultSrc:  "'self'",
		CSPScriptSrc:   "'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com",
		CSPStyleSrc:    "'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com https://fonts.googleapis.com",
		CSPImgSrc:      "'self' data: blob: https://*.tile.openstreetmap.org https://unpkg.com",
		CSPConnectSrc:  "'self' http://localhost:8590 http://localhost:8900 https://api.emailjs.com ws: wss:",
		CSPFontSrc:     "'self' data: https://fonts.gstatic.com",
		MaxUploadSize:  5 * 1024 * 1024, // 5MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"application/pdf",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-excel",
			"text/csv",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		AllowedExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".pdf", ".xlsx", ".xls", ".csv", ".doc", ".docx",
		},
	}
}

// SecurityHeadersMiddleware returns middleware that sets security-related HTTP headers
func SecurityHeadersMiddleware(cfg *SecurityConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Continue to next handler first to get the response
		if err := c.Next(); err != nil {
			return err
		}

		// Security Headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(self), camera=(self), microphone=(), payment=()")

		// Additional security headers (previously from helmet.New())
		c.Set("X-DNS-Prefetch-Control", "off")
		c.Set("X-Download-Options", "noopen")
		c.Set("X-Permitted-Cross-Domain-Policies", "none")
		c.Set("Origin-Agent-Cluster", "?1")

		// Content Security Policy
		if cfg != nil {
			csp := fmt.Sprintf(
				"default-src %s; script-src %s; style-src %s; img-src %s; connect-src %s; font-src %s; form-action 'self'; frame-ancestors 'none'; base-uri 'self'",
				cfg.CSPDefaultSrc, cfg.CSPScriptSrc, cfg.CSPStyleSrc,
				cfg.CSPImgSrc, cfg.CSPConnectSrc, cfg.CSPFontSrc,
			)
			c.Set("Content-Security-Policy", csp)
		}

		// HSTS (only if HTTPS)
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		return nil
	}
}

// FileUploadValidator returns middleware that validates file uploads
func FileUploadValidator(cfg *SecurityConfig) fiber.Handler {
	if cfg == nil {
		cfg = DefaultSecurityConfig()
	}

	return func(c *fiber.Ctx) error {
		// Only apply to multipart form data
		contentType := c.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			return c.Next()
		}

		// Parse multipart form with limit
		form, err := c.MultipartForm()
		if err != nil {
			// If there's an error parsing, it might be too large
			if strings.Contains(err.Error(), "http: request body too large") {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf("Ukuran file terlalu besar. Maksimal %d MB", cfg.MaxUploadSize/(1024*1024)),
				})
			}
			return c.Next()
		}

		// Validate each file
		for fieldName, fileHeaders := range form.File {
			for _, fh := range fileHeaders {
				// Check file size
				if fh.Size > cfg.MaxUploadSize {
					return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("File '%s' terlalu besar (%d MB). Maksimal %d MB", fh.Filename, fh.Size/(1024*1024), cfg.MaxUploadSize/(1024*1024)),
					})
				}

				// Check file extension
				ext := strings.ToLower(filepath.Ext(fh.Filename))
				if !isAllowedExtension(ext, cfg.AllowedExtensions) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("Tipe file '%s' tidak diizinkan. Ekstensi yang diizinkan: %s", fh.Filename, strings.Join(cfg.AllowedExtensions, ", ")),
					})
				}

				_ = fieldName // Field name not needed for validation
			}
		}

		return c.Next()
	}
}

// isAllowedExtension checks if a file extension is in the allowed list
func isAllowedExtension(ext string, allowed []string) bool {
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}


