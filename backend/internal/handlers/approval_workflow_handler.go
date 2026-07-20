package handlers

import (
	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ApprovalWorkflowHandler struct {
	svc *service.ApprovalWorkflowService
}

func NewApprovalWorkflowHandler(svc *service.ApprovalWorkflowService) *ApprovalWorkflowHandler {
	return &ApprovalWorkflowHandler{svc: svc}
}

// ─── Workflow Configuration ────────────────────────────────────

// ListWorkflows GET /api/approval-workflows?entity_type=leave
func (h *ApprovalWorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")
	workflows, err := h.svc.ListWorkflows(c.Context(), entityType)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse("Gagal memuat daftar workflow"))
	}
	return c.JSON(SuccessResponse(workflows, "Daftar workflow approval"))
}

// GetWorkflowDetail GET /api/approval-workflows/:id
func (h *ApprovalWorkflowHandler) GetWorkflowDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	detail, err := h.svc.GetWorkflowDetail(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse(err.Error()))
	}
	if detail == nil {
		return c.Status(404).JSON(ErrorResponse("Workflow tidak ditemukan"))
	}
	return c.JSON(SuccessResponse(detail, "Detail workflow approval"))
}

// CreateWorkflow POST /api/approval-workflows
func (h *ApprovalWorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	var req models.CreateApprovalWorkflowReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse("Format request tidak valid"))
	}
	workflow, err := h.svc.CreateWorkflow(c.Context(), &req)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	return c.JSON(SuccessResponse(workflow, "Workflow berhasil dibuat"))
}

// UpdateWorkflow PUT /api/approval-workflows/:id
func (h *ApprovalWorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UpdateApprovalWorkflowReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse("Format request tidak valid"))
	}
	workflow, err := h.svc.UpdateWorkflow(c.Context(), id, &req)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	return c.JSON(SuccessResponse(workflow, "Workflow berhasil diperbarui"))
}

// DeleteWorkflow DELETE /api/approval-workflows/:id
func (h *ApprovalWorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteWorkflow(c.Context(), id); err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	return c.JSON(SuccessResponse(nil, "Workflow berhasil dihapus"))
}

// ─── Workflow Steps ────────────────────────────────────────────

// AddWorkflowStep POST /api/approval-workflows/:id/steps
func (h *ApprovalWorkflowHandler) AddWorkflowStep(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	var req models.CreateApprovalWorkflowStepReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse("Format request tidak valid"))
	}
	req.WorkflowID = workflowID
	step, err := h.svc.AddWorkflowStep(c.Context(), &req)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	return c.JSON(SuccessResponse(step, "Step workflow berhasil ditambahkan"))
}

// UpdateWorkflowStep PUT /api/approval-workflow-steps/:id
func (h *ApprovalWorkflowHandler) UpdateWorkflowStep(c *fiber.Ctx) error {
	stepID := c.Params("id")
	var req models.UpdateApprovalWorkflowStepReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse("Format request tidak valid"))
	}
	step, err := h.svc.UpdateWorkflowStep(c.Context(), stepID, &req)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	if step == nil {
		return c.Status(404).JSON(ErrorResponse("Step tidak ditemukan"))
	}
	return c.JSON(SuccessResponse(step, "Step workflow berhasil diupdate"))
}

// DeleteWorkflowStep DELETE /api/approval-workflow-steps/:id
func (h *ApprovalWorkflowHandler) DeleteWorkflowStep(c *fiber.Ctx) error {
	stepID := c.Params("id")
	if err := h.svc.DeleteWorkflowStep(c.Context(), stepID); err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}
	return c.JSON(SuccessResponse(nil, "Step workflow berhasil dihapus"))
}

// ─── Pending Approvals ─────────────────────────────────────────

// GetPendingApprovals GET /api/approvals/pending
func (h *ApprovalWorkflowHandler) GetPendingApprovals(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(401).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	result, err := h.svc.GetPendingApprovals(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse(result, "Pending approvals"))
}

// ─── Process Approval ──────────────────────────────────────────

// ProcessApproval PUT /api/approvals/:entityType/:entityId/process
func (h *ApprovalWorkflowHandler) ProcessApproval(c *fiber.Ctx) error {
	userID := database.UserIDFromContext(c.Locals("user_id"))
	if userID == "" {
		return c.Status(401).JSON(ErrorResponse("User tidak terautentikasi"))
	}

	entityType := c.Params("entityType")
	entityID := c.Params("entityId")

	var req models.ApprovalActionReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse("Format request tidak valid"))
	}

	if req.Action != "approve" && req.Action != "reject" {
		return c.Status(400).JSON(ErrorResponse("Aksi harus 'approve' atau 'reject'"))
	}

	result, err := h.svc.ProcessApproval(c.Context(), entityType, entityID, userID, req.Action, req.Note)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse(result, result.Message))
}

// ─── Initialize Tracking for New Request ───────────────────────

// InitializeTracking POST /api/approvals/:entityType/:entityId/init
func (h *ApprovalWorkflowHandler) InitializeTracking(c *fiber.Ctx) error {
	entityType := c.Params("entityType")
	entityID := c.Params("entityId")

	// Get employee ID from entity
	var employeeID string
	switch entityType {
	case "leave":
		database.Pool.QueryRow(c.Context(),
			`SELECT employee_id::text FROM leave_requests WHERE id::text = $1`, entityID).Scan(&employeeID)
	case "overtime":
		database.Pool.QueryRow(c.Context(),
			`SELECT employee_id::text FROM overtime_requests WHERE id::text = $1`, entityID).Scan(&employeeID)
	case "reimbursement":
		database.Pool.QueryRow(c.Context(),
			`SELECT employee_id::text FROM reimbursements WHERE id::text = $1`, entityID).Scan(&employeeID)
	case "shift_change":
		database.Pool.QueryRow(c.Context(),
			`SELECT employee_id::text FROM shift_change_requests WHERE id::text = $1`, entityID).Scan(&employeeID)
	case "mutation":
		database.Pool.QueryRow(c.Context(),
			`SELECT employee_id::text FROM employee_mutations WHERE id::text = $1`, entityID).Scan(&employeeID)
	}

	if employeeID == "" {
		return c.Status(400).JSON(ErrorResponse("Entity atau karyawan tidak ditemukan"))
	}

	result, err := h.svc.ResolveWorkflowForRequest(c.Context(), entityType, entityID, employeeID)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse(result, result.Message))
}
