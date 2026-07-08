package repository

import (
	"context"
	"fmt"
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"
)

// MutationRepo handles database operations for employee mutations
type MutationRepo struct{}

// NewMutationRepo creates a new MutationRepo
func NewMutationRepo() *MutationRepo {
	return &MutationRepo{}
}

// List returns paginated mutations with employee info
func (r *MutationRepo) List(ctx context.Context, page, perPage int, status, employeeID string) ([]models.EmployeeMutation, int, error) {
	where := []string{"em.deleted_at IS NULL"}
	args := []interface{}{}
	idx := 0

	if status != "" {
		idx++
		where = append(where, fmt.Sprintf("em.status::text = $%d", idx))
		args = append(args, status)
	}
	if employeeID != "" {
		idx++
		where = append(where, fmt.Sprintf("em.employee_id::text = $%d", idx))
		args = append(args, employeeID)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := `SELECT COUNT(*) FROM employee_mutations em WHERE ` + whereClause
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count mutations: %w", err)
	}

	offset := (page - 1) * perPage
	idx++
	listQuery := fmt.Sprintf(`
		SELECT em.id::text, em.employee_id::text, COALESCE(e.full_name, ''),
			em.mutation_type,
			em.old_department_id::text, COALESCE(od.name, ''),
			em.old_position_id::text, COALESCE(op.name, ''),
			em.old_position_grade_id::text, COALESCE(opg.name, ''),
			em.old_employment_status, em.old_base_salary,
			em.new_department_id::text, COALESCE(nd.name, ''),
			em.new_position_id::text, COALESCE(np.name, ''),
			em.new_position_grade_id::text, COALESCE(npg.name, ''),
			em.new_employment_status, em.new_base_salary,
			COALESCE(em.reason, ''), COALESCE(em.document_url, ''),
			em.effective_date::text, COALESCE(em.notes, ''),
			em.status::text,
			em.approved_by::text, COALESCE(ab.full_name, ''),
			em.approved_at, COALESCE(em.rejection_reason, ''),
			em.created_by::text, COALESCE(cb.full_name, ''),
			em.created_at, em.updated_at
		FROM employee_mutations em
		LEFT JOIN employees e ON e.id = em.employee_id
		LEFT JOIN departments od ON od.id = em.old_department_id
		LEFT JOIN departments nd ON nd.id = em.new_department_id
		LEFT JOIN positions op ON op.id = em.old_position_id
		LEFT JOIN positions np ON np.id = em.new_position_id
		LEFT JOIN position_grades opg ON opg.id = em.old_position_grade_id
		LEFT JOIN position_grades npg ON npg.id = em.new_position_grade_id
		LEFT JOIN employees ab ON ab.id = em.approved_by
		LEFT JOIN employees cb ON cb.id = em.created_by
		WHERE %s
		ORDER BY em.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, idx, idx+1)

	allArgs := append(args, perPage, offset)
	rows, err := database.Pool.Query(ctx, listQuery, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list mutations: %w", err)
	}
	defer rows.Close()

	var mutations []models.EmployeeMutation
	for rows.Next() {
		var m models.EmployeeMutation
		if err := rows.Scan(
			&m.ID, &m.EmployeeID, &m.EmployeeName,
			&m.MutationType,
			&m.OldDepartmentID, &m.OldDepartmentName,
			&m.OldPositionID, &m.OldPositionName,
			&m.OldPositionGradeID, &m.OldPositionGradeName,
			&m.OldEmploymentStatus, &m.OldBaseSalary,
			&m.NewDepartmentID, &m.NewDepartmentName,
			&m.NewPositionID, &m.NewPositionName,
			&m.NewPositionGradeID, &m.NewPositionGradeName,
			&m.NewEmploymentStatus, &m.NewBaseSalary,
			&m.Reason, &m.DocumentURL,
			&m.EffectiveDate, &m.Notes,
			&m.Status,
			&m.ApprovedBy, &m.ApprovedByName,
			&m.ApprovedAt, &m.RejectionReason,
			&m.CreatedBy, &m.CreatedByName,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan mutation: %w", err)
		}
		mutations = append(mutations, m)
	}

	if mutations == nil {
		mutations = []models.EmployeeMutation{}
	}
	return mutations, total, nil
}

// GetByID returns a single mutation by ID
func (r *MutationRepo) GetByID(ctx context.Context, id string) (*models.EmployeeMutation, error) {
	query := `
		SELECT em.id::text, em.employee_id::text, COALESCE(e.full_name, ''),
			em.mutation_type,
			em.old_department_id::text, COALESCE(od.name, ''),
			em.old_position_id::text, COALESCE(op.name, ''),
			em.old_position_grade_id::text, COALESCE(opg.name, ''),
			em.old_employment_status, em.old_base_salary,
			em.new_department_id::text, COALESCE(nd.name, ''),
			em.new_position_id::text, COALESCE(np.name, ''),
			em.new_position_grade_id::text, COALESCE(npg.name, ''),
			em.new_employment_status, em.new_base_salary,
			COALESCE(em.reason, ''), COALESCE(em.document_url, ''),
			em.effective_date::text, COALESCE(em.notes, ''),
			em.status::text,
			em.approved_by::text, COALESCE(ab.full_name, ''),
			em.approved_at, COALESCE(em.rejection_reason, ''),
			em.created_by::text, COALESCE(cb.full_name, ''),
			em.created_at, em.updated_at
		FROM employee_mutations em
		LEFT JOIN employees e ON e.id = em.employee_id
		LEFT JOIN departments od ON od.id = em.old_department_id
		LEFT JOIN departments nd ON nd.id = em.new_department_id
		LEFT JOIN positions op ON op.id = em.old_position_id
		LEFT JOIN positions np ON np.id = em.new_position_id
		LEFT JOIN position_grades opg ON opg.id = em.old_position_grade_id
		LEFT JOIN position_grades npg ON npg.id = em.new_position_grade_id
		LEFT JOIN employees ab ON ab.id = em.approved_by
		LEFT JOIN employees cb ON cb.id = em.created_by
		WHERE em.id::text = $1 AND em.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var m models.EmployeeMutation
	err := row.Scan(
		&m.ID, &m.EmployeeID, &m.EmployeeName,
		&m.MutationType,
		&m.OldDepartmentID, &m.OldDepartmentName,
		&m.OldPositionID, &m.OldPositionName,
		&m.OldPositionGradeID, &m.OldPositionGradeName,
		&m.OldEmploymentStatus, &m.OldBaseSalary,
		&m.NewDepartmentID, &m.NewDepartmentName,
		&m.NewPositionID, &m.NewPositionName,
		&m.NewPositionGradeID, &m.NewPositionGradeName,
		&m.NewEmploymentStatus, &m.NewBaseSalary,
		&m.Reason, &m.DocumentURL,
		&m.EffectiveDate, &m.Notes,
		&m.Status,
		&m.ApprovedBy, &m.ApprovedByName,
		&m.ApprovedAt, &m.RejectionReason,
		&m.CreatedBy, &m.CreatedByName,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get mutation: %w", err)
	}
	return &m, nil
}

// Create inserts a new mutation request
func (r *MutationRepo) Create(ctx context.Context, req *models.CreateMutationRequest, creatorID string) (*models.EmployeeMutation, error) {
	query := `
		INSERT INTO employee_mutations (employee_id, mutation_type,
			new_department_id, new_position_id, new_position_grade_id,
			new_employment_status, new_base_salary,
			reason, document_url, effective_date, notes,
			created_by)
		VALUES ($1::uuid, $2,
			NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, NULLIF($5, '')::uuid,
			NULLIF($6, ''), $7,
			$8, NULLIF($9, ''), $10::date, NULLIF($11, ''),
			$12::uuid)
		RETURNING id::text, employee_id::text, mutation_type,
			NULL::text, ''::text, NULL::text, ''::text,
			NULL::text, ''::text, NULL::text,
			NULL::numeric,
			new_department_id::text, ''::text,
			new_position_id::text, ''::text,
			new_position_grade_id::text, ''::text,
			new_employment_status, new_base_salary,
			reason, COALESCE(document_url, ''),
			effective_date::text, COALESCE(notes, ''),
			status::text,
			NULL::text, ''::text, NULL::timestamptz, '',
			created_by::text, ''::text,
			created_at, updated_at
	`
	row := database.Pool.QueryRow(ctx, query,
		req.EmployeeID, req.MutationType,
		req.NewDepartmentID, req.NewPositionID, req.NewPositionGradeID,
		req.NewEmploymentStatus, req.NewBaseSalary,
		req.Reason, req.DocumentURL, req.EffectiveDate, req.Notes,
		creatorID,
	)

	var m models.EmployeeMutation
	err := row.Scan(
		&m.ID, &m.EmployeeID, &m.MutationType,
		&m.OldDepartmentID, &m.OldDepartmentName,
		&m.OldPositionID, &m.OldPositionName,
		&m.OldPositionGradeID, &m.OldPositionGradeName,
		&m.OldEmploymentStatus, &m.OldBaseSalary,
		&m.NewDepartmentID, &m.NewDepartmentName,
		&m.NewPositionID, &m.NewPositionName,
		&m.NewPositionGradeID, &m.NewPositionGradeName,
		&m.NewEmploymentStatus, &m.NewBaseSalary,
		&m.Reason, &m.DocumentURL,
		&m.EffectiveDate, &m.Notes,
		&m.Status,
		&m.ApprovedBy, &m.ApprovedByName, &m.ApprovedAt, &m.RejectionReason,
		&m.CreatedBy, &m.CreatedByName,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create mutation: %w", err)
	}
	return &m, nil
}

// UpdateStatus updates the mutation status (approve/reject)
func (r *MutationRepo) UpdateStatus(ctx context.Context, id, status, approvedBy, rejectionReason string) error {
	query := `
		UPDATE employee_mutations
		SET status = $2::text,
			approved_by = CASE WHEN $2 IN ('approved', 'rejected') THEN $3::uuid ELSE NULL END,
			approved_at = CASE WHEN $2 IN ('approved', 'rejected') THEN NOW() ELSE NULL END,
			rejection_reason = NULLIF($4, ''),
			updated_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, id, status, approvedBy, rejectionReason)
	if err != nil {
		return fmt.Errorf("update mutation status: %w", err)
	}
	return nil
}

// nullIfEmpty returns nil if the string is empty, useful for nullable UUID fields
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// UpdateOldValues sets the old values from employee data (called after creation)
func (r *MutationRepo) UpdateOldValues(ctx context.Context, id, deptID, posID, gradeID, empStatus string, baseSalary *float64) error {
	query := `
		UPDATE employee_mutations
		SET old_department_id = $2::uuid,
			old_position_id = $3::uuid,
			old_position_grade_id = $4::uuid,
			old_employment_status = $5,
			old_base_salary = $6,
			updated_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, id,
		nullIfEmpty(deptID), nullIfEmpty(posID), nullIfEmpty(gradeID), empStatus, baseSalary)
	if err != nil {
		return fmt.Errorf("update old values: %w", err)
	}
	return nil
}

// ApplyMutation applies the mutation changes to the employee record
func (r *MutationRepo) ApplyMutation(ctx context.Context, id string) error {
	query := `
		UPDATE employees e
		SET
			department_id = COALESCE(em.new_department_id, e.department_id),
			position_id = COALESCE(em.new_position_id, e.position_id),
			position_grade_id = COALESCE(em.new_position_grade_id, e.position_grade_id),
			employment_status = COALESCE(em.new_employment_status::employment_status, e.employment_status),
			base_salary = COALESCE(em.new_base_salary, e.base_salary)
		FROM employee_mutations em
		WHERE em.id::text = $1 AND e.id = em.employee_id AND em.deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("apply mutation: %w", err)
	}
	return nil
}

// GetEmployeeData returns the current employee data needed for mutation
func (r *MutationRepo) GetEmployeeData(ctx context.Context, employeeID string) (deptID, posID, gradeID, empStatus string, baseSalary *float64, err error) {
	query := `
		SELECT e.department_id::text, e.position_id::text, COALESCE(e.position_grade_id::text, ''),
			e.employment_status, e.base_salary
		FROM employees e
		WHERE e.id::text = $1 AND e.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, employeeID)
	err = row.Scan(&deptID, &posID, &gradeID, &empStatus, &baseSalary)
	if err != nil {
		err = fmt.Errorf("get employee data: %w", err)
	}
	return
}

// ListAll returns all mutations without pagination (for export)
func (r *MutationRepo) ListAll(ctx context.Context, status, employeeID string) ([]models.EmployeeMutation, error) {
	where := []string{"em.deleted_at IS NULL"}
	args := []interface{}{}
	idx := 0

	if status != "" {
		idx++
		where = append(where, fmt.Sprintf("em.status::text = $%d", idx))
		args = append(args, status)
	}
	if employeeID != "" {
		idx++
		where = append(where, fmt.Sprintf("em.employee_id::text = $%d", idx))
		args = append(args, employeeID)
	}

	whereClause := strings.Join(where, " AND ")

	query := fmt.Sprintf(`
		SELECT em.id::text, em.employee_id::text, COALESCE(e.full_name, ''),
			em.mutation_type,
			em.old_department_id::text, COALESCE(od.name, ''),
			em.old_position_id::text, COALESCE(op.name, ''),
			em.old_position_grade_id::text, COALESCE(opg.name, ''),
			em.old_employment_status, em.old_base_salary,
			em.new_department_id::text, COALESCE(nd.name, ''),
			em.new_position_id::text, COALESCE(np.name, ''),
			em.new_position_grade_id::text, COALESCE(npg.name, ''),
			em.new_employment_status, em.new_base_salary,
			COALESCE(em.reason, ''), COALESCE(em.document_url, ''),
			em.effective_date::text, COALESCE(em.notes, ''),
			em.status::text,
			em.approved_by::text, COALESCE(ab.full_name, ''),
			em.approved_at, COALESCE(em.rejection_reason, ''),
			em.created_by::text, COALESCE(cb.full_name, ''),
			em.created_at, em.updated_at
		FROM employee_mutations em
		LEFT JOIN employees e ON e.id = em.employee_id
		LEFT JOIN departments od ON od.id = em.old_department_id
		LEFT JOIN departments nd ON nd.id = em.new_department_id
		LEFT JOIN positions op ON op.id = em.old_position_id
		LEFT JOIN positions np ON np.id = em.new_position_id
		LEFT JOIN position_grades opg ON opg.id = em.old_position_grade_id
		LEFT JOIN position_grades npg ON npg.id = em.new_position_grade_id
		LEFT JOIN employees ab ON ab.id = em.approved_by
		LEFT JOIN employees cb ON cb.id = em.created_by
		WHERE %s
		ORDER BY em.created_at DESC
	`, whereClause)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all mutations: %w", err)
	}
	defer rows.Close()

	var mutations []models.EmployeeMutation
	for rows.Next() {
		var m models.EmployeeMutation
		if err := rows.Scan(
			&m.ID, &m.EmployeeID, &m.EmployeeName,
			&m.MutationType,
			&m.OldDepartmentID, &m.OldDepartmentName,
			&m.OldPositionID, &m.OldPositionName,
			&m.OldPositionGradeID, &m.OldPositionGradeName,
			&m.OldEmploymentStatus, &m.OldBaseSalary,
			&m.NewDepartmentID, &m.NewDepartmentName,
			&m.NewPositionID, &m.NewPositionName,
			&m.NewPositionGradeID, &m.NewPositionGradeName,
			&m.NewEmploymentStatus, &m.NewBaseSalary,
			&m.Reason, &m.DocumentURL,
			&m.EffectiveDate, &m.Notes,
			&m.Status,
			&m.ApprovedBy, &m.ApprovedByName,
			&m.ApprovedAt, &m.RejectionReason,
			&m.CreatedBy, &m.CreatedByName,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mutation for export: %w", err)
		}
		mutations = append(mutations, m)
	}

	if mutations == nil {
		mutations = []models.EmployeeMutation{}
	}
	return mutations, nil
}

// UpdateApprovalTrail updates the approval trail JSON
func (r *MutationRepo) UpdateApprovalTrail(ctx context.Context, id, trail string) error {
	_, err := database.Pool.Exec(ctx,
		`UPDATE employee_mutations SET approval_trail = $2::jsonb, updated_at = NOW() WHERE id::text = $1`,
		id, trail,
	)
	return err
}
