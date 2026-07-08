package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ==================== Workflows ====================

func ListApprovalWorkflows(ctx context.Context, entityType string) ([]models.ApprovalWorkflowSummary, error) {
	query := `SELECT aw.id, aw.entity_type, aw.name, COALESCE(aw.description, ''), aw.is_active,
		(SELECT COUNT(*) FROM approval_workflow_steps WHERE workflow_id = aw.id)
	FROM approval_workflows aw
	WHERE aw.deleted_at IS NULL`
	var args []interface{}
	argIdx := 1

	if entityType != "" {
		query += fmt.Sprintf(" AND aw.entity_type = $%d", argIdx)
		args = append(args, entityType)
		argIdx++
	}
	query += " ORDER BY aw.entity_type, aw.name"

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []models.ApprovalWorkflowSummary
	for rows.Next() {
		var w models.ApprovalWorkflowSummary
		if err := rows.Scan(&w.ID, &w.EntityType, &w.Name, &w.Description, &w.IsActive, &w.StepCount); err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}
	if workflows == nil {
		workflows = []models.ApprovalWorkflowSummary{}
	}
	return workflows, nil
}

func GetApprovalWorkflow(ctx context.Context, workflowID string) (*models.ApprovalWorkflow, error) {
	var w models.ApprovalWorkflow
	err := database.Pool.QueryRow(ctx, `
		SELECT id, entity_type, name, COALESCE(description, ''), is_active, created_at, updated_at
		FROM approval_workflows WHERE id::text = $1 AND deleted_at IS NULL
	`, workflowID).Scan(&w.ID, &w.EntityType, &w.Name, &w.Description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func GetApprovalWorkflowWithSteps(ctx context.Context, workflowID string) (*models.WorkflowDetailResponse, error) {
	w, err := GetApprovalWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, nil
	}

	steps, err := ListApprovalWorkflowSteps(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	w.Steps = steps

	return &models.WorkflowDetailResponse{
		Workflow: *w,
		Steps:    steps,
	}, nil
}

func GetActiveWorkflowByEntityType(ctx context.Context, entityType string) (*models.ApprovalWorkflow, error) {
	var w models.ApprovalWorkflow
	var description string
	err := database.Pool.QueryRow(ctx, `
		SELECT id, entity_type, name, COALESCE(description, ''), is_active, created_at, updated_at
		FROM approval_workflows
		WHERE entity_type = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT 1
	`, entityType).Scan(&w.ID, &w.EntityType, &w.Name, &description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	w.Description = description
	return &w, nil
}

func CreateApprovalWorkflow(ctx context.Context, req *models.CreateApprovalWorkflowReq) (*models.ApprovalWorkflow, error) {
	var w models.ApprovalWorkflow
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO approval_workflows (entity_type, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, entity_type, name, COALESCE(description, ''), is_active, created_at, updated_at
	`, req.EntityType, req.Name, req.Description).Scan(
		&w.ID, &w.EntityType, &w.Name, &w.Description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func UpdateApprovalWorkflow(ctx context.Context, workflowID string, req *models.CreateApprovalWorkflowReq) (*models.ApprovalWorkflow, error) {
	var w models.ApprovalWorkflow
	err := database.Pool.QueryRow(ctx, `
		UPDATE approval_workflows
		SET entity_type = $2, name = $3, description = $4, updated_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
		RETURNING id, entity_type, name, COALESCE(description, ''), is_active, created_at, updated_at
	`, workflowID, req.EntityType, req.Name, req.Description).Scan(
		&w.ID, &w.EntityType, &w.Name, &w.Description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func DeleteApprovalWorkflow(ctx context.Context, workflowID string) error {
	tag, err := database.Pool.Exec(ctx, `
		UPDATE approval_workflows SET deleted_at = NOW(), is_active = FALSE
		WHERE id::text = $1 AND deleted_at IS NULL
	`, workflowID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("workflow tidak ditemukan")
	}
	return nil
}

// ==================== Workflow Steps ====================

func ListApprovalWorkflowSteps(ctx context.Context, workflowID string) ([]models.ApprovalWorkflowStep, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT id, workflow_id, step_order, approver_type, approver_role_id,
			condition_field, condition_operator, condition_value,
			escalation_hours, created_at, updated_at
		FROM approval_workflow_steps
		WHERE workflow_id::text = $1
		ORDER BY step_order ASC
	`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.ApprovalWorkflowStep
	for rows.Next() {
		var s models.ApprovalWorkflowStep
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.StepOrder, &s.ApproverType, &s.ApproverRoleID,
			&s.ConditionField, &s.ConditionOp, &s.ConditionValue,
			&s.EscalationHours, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	if steps == nil {
		steps = []models.ApprovalWorkflowStep{}
	}
	return steps, nil
}

func CreateApprovalWorkflowStep(ctx context.Context, req *models.CreateApprovalWorkflowStepReq) (*models.ApprovalWorkflowStep, error) {
	var s models.ApprovalWorkflowStep
	var roleID *string
	if req.ApproverRoleID != nil {
		roleID = req.ApproverRoleID
	}

	err := database.Pool.QueryRow(ctx, `
		INSERT INTO approval_workflow_steps (workflow_id, step_order, approver_type, approver_role_id,
			condition_field, condition_operator, condition_value, escalation_hours)
		VALUES ($1::uuid, $2, $3, $4::uuid, $5, $6, $7, $8)
		RETURNING id, workflow_id, step_order, approver_type, approver_role_id,
			condition_field, condition_operator, condition_value, escalation_hours, created_at, updated_at
	`, req.WorkflowID, req.StepOrder, req.ApproverType, roleID,
		req.ConditionField, req.ConditionOp, req.ConditionValue, req.EscalationHours).Scan(
		&s.ID, &s.WorkflowID, &s.StepOrder, &s.ApproverType, &s.ApproverRoleID,
		&s.ConditionField, &s.ConditionOp, &s.ConditionValue,
		&s.EscalationHours, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func UpdateApprovalWorkflowStep(ctx context.Context, stepID string, req *models.UpdateApprovalWorkflowStepReq) (*models.ApprovalWorkflowStep, error) {
	var s models.ApprovalWorkflowStep
	var roleID *string
	if req.ApproverRoleID != nil {
		roleID = req.ApproverRoleID
	}

	err := database.Pool.QueryRow(ctx, `
		UPDATE approval_workflow_steps
		SET step_order = $2, approver_type = $3, approver_role_id = $4::uuid,
			condition_field = $5, condition_operator = $6, condition_value = $7,
			escalation_hours = $8, updated_at = NOW()
		WHERE id::text = $1
		RETURNING id, workflow_id, step_order, approver_type, approver_role_id,
			condition_field, condition_operator, condition_value, escalation_hours, created_at, updated_at
	`, stepID, req.StepOrder, req.ApproverType, roleID,
		req.ConditionField, req.ConditionOp, req.ConditionValue, req.EscalationHours).Scan(
		&s.ID, &s.WorkflowID, &s.StepOrder, &s.ApproverType, &s.ApproverRoleID,
		&s.ConditionField, &s.ConditionOp, &s.ConditionValue,
		&s.EscalationHours, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func DeleteApprovalWorkflowStep(ctx context.Context, stepID string) error {
	tag, err := database.Pool.Exec(ctx, `DELETE FROM approval_workflow_steps WHERE id::text = $1`, stepID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("step tidak ditemukan")
	}
	return nil
}

// ==================== Resolve Workflow for Entity ====================

// GetActiveStepsForEntityType returns all steps for the active workflow of an entity type
func GetActiveStepsForEntityType(ctx context.Context, entityType string) ([]models.ApprovalWorkflowStep, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT aws.id, aws.workflow_id, aws.step_order, aws.approver_type, aws.approver_role_id,
			aws.condition_field, aws.condition_operator, aws.condition_value,
			aws.escalation_hours, aws.created_at, aws.updated_at
		FROM approval_workflow_steps aws
		JOIN approval_workflows aw ON aw.id = aws.workflow_id
		WHERE aw.entity_type = $1 AND aw.is_active = TRUE AND aw.deleted_at IS NULL
		ORDER BY aws.step_order ASC
	`, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.ApprovalWorkflowStep
	for rows.Next() {
		var s models.ApprovalWorkflowStep
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.StepOrder, &s.ApproverType, &s.ApproverRoleID,
			&s.ConditionField, &s.ConditionOp, &s.ConditionValue,
			&s.EscalationHours, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, nil
}

// GetApproverByType returns the employee ID who should approve based on approver type
func GetApproverByType(ctx context.Context, approverType string, requestorID string, entityType string) (string, string, error) {
	switch approverType {
	case "approval_line":
		// Get the requestor's approval_line_id
		var approverID, approverName string
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(e.approval_line_id::text, ''), COALESCE(ae.full_name, '')
			FROM employees e
			LEFT JOIN employees ae ON ae.id = e.approval_line_id
			WHERE e.id::text = $1 AND e.deleted_at IS NULL
		`, requestorID).Scan(&approverID, &approverName)
		if err != nil {
			return "", "", err
		}
		if approverID == "" {
			return "", "", errors.New("approval line tidak ditemukan untuk karyawan ini")
		}
		return approverID, approverName, nil

	case "hr_manager":
		// Find employees with hr_manager role
		var approverID, approverName string
		err := database.Pool.QueryRow(ctx, `
			SELECT e.id::text, e.full_name FROM employees e
			JOIN roles r ON r.id = e.role_id
			WHERE r.slug = 'hr_manager' AND e.is_active = TRUE AND e.deleted_at IS NULL
			ORDER BY e.created_at ASC LIMIT 1
		`).Scan(&approverID, &approverName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Fallback to super_admin
				err = database.Pool.QueryRow(ctx, `
					SELECT e.id::text, e.full_name FROM employees e
					JOIN roles r ON r.id = e.role_id
					WHERE r.slug = 'super_admin' AND e.is_active = TRUE AND e.deleted_at IS NULL
					ORDER BY e.created_at ASC LIMIT 1
				`).Scan(&approverID, &approverName)
				if err != nil {
					return "", "", errors.New("tidak ada HR Manager atau admin yang tersedia")
				}
				return approverID, approverName, nil
			}
			return "", "", err
		}
		return approverID, approverName, nil

	case "finance":
		var approverID, approverName string
		err := database.Pool.QueryRow(ctx, `
			SELECT e.id::text, e.full_name FROM employees e
			JOIN roles r ON r.id = e.role_id
			WHERE r.slug = 'finance' AND e.is_active = TRUE AND e.deleted_at IS NULL
			ORDER BY e.created_at ASC LIMIT 1
		`).Scan(&approverID, &approverName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", "", errors.New("tidak ada staff Finance yang tersedia")
			}
			return "", "", err
		}
		return approverID, approverName, nil

	case "director":
		var approverID, approverName string
		err := database.Pool.QueryRow(ctx, `
			SELECT e.id::text, e.full_name FROM employees e
			JOIN roles r ON r.id = e.role_id
			WHERE r.slug = 'director' AND e.is_active = TRUE AND e.deleted_at IS NULL
			ORDER BY e.created_at ASC LIMIT 1
		`).Scan(&approverID, &approverName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", "", errors.New("tidak ada Direktur yang tersedia")
			}
			return "", "", err
		}
		return approverID, approverName, nil

	case "department_head":
		var approverID, approverName string
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(d.head_id::text, ''), COALESCE(e2.full_name, '')
			FROM employees e
			LEFT JOIN departments d ON d.id = e.department_id
			LEFT JOIN employees e2 ON e2.id = d.head_id
			WHERE e.id::text = $1 AND e.deleted_at IS NULL
		`, requestorID).Scan(&approverID, &approverName)
		if err != nil {
			return "", "", err
		}
		if approverID == "" {
			return "", "", errors.New("kepala departemen tidak ditemukan")
		}
		return approverID, approverName, nil

	default:
		return "", "", fmt.Errorf("tipe approver tidak dikenal: %s", approverType)
	}
}

// ==================== Approval Tracking ====================

func CreateApprovalTracking(ctx context.Context, entityType, entityID string, workflowID string, totalSteps int) (*models.ApprovalRequestTracking, error) {
	var t models.ApprovalRequestTracking
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO approval_request_tracking (entity_type, entity_id, workflow_id, current_step, total_steps, status)
		VALUES ($1, $2::uuid, $3::uuid, 1, $4, 'pending')
		ON CONFLICT (entity_type, entity_id) DO UPDATE
		SET current_step = 1, total_steps = $4, status = 'pending', updated_at = NOW()
		RETURNING id, entity_type, entity_id, workflow_id, current_step, total_steps, status, created_at, updated_at
	`, entityType, entityID, workflowID, totalSteps).Scan(
		&t.ID, &t.EntityType, &t.EntityID, &t.WorkflowID, &t.CurrentStep, &t.TotalSteps, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetApprovalTracking(ctx context.Context, entityType, entityID string) (*models.ApprovalRequestTracking, error) {
	var t models.ApprovalRequestTracking
	err := database.Pool.QueryRow(ctx, `
		SELECT id, entity_type, entity_id, workflow_id, current_step, total_steps, status, created_at, updated_at
		FROM approval_request_tracking
		WHERE entity_type = $1 AND entity_id::text = $2
	`, entityType, entityID).Scan(
		&t.ID, &t.EntityType, &t.EntityID, &t.WorkflowID, &t.CurrentStep, &t.TotalSteps, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func UpdateApprovalTrackingStatus(ctx context.Context, trackingID string, status string) error {
	_, err := database.Pool.Exec(ctx, `
		UPDATE approval_request_tracking
		SET current_step = 0, status = $2, updated_at = NOW()
		WHERE id::text = $1
	`, trackingID, status)
	return err
}

func UpdateApprovalTrackingStep(ctx context.Context, trackingID string, newStep int, status string) error {
	_, err := database.Pool.Exec(ctx, `
		UPDATE approval_request_tracking
		SET current_step = $2, status = $3, updated_at = NOW()
		WHERE id::text = $1
	`, trackingID, newStep, status)
	return err
}

// GetPendingApprovalsByUser returns all pending approval requests for a user
// based on their role and approval_line relationships
func GetPendingApprovalsByUser(ctx context.Context, userID string) ([]models.PendingApprovalItem, error) {
	// Get user's role
	var roleSlug string
	err := database.Pool.QueryRow(ctx, `
		SELECT r.slug FROM employees e
		JOIN roles r ON r.id = e.role_id
		WHERE e.id::text = $1 AND e.deleted_at IS NULL
	`, userID).Scan(&roleSlug)
	if err != nil {
		return nil, err
	}

	var items []models.PendingApprovalItem

	// 1. Pending as approval_line (direct reports)
	rows, err := database.Pool.Query(ctx, `
		SELECT
			art.id, art.entity_type, art.entity_id,
			e.full_name AS requestor_name, e.id AS requestor_id,
			art.current_step, art.total_steps,
			art.created_at,
			EXTRACT(EPOCH FROM (NOW() - art.created_at))/3600 AS elapsed_hours
		FROM approval_request_tracking art
		JOIN employees e ON e.id = art.entity_id
		WHERE art.status = 'pending'
			AND art.current_step > 0
			AND e.approval_line_id::text = $1
			AND e.deleted_at IS NULL
		ORDER BY art.created_at ASC
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item models.PendingApprovalItem
			if err := rows.Scan(&item.TrackingID, &item.EntityType, &item.EntityID,
				&item.RequestorName, &item.RequestorID,
				&item.CurrentStep, &item.TotalSteps,
				&item.CreatedAt, &item.ElapsedHours); err != nil {
				continue
			}
			item.ApproverType = "approval_line"
			items = append(items, item)
		}
	}

	// 2. Pending as role-based approver (hr_manager, finance, director)
	if roleSlug == "hr_manager" || roleSlug == "super_admin" {
		hrRows, err := database.Pool.Query(ctx, `
			SELECT art.id, art.entity_type, art.entity_id,
				e.full_name AS requestor_name, e.id AS requestor_id,
				art.current_step, art.total_steps,
				art.created_at,
				EXTRACT(EPOCH FROM (NOW() - art.created_at))/3600 AS elapsed_hours
			FROM approval_request_tracking art
			JOIN employees e ON e.id = art.entity_id
			WHERE art.status = 'pending'
				AND art.current_step > 0
				AND e.deleted_at IS NULL
				AND (
					EXISTS (
						SELECT 1 FROM approval_workflow_steps aws
						JOIN approval_workflows aw ON aw.id = aws.workflow_id AND aw.id = art.workflow_id
						WHERE aws.step_order = art.current_step
						AND aws.approver_type IN ('hr_manager', 'director')
					)
					OR $1 = 'super_admin'
				)
			ORDER BY art.created_at ASC
		`, userID)
		if err == nil {
			defer hrRows.Close()
			for hrRows.Next() {
				var item models.PendingApprovalItem
				if err := hrRows.Scan(&item.TrackingID, &item.EntityType, &item.EntityID,
					&item.RequestorName, &item.RequestorID,
					&item.CurrentStep, &item.TotalSteps,
					&item.CreatedAt, &item.ElapsedHours); err != nil {
					continue
				}
				// Check for duplicate
				dup := false
				for i := range items {
					if items[i].TrackingID == item.TrackingID {
						dup = true
						break
					}
				}
				if !dup {
					item.ApproverType = "hr_manager"
					items = append(items, item)
				}
			}
		}
	}

	if roleSlug == "finance" || roleSlug == "super_admin" {
		finRows, err := database.Pool.Query(ctx, `
			SELECT art.id, art.entity_type, art.entity_id,
				e.full_name AS requestor_name, e.id AS requestor_id,
				art.current_step, art.total_steps,
				art.created_at,
				EXTRACT(EPOCH FROM (NOW() - art.created_at))/3600 AS elapsed_hours
			FROM approval_request_tracking art
			JOIN employees e ON e.id = art.entity_id
			WHERE art.status = 'pending'
				AND art.current_step > 0
				AND e.deleted_at IS NULL
				AND (
					EXISTS (
						SELECT 1 FROM approval_workflow_steps aws
						JOIN approval_workflows aw ON aw.id = aws.workflow_id AND aw.id = art.workflow_id
						WHERE aws.step_order = art.current_step
						AND aws.approver_type = 'finance'
					)
					OR $1 = 'super_admin'
				)
			ORDER BY art.created_at ASC
		`, userID)
		if err == nil {
			defer finRows.Close()
			for finRows.Next() {
				var item models.PendingApprovalItem
				if err := finRows.Scan(&item.TrackingID, &item.EntityType, &item.EntityID,
					&item.RequestorName, &item.RequestorID,
					&item.CurrentStep, &item.TotalSteps,
					&item.CreatedAt, &item.ElapsedHours); err != nil {
					continue
				}
				dup := false
				for i := range items {
					if items[i].TrackingID == item.TrackingID {
						dup = true
						break
					}
				}
				if !dup {
					item.ApproverType = "finance"
					items = append(items, item)
				}
			}
		}
	}

	if items == nil {
		items = []models.PendingApprovalItem{}
	}
	return items, nil
}

// Helper: evaluate condition
func EvaluateCondition(step models.ApprovalWorkflowStep, value float64) bool {
	if step.ConditionField == nil || step.ConditionOp == nil || step.ConditionValue == nil {
		return true // no condition = always apply
	}
	switch *step.ConditionOp {
	case ">":
		return value > *step.ConditionValue
	case ">=":
		return value >= *step.ConditionValue
	case "<":
		return value < *step.ConditionValue
	case "<=":
		return value <= *step.ConditionValue
	case "==":
		return value == *step.ConditionValue
	default:
		return true
	}
}

// UpdateApprovalTrail updates the approval_trail JSONB field on an entity
func UpdateApprovalTrail(ctx context.Context, entityType, entityID, trailJSON string) error {
	var tableName, idColumn string
	switch entityType {
	case "leave":
		tableName = "leave_requests"
		idColumn = "id"
	case "overtime":
		tableName = "overtime_requests"
		idColumn = "id"
	case "reimbursement":
		tableName = "reimbursements"
		idColumn = "id"
	case "shift_change":
		tableName = "shift_change_requests"
		idColumn = "id"
	case "loan":
		tableName = "loans"
		idColumn = "id"
	case "manual_attendance":
		tableName = "manual_attendance_requests"
		idColumn = "id"
	case "resign":
		tableName = "resign_requests"
		idColumn = "id"
	case "mutation":
		tableName = "employee_mutations"
		idColumn = "id"
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	query := fmt.Sprintf(`UPDATE %s SET approval_trail = $1::jsonb, updated_at = NOW() WHERE %s::text = $2`, tableName, idColumn)
	_, err := database.Pool.Exec(ctx, query, trailJSON, entityID)
	return err
}

// UpdateEntityStatus updates the status field on an entity
func UpdateEntityStatus(ctx context.Context, entityType, entityID, status string) error {
	var tableName, idColumn string
	possibleColumns := []string{"status", "leave_status", "overtime_status", "reimbursement_status", "shift_change_status", "loan_status"}
	_ = possibleColumns // we'll use just "status" for now since that's what most tables use

	switch entityType {
	case "leave":
		tableName = "leave_requests"
		idColumn = "id"
	case "overtime":
		tableName = "overtime_requests"
		idColumn = "id"
	case "reimbursement":
		tableName = "reimbursements"
		idColumn = "id"
	case "shift_change":
		tableName = "shift_change_requests"
		idColumn = "id"
	case "loan":
		tableName = "loans"
		idColumn = "id"
	case "manual_attendance":
		tableName = "manual_attendance_requests"
		idColumn = "id"
	case "resign":
		tableName = "resign_requests"
		idColumn = "id"
	case "mutation":
		tableName = "employee_mutations"
		idColumn = "id"
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	query := fmt.Sprintf(`UPDATE %s SET status = $1::text, updated_at = NOW() WHERE %s::text = $2`, tableName, idColumn)
	tag, err := database.Pool.Exec(ctx, query, status, entityID)
	if err != nil {
		// Try with status text without cast
		query = fmt.Sprintf(`UPDATE %s SET status = $1, updated_at = NOW() WHERE %s::text = $2`, tableName, idColumn)
		tag, err = database.Pool.Exec(ctx, query, status, entityID)
		if err != nil {
			return err
		}
	}
	if tag.RowsAffected() == 0 {
		return errors.New("entity tidak ditemukan")
	}
	return nil
}


