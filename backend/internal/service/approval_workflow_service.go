package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type ApprovalWorkflowService struct{}

func NewApprovalWorkflowService() *ApprovalWorkflowService {
	return &ApprovalWorkflowService{}
}

// ─── Workflow Configuration ────────────────────────────────────

func (s *ApprovalWorkflowService) ListWorkflows(ctx context.Context, entityType string) (*[]models.ApprovalWorkflowSummary, error) {
	workflows, err := repository.ListApprovalWorkflows(ctx, entityType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar workflow: %w", err)
	}
	return &workflows, nil
}

func (s *ApprovalWorkflowService) GetWorkflowDetail(ctx context.Context, workflowID string) (*models.WorkflowDetailResponse, error) {
	if workflowID == "" {
		return nil, errors.New("ID workflow harus diisi")
	}
	return repository.GetApprovalWorkflowWithSteps(ctx, workflowID)
}

func (s *ApprovalWorkflowService) CreateWorkflow(ctx context.Context, req *models.CreateApprovalWorkflowReq) (*models.ApprovalWorkflow, error) {
	if req.EntityType == "" {
		return nil, errors.New("tipe entity harus diisi")
	}
	if req.Name == "" {
		return nil, errors.New("nama workflow harus diisi")
	}
	return repository.CreateApprovalWorkflow(ctx, req)
}

func (s *ApprovalWorkflowService) AddWorkflowStep(ctx context.Context, req *models.CreateApprovalWorkflowStepReq) (*models.ApprovalWorkflowStep, error) {
	if req.WorkflowID == "" {
		return nil, errors.New("ID workflow harus diisi")
	}
	if req.ApproverType == "" {
		return nil, errors.New("tipe approver harus diisi")
	}
	if req.StepOrder < 1 {
		return nil, errors.New("urutan step harus >= 1")
	}
	return repository.CreateApprovalWorkflowStep(ctx, req)
}

func (s *ApprovalWorkflowService) UpdateWorkflowStep(ctx context.Context, stepID string, req *models.UpdateApprovalWorkflowStepReq) (*models.ApprovalWorkflowStep, error) {
	if stepID == "" {
		return nil, errors.New("ID step harus diisi")
	}
	return repository.UpdateApprovalWorkflowStep(ctx, stepID, req)
}

func (s *ApprovalWorkflowService) DeleteWorkflowStep(ctx context.Context, stepID string) error {
	if stepID == "" {
		return errors.New("ID step harus diisi")
	}
	return repository.DeleteApprovalWorkflowStep(ctx, stepID)
}

func (s *ApprovalWorkflowService) DeleteWorkflow(ctx context.Context, workflowID string) error {
	if workflowID == "" {
		return errors.New("ID workflow harus diisi")
	}
	return repository.DeleteApprovalWorkflow(ctx, workflowID)
}

// ─── Workflow Resolution & Processing ──────────────────────────

// ResolveWorkflowForRequest determines which steps apply to a request and creates tracking.
// Called when a new request is created (e.g. leave, overtime).
func (s *ApprovalWorkflowService) ResolveWorkflowForRequest(ctx context.Context, entityType, entityID, employeeID string, conditionValue float64) (*models.ApprovalResult, error) {
	// Get active workflow
	workflow, err := repository.GetActiveWorkflowByEntityType(ctx, entityType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat workflow: %w", err)
	}
	if workflow == nil {
		return nil, errors.New("tidak ada workflow aktif untuk " + entityType)
	}

	// Get all steps
	allSteps, err := repository.GetActiveStepsForEntityType(ctx, entityType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat step workflow: %w", err)
	}

	// Filter steps by condition
	var applicableSteps []models.ApprovalWorkflowStep
	for _, step := range allSteps {
		if repository.EvaluateCondition(step, conditionValue) {
			applicableSteps = append(applicableSteps, step)
		}
	}

	if len(applicableSteps) == 0 {
		return nil, errors.New("tidak ada step approval yang sesuai untuk request ini")
	}

	// Create tracking
	_, err = repository.CreateApprovalTracking(ctx, entityType, entityID, workflow.ID.String(), len(applicableSteps))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat tracking approval: %w", err)
	}

	// Get the first approver
	firstStep := applicableSteps[0]
	approverID, approverName, err := repository.GetApproverByType(ctx, firstStep.ApproverType, employeeID, entityType)
	if err != nil {
		return nil, fmt.Errorf("gagal menentukan approver pertama: %w", err)
	}

	// Initialize approval trail
	initTrail := []map[string]interface{}{
		{
			"step":          firstStep.StepOrder,
			"level":         1,
			"approver_id":   approverID,
			"approver_name": approverName,
			"status":        "pending",
			"note":          "",
			"date":          nil,
		},
	}
	trailJSON, _ := json.Marshal(initTrail)

	_ = repository.UpdateApprovalTrail(ctx, entityType, entityID, string(trailJSON))

	// Send notification to approver
	notifRepo := repository.NewNotificationRepo()
	entityLabel := s.getEntityLabel(entityType)
	notifReq := &models.CreateNotificationRequest{
		UserID:           approverID,
		NotificationType: "approval_request",
		Title:            "Pengajuan Baru Perlu Disetujui",
		Body:             fmt.Sprintf("Ada pengajuan %s baru yang perlu Anda setujui. Step %d dari %d.", entityLabel, 1, len(applicableSteps)),
		Data: map[string]any{
			"type":       entityType,
			"entity_id":  entityID,
			"step":       1,
			"totalSteps": len(applicableSteps),
		},
	}
	_, _ = notifRepo.CreateNotification(ctx, notifReq)
	// Send email to approver (if email service is configured)
	SendEmailForUser(ctx, approverID, "Pengajuan Baru: "+entityLabel,
		fmt.Sprintf("Ada pengajuan %s baru yang perlu Anda setujui.", entityLabel))

	return &models.ApprovalResult{
		Action:       "pending",
		CurrentStep:  1,
		TotalSteps:   len(applicableSteps),
		NextApprover: approverName,
		Message:      fmt.Sprintf("Menunggu approval %s", approverName),
	}, nil
}

// isSuperAdmin checks if the given user is a super_admin
func (s *ApprovalWorkflowService) isSuperAdmin(ctx context.Context, userID string) bool {
	var roleSlug string
	err := database.Pool.QueryRow(ctx, `
		SELECT r.slug FROM employees e
		JOIN roles r ON r.id = e.role_id
		WHERE e.id::text = $1 AND e.deleted_at IS NULL
	`, userID).Scan(&roleSlug)
	return err == nil && roleSlug == "super_admin"
}

// ProcessApproval handles approve/reject for a request at the current level.
// Uses transaction with WithUserContext for data integrity + audit trail.
func (s *ApprovalWorkflowService) ProcessApproval(ctx context.Context, entityType, entityID, approverID, action, note string) (*models.ApprovalResult, error) {
	// Get tracking
	tracking, err := repository.GetApprovalTracking(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat tracking approval: %w", err)
	}
	if tracking == nil {
		return nil, errors.New("tracking approval tidak ditemukan")
	}
	if tracking.Status != "pending" {
		return nil, errors.New("request ini sudah diproses")
	}
	if tracking.CurrentStep < 1 {
		return nil, errors.New("request ini sudah selesai diproses")
	}

	// Check super_admin bypass
	isSuperAdmin := s.isSuperAdmin(ctx, approverID)

	// Get workflow steps
	allSteps, err := repository.GetActiveStepsForEntityType(ctx, entityType)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat step workflow: %w", err)
	}

	// Get condition value from entity with proper error handling
	conditionValue, err := s.getEntityConditionValue(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	// Filter steps by condition
	var applicableSteps []models.ApprovalWorkflowStep
	for _, step := range allSteps {
		if repository.EvaluateCondition(step, conditionValue) {
			applicableSteps = append(applicableSteps, step)
		}
	}

	// Find current step index
	currentStepIdx := -1
	for i, step := range applicableSteps {
		if step.StepOrder == tracking.CurrentStep {
			currentStepIdx = i
			break
		}
	}
	if currentStepIdx < 0 {
		return nil, errors.New("step saat ini tidak ditemukan dalam workflow")
	}

	currentStep := applicableSteps[currentStepIdx]

	// Resolve the actual approver and their name
	var expectedApproverID, expectedApproverName string

	if isSuperAdmin {
		// Super admin bypasses the approver verification entirely
		// Use their own name for the trail
		_ = database.Pool.QueryRow(ctx, `SELECT full_name FROM employees WHERE id::text = $1`, approverID).Scan(&expectedApproverName)
		if expectedApproverName == "" {
			expectedApproverName = approverID
		}
	} else {
		// Get the requestor's employee ID to resolve approval_line correctly
		requestorID, err := repository.GetEntityRequestorID(ctx, entityType, entityID)
		if err != nil {
			return nil, fmt.Errorf("gagal menentukan pengaju: %w", err)
		}

		expectedApproverID, expectedApproverName, err = repository.GetApproverByType(ctx, currentStep.ApproverType, requestorID, entityType)
		if err != nil {
			return nil, fmt.Errorf("gagal verifikasi approver: %w", err)
		}
		if expectedApproverID != approverID {
			return nil, errors.New("anda tidak berwenang untuk melakukan approval ini")
		}
	}

	actualApproverName := expectedApproverName

	// Execute approval in transaction
	return s.processApprovalTx(ctx, entityType, entityID, approverID, action, note,
		tracking, currentStep, currentStepIdx, applicableSteps, actualApproverName)
}

func (s *ApprovalWorkflowService) processApprovalTx(
	ctx context.Context,
	entityType, entityID, approverID, action, note string,
	tracking *models.ApprovalRequestTracking,
	currentStep models.ApprovalWorkflowStep,
	currentStepIdx int,
	applicableSteps []models.ApprovalWorkflowStep,
	actualApproverName string,
) (*models.ApprovalResult, error) {
	var result *models.ApprovalResult

	// Get the actual requestor (employee) ID from the entity table
	// tracking.EntityID is the request UUID, NOT the employee UUID
	requestorID, _ := repository.GetEntityRequestorID(ctx, entityType, entityID)
	if requestorID == "" {
		requestorID = tracking.EntityID.String()
	}

	err := database.WithUserContext(ctx, approverID, func(tx pgx.Tx) error {
		if action == "reject" {
			// Update tracking (within transaction)
			if err := repository.UpdateApprovalTrackingStepTx(ctx, tx, tracking.ID.String(), -1, "rejected"); err != nil {
				return fmt.Errorf("gagal update tracking: %w", err)
			}
			// Update entity status (within transaction)
			if err := repository.UpdateEntityStatusTx(ctx, tx, entityType, entityID, "rejected"); err != nil {
				return fmt.Errorf("gagal update status: %w", err)
			}

			// Update approval trail (within transaction)
			trailEntry := map[string]interface{}{
				"step":          currentStep.StepOrder,
				"level":         currentStepIdx + 1,
				"approver_id":   approverID,
				"approver_name": actualApproverName,
				"status":        "rejected",
				"note":          note,
				"date":          time.Now().UTC().Format(time.RFC3339),
			}
			s.appendToApprovalTrailTx(ctx, tx, entityType, entityID, trailEntry)

			// Notify requestor about rejection (notifications don't need to be transactional)
			notifRepo := repository.NewNotificationRepo()
			entityLabel := s.getEntityLabel(entityType)
			_, _ = notifRepo.CreateNotification(ctx, &models.CreateNotificationRequest{
				UserID:           requestorID,
				NotificationType: entityType + "_rejected",
				Title:            entityLabel + " Ditolak",
				Body:             fmt.Sprintf("Pengajuan %s Anda ditolak oleh %s.", entityLabel, actualApproverName),
				Data: map[string]any{
					"type":      entityType,
					"entity_id": entityID,
					"status":    "rejected",
					"note":      note,
				},
			})
			// Send email to requestor
			SendEmailForUser(ctx, requestorID, entityLabel+" Ditolak",
				fmt.Sprintf("Pengajuan %s Anda ditolak oleh %s.", entityLabel, actualApproverName))

			result = &models.ApprovalResult{
				Action:      "rejected",
				FinalStatus: "rejected",
				Message:     "Pengajuan ditolak",
			}
			return nil
		}

		// Approve at current level
		trailEntry := map[string]interface{}{
			"step":          currentStep.StepOrder,
			"level":         currentStepIdx + 1,
			"approver_id":   approverID,
			"approver_name": actualApproverName,
			"status":        "approved",
			"note":          note,
			"date":          time.Now().UTC().Format(time.RFC3339),
		}
		s.appendToApprovalTrailTx(ctx, tx, entityType, entityID, trailEntry)

		// Check if there are more steps
		if currentStepIdx+1 >= len(applicableSteps) {
			// Final approval (within transaction)
			if err := repository.UpdateApprovalTrackingStepTx(ctx, tx, tracking.ID.String(), 0, "approved"); err != nil {
				return fmt.Errorf("gagal update tracking: %w", err)
			}
			if err := repository.UpdateEntityStatusTx(ctx, tx, entityType, entityID, "approved"); err != nil {
				return fmt.Errorf("gagal update status: %w", err)
			}

			// Notify requestor about final approval
			notifRepo := repository.NewNotificationRepo()
			entityLabel := s.getEntityLabel(entityType)
			_, _ = notifRepo.CreateNotification(ctx, &models.CreateNotificationRequest{
				UserID:           requestorID,
				NotificationType: entityType + "_approved",
				Title:            entityLabel + " Disetujui",
				Body:             fmt.Sprintf("Pengajuan %s Anda telah disetujui sepenuhnya oleh %s.", entityLabel, actualApproverName),
				Data: map[string]any{
					"type":      entityType,
					"entity_id": entityID,
					"status":    "approved",
				},
			})
			// Send email to requestor about final approval
			SendEmailForUser(ctx, requestorID, entityLabel+" Disetujui",
				fmt.Sprintf("Pengajuan %s Anda telah disetujui sepenuhnya oleh %s.", entityLabel, actualApproverName))

			result = &models.ApprovalResult{
				Action:      "approved",
				FinalStatus: "approved",
				Message:     "Pengajuan disetujui sepenuhnya",
			}
			return nil
		}

		// Notify requestor about current level approval
		notifRepo := repository.NewNotificationRepo()
		entityLabel := s.getEntityLabel(entityType)
		_, _ = notifRepo.CreateNotification(ctx, &models.CreateNotificationRequest{
			UserID:           requestorID,
			NotificationType: entityType + "_level_approved",
			Title:            entityLabel + " Disetujui Level " + fmt.Sprintf("%d", currentStepIdx+1),
			Body:             fmt.Sprintf("Pengajuan %s Anda telah disetujui di level %d oleh %s.", entityLabel, currentStepIdx+1, actualApproverName),
			Data: map[string]any{
				"type":      entityType,
				"entity_id": entityID,
				"status":    "pending",
				"level":     currentStepIdx + 1,
			},
		})
		// Send email to requestor about level approval
		SendEmailForUser(ctx, requestorID, entityLabel+" Disetujui Level "+fmt.Sprintf("%d", currentStepIdx+1),
			fmt.Sprintf("Pengajuan %s Anda telah disetujui di level %d oleh %s.", entityLabel, currentStepIdx+1, actualApproverName))

		// Move to next step
		nextStep := applicableSteps[currentStepIdx+1]
		nextApproverID, nextApproverName, err := repository.GetApproverByType(ctx, nextStep.ApproverType, "", entityType)
		if err != nil {
			return fmt.Errorf("gagal menentukan approver berikutnya: %w", err)
		}

		if err := repository.UpdateApprovalTrackingStepTx(ctx, tx, tracking.ID.String(), nextStep.StepOrder, "pending"); err != nil {
			return fmt.Errorf("gagal update tracking: %w", err)
		}

		// Add next step to trail (within transaction)
		nextTrailEntry := map[string]interface{}{
			"step":          nextStep.StepOrder,
			"level":         currentStepIdx + 2,
			"approver_id":   nextApproverID,
			"approver_name": nextApproverName,
			"status":        "pending",
			"note":          "",
			"date":          nil,
		}
		s.appendToApprovalTrailTx(ctx, tx, entityType, entityID, nextTrailEntry)

		// Notify next approver
		nextNotifReq := &models.CreateNotificationRequest{
			UserID:           nextApproverID,
			NotificationType: "approval_request",
			Title:            "Pengajuan Baru Perlu Disetujui",
			Body:             fmt.Sprintf("Ada pengajuan %s yang telah disetujui level sebelumnya dan kini menunggu persetujuan Anda. Step %d dari %d.", entityLabel, currentStepIdx+2, len(applicableSteps)),
			Data: map[string]any{
				"type":       entityType,
				"entity_id":  entityID,
				"step":       currentStepIdx + 2,
				"totalSteps": len(applicableSteps),
			},
		}
		_, _ = notifRepo.CreateNotification(ctx, nextNotifReq)
		// Send email to next approver
		SendEmailForUser(ctx, nextApproverID, "Pengajuan Baru Perlu Disetujui",
			fmt.Sprintf("Ada pengajuan %s yang menunggu persetujuan Anda. Step %d dari %d.", entityLabel, currentStepIdx+2, len(applicableSteps)))

		result = &models.ApprovalResult{
			Action:       "pending_next_level",
			CurrentStep:  nextStep.StepOrder,
			TotalSteps:   len(applicableSteps),
			NextApprover: nextApproverName,
			Message:      fmt.Sprintf("Disetujui level %d. Menunggu approval %s", currentStepIdx+1, nextApproverName),
		}
		return nil
	})

	return result, err
}

// getEntityConditionValue retrieves the numeric value for condition evaluation
func (s *ApprovalWorkflowService) getEntityConditionValue(ctx context.Context, entityType, entityID string) (float64, error) {
	switch entityType {
	case "leave":
		var totalDays float64
		err := database.Pool.QueryRow(ctx,
			`SELECT total_days::float FROM leave_requests WHERE id::text = $1`, entityID).Scan(&totalDays)
		if err != nil {
			return 0, errors.New("data cuti tidak ditemukan")
		}
		return totalDays, nil
	case "reimbursement":
		var amount float64
		err := database.Pool.QueryRow(ctx,
			`SELECT amount FROM reimbursements WHERE id::text = $1`, entityID).Scan(&amount)
		if err != nil {
			return 0, errors.New("data reimbursement tidak ditemukan")
		}
		return amount, nil
	case "loan":
		var amount float64
		err := database.Pool.QueryRow(ctx,
			`SELECT amount FROM loans WHERE id::text = $1`, entityID).Scan(&amount)
		if err != nil {
			return 0, errors.New("data pinjaman tidak ditemukan")
		}
		return amount, nil
	case "mutation":
		return 0, nil
	default:
		return 0, nil
	}
}

// CancelWorkflowTracking cancels pending approval tracking for a request when it's cancelled by the user.
func (s *ApprovalWorkflowService) CancelWorkflowTracking(ctx context.Context, entityType, entityID string) error {
	tracking, err := repository.GetApprovalTracking(ctx, entityType, entityID)
	if err != nil {
		return fmt.Errorf("gagal memuat tracking approval: %w", err)
	}
	if tracking == nil {
		return nil // no tracking to cancel
	}
	if tracking.Status != "pending" {
		return nil // already processed, no need to cancel
	}

	// Update tracking to cancelled
	if err := repository.UpdateApprovalTrackingStatus(ctx, tracking.ID.String(), "cancelled"); err != nil {
		return fmt.Errorf("gagal update status tracking: %w", err)
	}

	// Update approval_trail: mark all pending steps as cancelled
	var currentTrail string
	switch entityType {
	case "leave":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM leave_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "overtime":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM overtime_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "reimbursement":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM reimbursements WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "shift_change":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM shift_change_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "loan":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM loans WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "manual_attendance":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM manual_attendance_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "resign":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM resign_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
	case "mutation":
		_ = database.Pool.QueryRow(ctx, `SELECT COALESCE(approval_trail, '[]') FROM employee_mutations WHERE id::text = $1`, entityID).Scan(&currentTrail)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	if currentTrail == "" || currentTrail == "[]" {
		return nil
	}

	var trail []map[string]interface{}
	if err := json.Unmarshal([]byte(currentTrail), &trail); err != nil {
		return nil
	}

	updated := false
	for i, t := range trail {
		if t["status"] == "pending" {
			trail[i]["status"] = "cancelled"
			trail[i]["note"] = "Dibatalkan oleh pengaju"
			trail[i]["date"] = time.Now().UTC().Format(time.RFC3339)
			updated = true
		}
	}

	if updated {
		updatedTrail, _ := json.Marshal(trail)
		repository.UpdateApprovalTrail(ctx, entityType, entityID, string(updatedTrail))
	}

	return nil
}

// ─── Pending Approvals ─────────────────────────────────────────

func (s *ApprovalWorkflowService) GetPendingApprovals(ctx context.Context, userID string) (*models.PendingApprovalListResponse, error) {
	if userID == "" {
		return nil, errors.New("user tidak ditemukan")
	}

	items, err := repository.GetPendingApprovalsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat pending approvals: %w", err)
	}

	// Enrich items with titles
	for i, item := range items {
		title, desc, amount := s.getEntityInfo(ctx, item.EntityType, item.EntityID.String())
		items[i].Title = title
		items[i].Description = desc
		items[i].Amount = amount
	}

	return &models.PendingApprovalListResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ─── Helper Functions ──────────────────────────────────────────

func (s *ApprovalWorkflowService) getEntityLabel(entityType string) string {
	switch entityType {
	case "leave":
		return "Cuti"
	case "overtime":
		return "Lembur"
	case "reimbursement":
		return "Reimbursement"
	case "shift_change":
		return "Perubahan Shift"
	case "loan":
		return "Pinjaman"
	case "manual_attendance":
		return "Absensi Manual"
	case "resign":
		return "Resign"
	case "mutation":
		return "Mutasi"
	default:
		return entityType
	}
}

func (s *ApprovalWorkflowService) appendToApprovalTrail(ctx context.Context, entityType, entityID string, entry map[string]interface{}) {
	s.appendToApprovalTrailTx(ctx, nil, entityType, entityID, entry)
}

func (s *ApprovalWorkflowService) appendToApprovalTrailTx(ctx context.Context, tx pgx.Tx, entityType, entityID string, entry map[string]interface{}) {
	var currentTrail string
	var err error

	switch entityType {
	case "leave":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM leave_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM leave_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "overtime":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM overtime_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM overtime_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "reimbursement":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM reimbursements WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM reimbursements WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "shift_change":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM shift_change_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM shift_change_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "manual_attendance":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM manual_attendance_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM manual_attendance_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "resign":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM resign_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM resign_requests WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	case "mutation":
		if tx != nil {
			err = tx.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM employee_mutations WHERE id::text = $1`, entityID).Scan(&currentTrail)
		} else {
			err = database.Pool.QueryRow(ctx,
				`SELECT COALESCE(approval_trail, '[]') FROM employee_mutations WHERE id::text = $1`, entityID).Scan(&currentTrail)
		}
	default:
		return
	}

	if err != nil {
		return
	}

	// Parse and append
	var trail []map[string]interface{}
	json.Unmarshal([]byte(currentTrail), &trail)
	if trail == nil {
		trail = []map[string]interface{}{}
	}

	// Check if entry with same step already exists — update it
	found := false
	for i, t := range trail {
		if t["step"] == entry["step"] {
			trail[i] = entry
			found = true
			break
		}
	}
	if !found {
		trail = append(trail, entry)
	}

	updatedTrail, _ := json.Marshal(trail)
	if tx != nil {
		repository.UpdateApprovalTrailTx(ctx, tx, entityType, entityID, string(updatedTrail))
	} else {
		repository.UpdateApprovalTrail(ctx, entityType, entityID, string(updatedTrail))
	}
}

func (s *ApprovalWorkflowService) getEntityInfo(ctx context.Context, entityType, entityID string) (title, description string, amount float64) {
	switch entityType {
	case "leave":
		var reason, typeName string
		var totalDays float64
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(lr.reason, ''), COALESCE(lt.name, ''), COALESCE(lr.total_days, 0)::float
			FROM leave_requests lr
			LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
			WHERE lr.id::text = $1
		`, entityID).Scan(&reason, &typeName, &totalDays)
		if err != nil {
			return "Pengajuan Cuti", "", 0
		}
		title = typeName
		description = reason
		amount = totalDays

	case "overtime":
		var reason, otType string
		var hours float64
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(reason, ''), COALESCE(overtime_type, ''), COALESCE(total_hours, 0)
			FROM overtime_requests WHERE id::text = $1
		`, entityID).Scan(&reason, &otType, &hours)
		if err != nil {
			return "Pengajuan Lembur", "", 0
		}
		title = otType
		description = reason
		amount = hours

	case "reimbursement":
		var desc, rType string
		var amt float64
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(description, ''), COALESCE(type, ''), COALESCE(amount, 0)
			FROM reimbursements WHERE id::text = $1
		`, entityID).Scan(&desc, &rType, &amt)
		if err != nil {
			return "Pengajuan Reimbursement", "", 0
		}
		title = rType
		description = desc
		amount = amt

	case "shift_change":
		var reason string
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(reason, '') FROM shift_change_requests WHERE id::text = $1
		`, entityID).Scan(&reason)
		if err != nil {
			return "Perubahan Shift", "", 0
		}
		title = "Perubahan Shift"
		description = reason

	case "loan":
		var purpose, lType string
		var amt float64
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(purpose, ''), COALESCE(type, ''), COALESCE(amount, 0)
			FROM loans WHERE id::text = $1
		`, entityID).Scan(&purpose, &lType, &amt)
		if err != nil {
			return "Pengajuan Pinjaman", "", 0
		}
		title = lType
		description = purpose
		amount = amt

	case "manual_attendance":
		var reason, empID string
		var rDate time.Time
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(mar.reason, ''), COALESCE(e.full_name, ''), mar.date
			FROM manual_attendance_requests mar
			LEFT JOIN employees e ON e.id = mar.employee_id
			WHERE mar.id::text = $1
		`, entityID).Scan(&reason, &empID, &rDate)
		if err != nil {
			return "Absensi Manual", "", 0
		}
		title = "Absensi Manual - " + rDate.Format("2006-01-02")
		description = reason

	case "resign":
		var resignReason, resignType, empName string
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(rr.reason, ''), COALESCE(rr.resign_type, ''), COALESCE(e.full_name, '')
			FROM resign_requests rr
			LEFT JOIN employees e ON e.id = rr.employee_id
			WHERE rr.id::text = $1
		`, entityID).Scan(&resignReason, &resignType, &empName)
		if err != nil {
			return "Pengajuan Resign", "", 0
		}
		title = "Resign - " + empName
		description = resignReason

	case "mutation":
		var mReason, mType, mEmpName string
		err := database.Pool.QueryRow(ctx, `
			SELECT COALESCE(em.reason, ''), COALESCE(em.mutation_type, ''), COALESCE(e.full_name, '')
			FROM employee_mutations em
			LEFT JOIN employees e ON e.id = em.employee_id
			WHERE em.id::text = $1
		`, entityID).Scan(&mReason, &mType, &mEmpName)
		if err != nil {
			return "Pengajuan Mutasi", "", 0
		}
		title = "Mutasi - " + mEmpName
		description = mReason
	}
	return
}
