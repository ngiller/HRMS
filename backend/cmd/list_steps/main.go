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

	rows, err := database.Pool.Query(ctx, `
		SELECT aws.step_order, aws.approver_type 
		FROM approval_workflow_steps aws
		JOIN approval_workflows aw ON aws.workflow_id = aw.id
		WHERE aw.entity_type = 'manual_attendance' AND aw.is_active = true
		ORDER BY aws.step_order
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Steps for manual_attendance:")
	for rows.Next() {
		var stepOrder int
		var approverType string
		rows.Scan(&stepOrder, &approverType)
		fmt.Printf("Step %d: %s\n", stepOrder, approverType)
	}
}
