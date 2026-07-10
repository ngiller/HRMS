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

	rows, err := database.Pool.Query(ctx, "SELECT id::text, entity_type, is_active FROM approval_workflows")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var etype string
		var active bool
		rows.Scan(&id, &etype, &active)
		fmt.Printf("WF ID: %s, Type: %s, Active: %v\n", id, etype, active)
	}

	fmt.Println("--- tracking ---")
	rows2, err := database.Pool.Query(ctx, "SELECT id::text, entity_type, current_step, status FROM approval_trackings")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id, etype, status string
		var step int
		rows2.Scan(&id, &etype, &step, &status)
		fmt.Printf("Track ID: %s, Type: %s, Step: %d, Status: %s\n", id, etype, step, status)
	}
}
