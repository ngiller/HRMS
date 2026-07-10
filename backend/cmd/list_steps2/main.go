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
		SELECT aw.entity_type, aw.is_active, COALESCE(aws.step_order, 0), COALESCE(aws.approver_type, '') 
		FROM approval_workflows aw
		LEFT JOIN approval_workflow_steps aws ON aws.workflow_id = aw.id
		WHERE aw.entity_type = 'manual_attendance'
		ORDER BY aws.step_order
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Workflows for manual_attendance:")
	for rows.Next() {
		var etype string
		var isActive bool
		var stepOrder int
		var approverType string
		rows.Scan(&etype, &isActive, &stepOrder, &approverType)
		fmt.Printf("Type: %s, Active: %v, Step: %d, Approver: %s\n", etype, isActive, stepOrder, approverType)
	}
}
