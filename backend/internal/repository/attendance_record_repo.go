package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetTodayAttendanceByEmployee(ctx context.Context, employeeID string) (*models.AttendanceRecord, error) {
	query := `
		SELECT ar.id, ar.employee_id, ar.date, ar.check_in_time, ar.check_out_time,
			ar.check_in_photo_url, ar.check_in_lat, ar.check_in_lng,
			ar.check_in_location_id, COALESCE(ar.check_in_location_name, ''),
			ar.check_out_photo_url, ar.check_out_lat, ar.check_out_lng,
			ar.check_out_location_id, COALESCE(ar.check_out_location_name, ''),
			ar.status, ar.is_late, ar.late_minutes, ar.is_early_leave, ar.early_leave_minutes,
			ar.total_work_hours, ar.is_manual_entry, ar.manual_entry_reason,
			ar.created_at, ar.updated_at
		FROM attendance_records ar
		WHERE ar.employee_id::text = $1 AND ar.date = CURRENT_DATE
	`
	row := database.Pool.QueryRow(ctx, query, employeeID)
	record, err := scanAttendanceRecord(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return record, nil
}

func CreateCheckIn(ctx context.Context, employeeID string, scheduleID string, lat, lng *float64, locationID, locationName *string, isLate bool, lateMinutes int, photoURL *string) (*models.AttendanceRecord, error) {
	var locID *uuid.UUID
	if locationID != nil && *locationID != "" {
		uid, err := uuid.Parse(*locationID)
		if err == nil {
			locID = &uid
		}
	}

	status := "hadir"
	if isLate {
		status = "terlambat"
	}

	query := `
		INSERT INTO attendance_records (employee_id, date, check_in_time,
			check_in_lat, check_in_lng, check_in_location_id, check_in_location_name,
			status, is_late, late_minutes, check_in_photo_url)
		VALUES ($1::uuid, CURRENT_DATE, NOW(),
			$2, $3, $4, $5,
			$6, $7, $8, $9)
		RETURNING id, employee_id, date, check_in_time, check_out_time,
			check_in_photo_url, check_in_lat, check_in_lng,
			check_in_location_id, COALESCE(check_in_location_name, ''),
			check_out_photo_url, check_out_lat, check_out_lng,
			check_out_location_id, COALESCE(check_out_location_name, ''),
			status, is_late, late_minutes, is_early_leave, early_leave_minutes,
			total_work_hours, is_manual_entry, manual_entry_reason,
			created_at, updated_at
	`
	row := database.Pool.QueryRow(ctx, query, employeeID, lat, lng, locID, locationName,
		status, isLate, lateMinutes, photoURL)
	return scanAttendanceRecord(row)
}

func UpdateCheckOut(ctx context.Context, recordID string, lat, lng *float64, locationID, locationName *string, totalWorkHours *float64, photoURL *string) (*models.AttendanceRecord, error) {
	var locID *uuid.UUID
	if locationID != nil && *locationID != "" {
		uid, err := uuid.Parse(*locationID)
		if err == nil {
			locID = &uid
		}
	}

	query := `
		UPDATE attendance_records SET
			check_out_time = NOW(),
			check_out_lat = $2,
			check_out_lng = $3,
			check_out_location_id = $4,
			check_out_location_name = $5,
			total_work_hours = $6,
			check_out_photo_url = COALESCE($7, check_out_photo_url)
		WHERE id::text = $1
		RETURNING id, employee_id, date, check_in_time, check_out_time,
			check_in_photo_url, check_in_lat, check_in_lng,
			check_in_location_id, COALESCE(check_in_location_name, ''),
			check_out_photo_url, check_out_lat, check_out_lng,
			check_out_location_id, COALESCE(check_out_location_name, ''),
			status, is_late, late_minutes, is_early_leave, early_leave_minutes,
			total_work_hours, is_manual_entry, manual_entry_reason,
			created_at, updated_at
	`
	row := database.Pool.QueryRow(ctx, query, recordID, lat, lng, locID, locationName, totalWorkHours, photoURL)
	return scanAttendanceRecord(row)
}

func ListMyAttendance(ctx context.Context, employeeID string, page, perPage int) ([]models.AttendanceRecordSummary, int, error) {
	offset := (page - 1) * perPage

	countQuery := `
		SELECT COUNT(*) FROM attendance_records ar
		WHERE ar.employee_id::text = $1
	`
	var total int
	err := database.Pool.QueryRow(ctx, countQuery, employeeID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT ar.id, ar.date, TO_CHAR(ar.date, 'Day') AS day_name,
			ar.check_in_time, ar.check_out_time,
			ar.status, ar.is_late, ar.late_minutes, ar.is_early_leave, ar.total_work_hours,
			COALESCE(ar.check_in_location_name, ''), ar.check_in_photo_url,
			ar.check_in_lat, ar.check_in_lng, COALESCE(ar.check_out_location_name, ''),
			ar.check_out_photo_url, ar.check_out_lat, ar.check_out_lng
		FROM attendance_records ar
		WHERE ar.employee_id::text = $1
		ORDER BY ar.date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := database.Pool.Query(ctx, query, employeeID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []models.AttendanceRecordSummary
	for rows.Next() {
		var r models.AttendanceRecordSummary
		if err := rows.Scan(&r.ID, &r.Date, &r.DayName, &r.CheckInTime, &r.CheckOutTime,
			&r.Status, &r.IsLate, &r.LateMinutes, &r.IsEarlyLeave, &r.TotalWorkHours, &r.CheckInLocationName,
			&r.CheckInPhotoURL, &r.CheckInLat, &r.CheckInLng, &r.CheckOutLocationName,
			&r.CheckOutPhotoURL, &r.CheckOutLat, &r.CheckOutLng); err != nil {
			return nil, 0, err
		}
		records = append(records, r)
	}
	if records == nil {
		records = []models.AttendanceRecordSummary{}
	}
	return records, total, nil
}

func ListAttendanceReport(ctx context.Context, page, perPage int, deptID, status, dateFrom, dateTo string) ([]models.AttendanceRecordSummary, int, error) {
	args := []interface{}{}
	argIdx := 0
	conditions := []string{}

	if deptID != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("e.department_id::text = $%d", argIdx))
		args = append(args, deptID)
	}
	if status != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("ar.status = $%d", argIdx))
		args = append(args, status)
	}
	if dateFrom != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("ar.date >= $%d", argIdx))
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf("ar.date <= $%d", argIdx))
		args = append(args, dateTo)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM attendance_records ar
		JOIN employees e ON e.id = ar.employee_id
		%s
	`, whereClause)
	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	argIdx++
	argIdx2 := argIdx + 1
	query := fmt.Sprintf(`
		SELECT ar.id, ar.date, TO_CHAR(ar.date, 'Day') AS day_name,
			ar.check_in_time, ar.check_out_time,
			ar.status, ar.is_late, ar.late_minutes, ar.is_early_leave, ar.total_work_hours,
			COALESCE(ar.check_in_location_name, ''), ar.check_in_photo_url,
			ar.check_in_lat, ar.check_in_lng, COALESCE(ar.check_out_location_name, ''),
			ar.check_out_photo_url, ar.check_out_lat, ar.check_out_lng,
			e.full_name, COALESCE(d.name, '')
		FROM attendance_records ar
		JOIN employees e ON e.id = ar.employee_id
		LEFT JOIN departments d ON d.id = e.department_id
		%s
		ORDER BY ar.date DESC, e.full_name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx2)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []models.AttendanceRecordSummary
	for rows.Next() {
		var r models.AttendanceRecordSummary
		if err := rows.Scan(&r.ID, &r.Date, &r.DayName, &r.CheckInTime, &r.CheckOutTime,
			&r.Status, &r.IsLate, &r.LateMinutes, &r.IsEarlyLeave, &r.TotalWorkHours, &r.CheckInLocationName,
			&r.CheckInPhotoURL, &r.CheckInLat, &r.CheckInLng, &r.CheckOutLocationName,
			&r.CheckOutPhotoURL, &r.CheckOutLat, &r.CheckOutLng,
			&r.EmployeeName, &r.DepartmentName); err != nil {
			return nil, 0, err
		}
		records = append(records, r)
	}
	if records == nil {
		records = []models.AttendanceRecordSummary{}
	}
	return records, total, nil
}

func GetActiveAttendanceLocations(ctx context.Context) ([]models.AttendanceLocationSummary, error) {
	query := `
		SELECT al.id, al.name, COALESCE(al.address, ''), al.latitude, al.longitude,
			al.radius_meters, al.is_active, al.created_at
		FROM attendance_locations al
		WHERE al.deleted_at IS NULL AND al.is_active = TRUE
		ORDER BY al.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.AttendanceLocationSummary
	for rows.Next() {
		var loc models.AttendanceLocationSummary
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Latitude, &loc.Longitude, &loc.RadiusMeters, &loc.IsActive, &loc.CreatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	if locations == nil {
		locations = []models.AttendanceLocationSummary{}
	}
	return locations, nil
}

func GetEmployeeScheduleInfo(ctx context.Context, employeeID string) (scheduleID *string, scheduleName string, startTime string, endTime string, err error) {
	// Step 1: Try the new flexible schedule system (employee_schedules + schedule_templates)
	date := time.Now().Format("2006-01-02")
	rs, resolveErr := ResolveEmployeeScheduleForDate(ctx, employeeID, date)
	if resolveErr == nil && rs != nil {
		sid := rs.ScheduleID.String()
		// Get schedule name from template or work_schedule
		var sname string
		switch rs.Source {
		case "date_override", "weekly_schedule":
			// Look up the employee schedule's template name
			database.Pool.QueryRow(ctx, `
				SELECT COALESCE(st.name, '') FROM employee_schedules es
				LEFT JOIN schedule_templates st ON st.id = es.template_id
				WHERE es.id = $1
			`, rs.ScheduleID).Scan(&sname)
			if sname == "" {
				sname = rs.Source
			}
		case "work_schedule":
			database.Pool.QueryRow(ctx, `SELECT COALESCE(ws.name, '') FROM employees e JOIN work_schedules ws ON ws.id = e.work_schedule_id WHERE e.id::text = $1`, employeeID).Scan(&sname)
		case "department_schedule":
			database.Pool.QueryRow(ctx, `
				SELECT COALESCE(ws.name, '') FROM employees e
				JOIN departments d ON d.id = e.department_id
				JOIN work_schedules ws ON ws.id = d.work_schedule_id
				WHERE e.id::text = $1
			`, employeeID).Scan(&sname)
		default:
			sname = rs.Source
		}
		if sname == "" {
			sname = "Jadwal Karyawan"
		}
		return &sid, sname, rs.StartTime, rs.EndTime, nil
	}

	// Step 2: Fallback to old work_schedules system
	query := `
		SELECT COALESCE(ws.id::text, ''), COALESCE(ws.name, ''),
			CASE
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 0 THEN COALESCE(ws.sunday_start::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 1 THEN COALESCE(ws.monday_start::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 2 THEN COALESCE(ws.tuesday_start::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 3 THEN COALESCE(ws.wednesday_start::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 4 THEN COALESCE(ws.thursday_start::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 5 THEN COALESCE(ws.friday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 6 THEN COALESCE(ws.saturday_start::text, '')
			END AS start_time,
			CASE
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 0 THEN COALESCE(ws.sunday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 1 THEN COALESCE(ws.monday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 2 THEN COALESCE(ws.tuesday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 3 THEN COALESCE(ws.wednesday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 4 THEN COALESCE(ws.thursday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 5 THEN COALESCE(ws.friday_end::text, '')
				WHEN EXTRACT(DOW FROM CURRENT_DATE) = 6 THEN COALESCE(ws.saturday_end::text, '')
			END AS end_time,
			COALESCE(ws.late_tolerance_minutes, 0)
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN work_schedules ws ON ws.id = COALESCE(e.work_schedule_id, d.work_schedule_id)
		WHERE e.id::text = $1
	`
	var sid, sname, stime, etime string
	var tolerance int
	row := database.Pool.QueryRow(ctx, query, employeeID)
	err = row.Scan(&sid, &sname, &stime, &etime, &tolerance)
	if err != nil {
		return nil, "", "", "", err
	}
	if sid == "" {
		return nil, "", "", "", nil
	}
	return &sid, sname, stime, etime, nil
}

func scanAttendanceRecord(row interface {
	Scan(dest ...interface{}) error
}) (*models.AttendanceRecord, error) {
	var r models.AttendanceRecord
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.Date, &r.CheckInTime, &r.CheckOutTime,
		&r.CheckInPhotoURL, &r.CheckInLat, &r.CheckInLng,
		&r.CheckInLocationID, &r.CheckInLocationName,
		&r.CheckOutPhotoURL, &r.CheckOutLat, &r.CheckOutLng,
		&r.CheckOutLocationID, &r.CheckOutLocationName,
		&r.Status, &r.IsLate, &r.LateMinutes, &r.IsEarlyLeave, &r.EarlyLeaveMinutes,
		&r.TotalWorkHours, &r.IsManualEntry, &r.ManualEntryReason,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
