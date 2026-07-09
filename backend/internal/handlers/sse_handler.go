package handlers

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SSEHandler handles Server-Sent Events connections
type SSEHandler struct {
	hub         *service.SSEHub
	authService *service.AuthService
}

// NewSSEHandler creates a new SSEHandler
func NewSSEHandler(hub *service.SSEHub, authService *service.AuthService) *SSEHandler {
	return &SSEHandler{
		hub:         hub,
		authService: authService,
	}
}

// HandleSSE handles SSE connections
// GET /api/sse/subscribe?token=JWT_TOKEN
func (h *SSEHandler) HandleSSE(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token diperlukan",
		})
	}

	// Validate JWT token
	claims, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token tidak valid",
		})
	}

	userID := claims.UserID.String()

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Subscribe to hub (defer cleanup)
	client := h.hub.Subscribe(userID)
	defer h.hub.Unsubscribe(client)

	log.Printf("[SSE] Connection established: user=%s", userID)

	// Use SetBodyStreamWriter which handles streaming + flushing via fasthttp's bufio writer
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		_ = w.Flush()

		keepalive := time.NewTicker(30 * time.Second)
		defer keepalive.Stop()

		for {
			select {
			case <-c.Context().Done():
				log.Printf("[SSE] Connection closed: user=%s", userID)
				return

			case data, ok := <-client.Events:
				if !ok {
					return
				}
				_, _ = fmt.Fprintf(w, "event: approval_update\ndata: %s\n\n", string(data))
				_ = w.Flush()

			case <-keepalive.C:
				// Send keepalive comment to prevent connection/load balancer timeout
				_, _ = fmt.Fprintf(w, ": keepalive\n\n")
				_ = w.Flush()
			}
		}
	})

	return nil
}
