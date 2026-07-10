package main

import (
	"context"
	"fmt"
	"log"

	"hrms-backend/internal/config"
	"hrms-backend/internal/database"
)

func main() {
	cfg := config.Load()
	if err := database.Connect(cfg.DatabaseURL(), cfg.EncryptionKey); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// 1. Ensure attendance object exists or set it if not
	// 2. Set approve to true
	for _, slug := range []string{"hr_manager", "hr_staff", "super_admin"} {
		// Merge existing attendance permissions or create a new one with approve: true
		_, err := database.Pool.Exec(ctx, `
			UPDATE roles 
			SET permissions = jsonb_set(
				permissions, 
				'{attendance}', 
				COALESCE(permissions->'attendance', '{}'::jsonb) || '{"approve": true, "read": true, "update": true, "create": true, "delete": true}'::jsonb,
				true
			)
			WHERE slug = $1
		`, slug)
		if err != nil {
			log.Printf("Failed for %s: %v", slug, err)
		} else {
			fmt.Printf("Fully updated attendance permissions for %s\n", slug)
		}
	}
}
