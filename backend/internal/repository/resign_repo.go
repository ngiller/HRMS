package repository

import (
	"context"
	"errors"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ==================== Resign Requests ====================

func CreateResignRequest(ctx context.Context, employeeID string, req *models.CreateResignRequest) (*models.ResignRequest, error) {
	var r models.ResignRequest
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO resign_requests (employee_id, resign_date, last_working_date, reason, resign_type)
		VALUES ($1::uuid, CURRENT_DATE, $2::date, $3, $4)
		RETURNING id, employee_id, resign_date, last_working_date, reason, resign_type, status, approved_by, approved_at, COALESCE(rejection_reason, '') AS rejection_reason, created_at, updated_at
	`, employeeID, req.LastWorkingDate, req.Reason, req.ResignType).Scan(
		&r.ID, &r.EmployeeID, &r.ResignDate, &r.LastWorkingDate, &r.Reason, &r.ResignType,
		&r.Status, &r.ApprovedBy, &r.ApprovedAt, &r.RejectionReason, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan resign: %w", err)
	}
	return &r, nil
}

func ListResignRequests(ctx context.Context, page, perPage int, status, employeeID string) (*models.ResignListResponse, error) {
	offset := (page - 1) * perPage
	args := []interface{}{}
	argIdx := 0
	conditions := []string{}

	if status != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("rr.status = $%d", argIdx))
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("rr.employee_id::text = $%d", argIdx))
		args = append(args, employeeID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM resign_requests rr %s`, whereClause)
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	argIdx++
	query := fmt.Sprintf(`
		SELECT rr.id, rr.employee_id, COALESCE(e.full_name, '') AS employee_name,
			rr.resign_date, rr.last_working_date, rr.reason, rr.resign_type,
			rr.status, rr.approved_by, COALESCE(ae.full_name, '') AS approved_by_name,
			rr.approved_at, COALESCE(rr.rejection_reason, '') AS rejection_reason,
			rr.created_at, rr.updated_at
		FROM resign_requests rr
		LEFT JOIN employees e ON e.id = rr.employee_id
		LEFT JOIN employees ae ON ae.id = rr.approved_by
		%s
		ORDER BY rr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resigns []models.ResignRequest
	for rows.Next() {
		var r models.ResignRequest
		if err := rows.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.ResignDate, &r.LastWorkingDate, &r.Reason, &r.ResignType,
			&r.Status, &r.ApprovedBy, &r.ApprovedByName,
			&r.ApprovedAt, &r.RejectionReason,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resigns = append(resigns, r)
	}
	if resigns == nil {
		resigns = []models.ResignRequest{}
	}

	return &models.ResignListResponse{
		Resigns: resigns,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func GetResignRequest(ctx context.Context, id string) (*models.ResignRequest, error) {
	var r models.ResignRequest
	err := database.Pool.QueryRow(ctx, `
		SELECT rr.id, rr.employee_id, COALESCE(e.full_name, '') AS employee_name,
			rr.resign_date, rr.last_working_date, rr.reason, rr.resign_type,
			rr.status, rr.approved_by, COALESCE(ae.full_name, '') AS approved_by_name,
			rr.approved_at, COALESCE(rr.rejection_reason, '') AS rejection_reason,
			rr.created_at, rr.updated_at
		FROM resign_requests rr
		LEFT JOIN employees e ON e.id = rr.employee_id
		LEFT JOIN employees ae ON ae.id = rr.approved_by
		WHERE rr.id::text = $1
	`, id).Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.ResignDate, &r.LastWorkingDate, &r.Reason, &r.ResignType,
		&r.Status, &r.ApprovedBy, &r.ApprovedByName,
		&r.ApprovedAt, &r.RejectionReason,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pengajuan resign tidak ditemukan")
		}
		return nil, err
	}
	return &r, nil
}

func UpdateResignRequestStatus(ctx context.Context, id, status, approvedBy, rejectionReason string) error {
	tag, err := database.Pool.Exec(ctx, `
		UPDATE resign_requests SET status = $2, approved_by = $3::uuid, approved_at = NOW(), rejection_reason = NULLIF($4, ''), updated_at = NOW()
		WHERE id::text = $1
	`, id, status, approvedBy, rejectionReason)
	if err != nil {
		return fmt.Errorf("gagal update status resign: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("pengajuan resign tidak ditemukan")
	}
	return nil
}

// ProcessEmployeeResignation handles the actual deactivation of an employee after resign is approved
func ProcessEmployeeResignation(ctx context.Context, resignID string) error {
	// Get the resign request
	var employeeID, lastWorkingDate string
	err := database.Pool.QueryRow(ctx, `SELECT employee_id::text, last_working_date::text FROM resign_requests WHERE id::text = $1`, resignID).Scan(&employeeID, &lastWorkingDate)
	if err != nil {
		return fmt.Errorf("gagal memuat data resign: %w", err)
	}

	// Soft-delete the employee
	tag, err := database.Pool.Exec(ctx, `
		UPDATE employees SET deleted_at = NOW(), is_active = FALSE, updated_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
	`, employeeID)
	if err != nil {
		return fmt.Errorf("gagal menonaktifkan karyawan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("karyawan tidak ditemukan")
	}

	return nil
}

// ==================== Exit Clearance Items ====================

// DefaultExitClearanceItems returns the standard clearance checklist
func DefaultExitClearanceItems() []struct {
	Name        string
	Description string
} {
	return []struct {
		Name        string
		Description string
	}{
		{"Pengembalian Aset Perusahaan", "Laptop, monitor, dan perlengkapan kantor"},
		{"Pengembalian Seragam/ID Card", "Seragam kerja dan kartu identitas karyawan"},
		{"Clearance Keuangan", "Pelunasan pinjaman, kasbon, atau tanggungan lainnya"},
		{"Clearance BPJS", "Pemutusan BPJS Kesehatan dan Ketenagakerjaan"},
		{"Exit Interview", "Wawancara keluar dengan HR"},
		{"Pengembalian Akses", "Akses email, sistem, dan aplikasi perusahaan"},
		{"Serah Terima Pekerjaan", "Dokumentasi dan transfer pengetahuan ke pengganti"},
		{"Penyelesaian Cuti", "Perhitungan sisa cuti yang belum diambil"},
		{"Penyelesaian Hak", "Perhitungan pesangon, hak pensiun, dan lain-lain"},
		{"Surat Keterangan Kerja", "Penerbitan surat referensi kerja"},
	}
}

func CreateExitClearanceItems(ctx context.Context, resignID string) error {
	items := DefaultExitClearanceItems()
	for i, item := range items {
		_, err := database.Pool.Exec(ctx, `
			INSERT INTO exit_clearance_items (resign_id, item_name, description, sort_order)
			VALUES ($1::uuid, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`, resignID, item.Name, item.Description, i+1)
		if err != nil {
			return fmt.Errorf("gagal membuat item clearance: %w", err)
		}
	}
	return nil
}

func ListExitClearanceItems(ctx context.Context, resignID string) ([]models.ExitClearanceItem, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT eci.id, eci.resign_id, eci.item_name, COALESCE(eci.description, '') AS description,
			eci.is_checked, eci.checked_by, COALESCE(ce.full_name, '') AS checked_by_name,
			eci.checked_at, eci.sort_order, eci.created_at, eci.updated_at
		FROM exit_clearance_items eci
		LEFT JOIN employees ce ON ce.id = eci.checked_by
		WHERE eci.resign_id::text = $1
		ORDER BY eci.sort_order ASC
	`, resignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ExitClearanceItem
	for rows.Next() {
		var item models.ExitClearanceItem
		if err := rows.Scan(
			&item.ID, &item.ResignID, &item.ItemName, &item.Description,
			&item.IsChecked, &item.CheckedBy, &item.CheckedByName,
			&item.CheckedAt, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.ExitClearanceItem{}
	}
	return items, nil
}

func UpdateExitClearanceItem(ctx context.Context, itemID, checkedBy string, isChecked bool) error {
	tag, err := database.Pool.Exec(ctx, `
		UPDATE exit_clearance_items SET is_checked = $2, checked_by = $3::uuid, checked_at = CASE WHEN $2 THEN NOW() ELSE NULL END, updated_at = NOW()
		WHERE id::text = $1
	`, itemID, isChecked, checkedBy)
	if err != nil {
		return fmt.Errorf("gagal update clearance item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("item clearance tidak ditemukan")
	}
	return nil
}

func GetResignIDByItemID(ctx context.Context, itemID string, resignID *string) error {
	return database.Pool.QueryRow(ctx, `SELECT resign_id::text FROM exit_clearance_items WHERE id::text = $1`, itemID).Scan(resignID)
}

func CheckAllClearanceItemsChecked(ctx context.Context, resignID string) (bool, error) {
	var count, checked int
	err := database.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN is_checked THEN 1 ELSE 0 END), 0)
		FROM exit_clearance_items WHERE resign_id::text = $1
	`, resignID).Scan(&count, &checked)
	if err != nil {
		return false, err
	}
	return count > 0 && count == checked, nil
}
