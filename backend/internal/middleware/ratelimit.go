package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// Rate limit tiers (based on Technical Spec section 15.2)
const (
	// TierCritical: Login, forgot-password — 5 req/min per IP
	TierCritical = iota + 1
	// TierHigh: Reset password, refresh token — 10 req/min per IP
	TierHigh
	// TierMedium: Check-in, create operations — 30 req/min per IP
	TierMedium
	// TierLow: All other endpoints — 100 req/min per IP
	TierLow
)

// RateLimitConfig returns a rate limiter middleware based on the specified tier.
//
// Usage:
//
//	auth.Post("/login", middleware.RateLimitConfig(middleware.TierCritical), handler.Login)
//	employees.Get("/", middleware.RateLimitConfig(middleware.TierLow), handler.List)
func RateLimitConfig(tier int) fiber.Handler {
	var cfg limiter.Config

	switch tier {
	case TierCritical:
		cfg = limiter.Config{
			Max:        5,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return "rl:critical:" + c.IP()
			},
			LimitReached: limitReachedHandler,
		}

	case TierHigh:
		cfg = limiter.Config{
			Max:        10,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return "rl:high:" + c.IP()
			},
			LimitReached: limitReachedHandler,
		}

	case TierMedium:
		cfg = limiter.Config{
			Max:        30,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return "rl:medium:" + c.IP()
			},
			LimitReached: limitReachedHandler,
		}

	case TierLow:
		fallthrough
	default:
		cfg = limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return "rl:low:" + c.IP()
			},
			LimitReached: limitReachedHandler,
		}
	}

	return limiter.New(cfg)
}

// limitReachedHandler is the callback when rate limit is exceeded.
func limitReachedHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"success": false,
		"message": "Terlalu banyak permintaan. Silakan coba lagi dalam 1 menit.",
	})
}

// ForgotPasswordRateLimit returns a rate limiter with per-email key.
// 3 requests per hour per email address.
func ForgotPasswordRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Parse email from body if possible, fallback to IP
			var body struct {
				Email string `json:"email"`
			}
			if err := c.BodyParser(&body); err == nil && body.Email != "" {
				return "rl:forgot-pw:" + body.Email
			}
			return "rl:forgot-pw:" + c.IP()
		},
		LimitReached: limitReachedHandler,
	})
}
