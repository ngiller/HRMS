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

	rows, err := database.Pool.Query(ctx, "SELECT id::text, status FROM manual_attendance_requests ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- Requests ---")
	for rows.Next() {
		var id, status string
		rows.Scan(&id, &status)
		fmt.Printf("Req ID: %s, Status: %s\n", id, status)
	}

	fmt.Println("--- tracking ---")
	rows2, err := database.Pool.Query(ctx, "SELECT id::text, entity_type, current_step, status FROM approval_request_tracking WHERE entity_type='manual_attendance' ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Query tracking failed: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id, etype, status string
		var step int
		rows2.Scan(&id, &etype, &step, &status)
		fmt.Printf("Track ID: %s, Type: %s, Step: %d, Status: %s\n", id, etype, step, status)
	}
}
