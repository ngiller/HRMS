package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main1() {
	dbURL := "postgres://tisen:tisen123@localhost:5432/hrms?sslmode=disable"
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// 1. Cuti (Single Example)
	pool.Exec(ctx, "UPDATE approval_workflows SET name = 'Cuti Tahunan (Contoh Single)' WHERE entity_type = 'leave'")
	pool.Exec(ctx, "DELETE FROM approval_workflow_steps WHERE workflow_id = (SELECT id FROM approval_workflows WHERE entity_type = 'leave')")
	pool.Exec(ctx, `
		INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, step_mode, escalation_hours)
		VALUES ((SELECT id FROM approval_workflows WHERE entity_type = 'leave'), 1, 'approval_line', 'single', 48)
	`)

	// 2. Shift Change (Parallel Example)
	pool.Exec(ctx, "UPDATE approval_workflows SET name = 'Perubahan Shift (Contoh Parallel)' WHERE entity_type = 'shift_change'")
	pool.Exec(ctx, "DELETE FROM approval_workflow_steps WHERE workflow_id = (SELECT id FROM approval_workflows WHERE entity_type = 'shift_change')")
	pool.Exec(ctx, `
		INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, step_mode, escalation_hours)
		VALUES 
		((SELECT id FROM approval_workflows WHERE entity_type = 'shift_change'), 1, 'department_head', 'any', 48),
		((SELECT id FROM approval_workflows WHERE entity_type = 'shift_change'), 1, 'hr_manager', 'any', 48)
	`)

	// 3. Reimbursement (Mixed Example)
	pool.Exec(ctx, "UPDATE approval_workflows SET name = 'Reimbursement (Contoh Mix)' WHERE entity_type = 'reimbursement'")
	pool.Exec(ctx, "DELETE FROM approval_workflow_steps WHERE workflow_id = (SELECT id FROM approval_workflows WHERE entity_type = 'reimbursement')")
	pool.Exec(ctx, `
		INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, step_mode, condition_field, condition_operator, condition_value, escalation_hours)
		VALUES 
		-- Step 1: Single (Atasan Langsung)
		((SELECT id FROM approval_workflows WHERE entity_type = 'reimbursement'), 1, 'approval_line', 'single', NULL, NULL, NULL, 48),
		-- Step 2: Parallel (HR atau Finance)
		((SELECT id FROM approval_workflows WHERE entity_type = 'reimbursement'), 2, 'hr_manager', 'any', NULL, NULL, NULL, 48),
		((SELECT id FROM approval_workflows WHERE entity_type = 'reimbursement'), 2, 'finance', 'any', NULL, NULL, NULL, 48),
		-- Step 3: Single Conditional (Direktur jika > 5 Juta)
		((SELECT id FROM approval_workflows WHERE entity_type = 'reimbursement'), 3, 'director', 'single', 'amount', '>', 5000000, 48)
	`)

	fmt.Println("Successfully seeded workflow examples!")
}
