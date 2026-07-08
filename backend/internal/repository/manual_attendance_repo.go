package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ==================== Manual Attendance Requests ====================

func CheckManualAttendanceCount(ctx context.Context, employeeID string, dateStr string) error {
	// Check company config for max manual attendance per month
	var maxPerMonth int
	err := database.Pool.QueryRow(ctx, `
		SELECT COALESCE((approval_config->>'manual_attendance_max_per_month')::int, 3)
		FROM companies WHERE is_active = TRUE LIMIT 1
	`).Scan(&maxPerMonth)
	if err != nil {
		maxPerMonth = 3
	}

	// Count manual attendance requests this month
	var count int
	database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM manual_attendance_requests
		WHERE employee_id::text = $1
		AND EXTRACT(YEAR FROM date) = EXTRACT(YEAR FROM $2::date)
		AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM $2::date)
		AND status != 'cancelled'
	`, employeeID, dateStr).Scan(&count)

	if count >= maxPerMonth {
		return fmt.Errorf("kuota pengajuan absensi manual bulan ini sudah habis (maksimal %d kali)", maxPerMonth)
	}
	return nil
}

func CreateManualAttendanceRequest(ctx context.Context, employeeID string, req *models.CreateManualAttendanceRequest) (*models.ManualAttendanceRequest, error) {
	var checkInTime, checkOutTime *time.Time
	if req.CheckInTime != "" {
		t, err := time.Parse("15:04", req.CheckInTime)
		if err == nil {
			parsed := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), t.Hour(), t.Minute(), 0, 0, time.Now().Location())
			checkInTime = &parsed
		}
	}
	if req.CheckOutTime != "" {
		t, err := time.Parse("15:04", req.CheckOutTime)
		if err == nil {
			parsed := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), t.Hour(), t.Minute(), 0, 0, time.Now().Location())
			checkOutTime = &parsed
		}
	}

	var r models.ManualAttendanceRequest
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO manual_attendance_requests (employee_id, date, check_in_time, check_out_time, reason)
		VALUES ($1::uuid, $2::date, $3, $4, $5)
		RETURNING id, employee_id, date, check_in_time, check_out_time, reason, status, approved_by, approved_at, COALESCE(rejection_reason, '') AS rejection_reason, created_at, updated_at
	`, employeeID, req.Date, checkInTime, checkOutTime, req.Reason).Scan(
		&r.ID, &r.EmployeeID, &r.Date, &r.CheckInTime, &r.CheckOutTime,
		&r.Reason, &r.Status, &r.ApprovedBy, &r.ApprovedAt, &r.RejectionReason,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan absensi manual: %w", err)
	}
	return &r, nil
}

func ListManualAttendanceRequests(ctx context.Context, page, perPage int, status, employeeID string) (*models.ManualAttendanceListResponse, error) {
	offset := (page - 1) * perPage
	args := []interface{}{}
	argIdx := 0
	conditions := []string{}

	if status != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("mar.status = $%d", argIdx))
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("mar.employee_id::text = $%d", argIdx))
		args = append(args, employeeID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM manual_attendance_requests mar
		%s
	`, whereClause)
	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Fetch
	argIdx++
	query := fmt.Sprintf(`
		SELECT mar.id, mar.employee_id, COALESCE(e.full_name, '') AS employee_name,
			mar.date, mar.check_in_time, mar.check_out_time, mar.reason,
			mar.status, mar.approved_by, COALESCE(ae.full_name, '') AS approved_by_name,
			mar.approved_at, COALESCE(mar.rejection_reason, '') AS rejection_reason,
			mar.created_at, mar.updated_at
		FROM manual_attendance_requests mar
		LEFT JOIN employees e ON e.id = mar.employee_id
		LEFT JOIN employees ae ON ae.id = mar.approved_by
		%s
		ORDER BY mar.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.ManualAttendanceRequest
	for rows.Next() {
		var r models.ManualAttendanceRequest
		if err := rows.Scan(
			&r.ID, &r.EmployeeID, &r.EmployeeName,
			&r.Date, &r.CheckInTime, &r.CheckOutTime, &r.Reason,
			&r.Status, &r.ApprovedBy, &r.ApprovedByName,
			&r.ApprovedAt, &r.RejectionReason,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	if requests == nil {
		requests = []models.ManualAttendanceRequest{}
	}

	return &models.ManualAttendanceListResponse{
		Requests: requests,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}, nil
}

func GetManualAttendanceRequest(ctx context.Context, id string) (*models.ManualAttendanceRequest, error) {
	var r models.ManualAttendanceRequest
	err := database.Pool.QueryRow(ctx, `
		SELECT mar.id, mar.employee_id, COALESCE(e.full_name, '') AS employee_name,
			mar.date, mar.check_in_time, mar.check_out_time, mar.reason,
			mar.status, mar.approved_by, COALESCE(ae.full_name, '') AS approved_by_name,
			mar.approved_at, COALESCE(mar.rejection_reason, '') AS rejection_reason,
			mar.created_at, mar.updated_at
		FROM manual_attendance_requests mar
		LEFT JOIN employees e ON e.id = mar.employee_id
		LEFT JOIN employees ae ON ae.id = mar.approved_by
		WHERE mar.id::text = $1
	`, id).Scan(
		&r.ID, &r.EmployeeID, &r.EmployeeName,
		&r.Date, &r.CheckInTime, &r.CheckOutTime, &r.Reason,
		&r.Status, &r.ApprovedBy, &r.ApprovedByName,
		&r.ApprovedAt, &r.RejectionReason,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pengajuan absensi manual tidak ditemukan")
		}
		return nil, err
	}
	return &r, nil
}

func UpdateManualAttendanceRequestStatus(ctx context.Context, id, status, approvedBy, rejectionReason string) error {
	_, err := database.Pool.Exec(ctx, `
		UPDATE manual_attendance_requests SET status = $2, approved_by = $3::uuid, approved_at = NOW(), rejection_reason = NULLIF($4, ''), updated_at = NOW()
		WHERE id::text = $1
	`, id, status, approvedBy, rejectionReason)
	if err != nil {
		return fmt.Errorf("gagal update status: %w", err)
	}
	return nil
}

// CreateAttendanceFromManualRequest creates an actual attendance record from an approved manual request
func CreateAttendanceFromManualRequest(ctx context.Context, req *models.ManualAttendanceRequest) error {
	status := "hadir"

	_, err := database.Pool.Exec(ctx, `
		INSERT INTO attendance_records (employee_id, date, check_in_time, check_out_time, status, is_manual_entry, manual_entry_reason)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6)
		ON CONFLICT (employee_id, date) DO UPDATE SET
			check_in_time = COALESCE($3, attendance_records.check_in_time),
			check_out_time = COALESCE($4, attendance_records.check_out_time),
			is_manual_entry = TRUE,
			manual_entry_reason = $6,
			status = CASE WHEN attendance_records.status = 'tanpa_keterangan' THEN $5 ELSE attendance_records.status END,
			updated_at = NOW()
	`, req.EmployeeID, req.Date, req.CheckInTime, req.CheckOutTime, status, req.Reason)
	if err != nil {
		return fmt.Errorf("gagal membuat record absensi: %w", err)
	}
	return nil
}
