package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================
// SCHEDULE TEMPLATES
// ============================================================

func ListScheduleTemplates(ctx context.Context, page, perPage int, search string) ([]models.ScheduleTemplateSummary, int, error) {
	var total int
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM schedule_templates WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT st.id, st.name, COALESCE(st.description, '') as description,
			st.schedule_type, st.is_active,
			(SELECT COUNT(*) FROM schedule_template_days std WHERE std.template_id = st.id) as day_count,
			st.created_at
		FROM schedule_templates st
		WHERE st.deleted_at IS NULL
		ORDER BY st.name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := database.Pool.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var templates []models.ScheduleTemplateSummary
	for rows.Next() {
		var t models.ScheduleTemplateSummary
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.ScheduleType, &t.IsActive, &t.DayCount, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}

func GetAllScheduleTemplates(ctx context.Context) ([]models.ScheduleTemplateSummary, error) {
	query := `
		SELECT st.id, st.name, COALESCE(st.description, ''), st.schedule_type, st.is_active,
			(SELECT COUNT(*) FROM schedule_template_days std WHERE std.template_id = st.id) as day_count,
			st.created_at
		FROM schedule_templates st WHERE st.deleted_at IS NULL AND st.is_active = TRUE
		ORDER BY st.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.ScheduleTemplateSummary
	for rows.Next() {
		var t models.ScheduleTemplateSummary
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.ScheduleType, &t.IsActive, &t.DayCount, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func GetScheduleTemplateByID(ctx context.Context, id string) (*models.ScheduleTemplate, error) {
	var t models.ScheduleTemplate
	err := database.Pool.QueryRow(ctx, `
		SELECT id, name, COALESCE(description, ''), schedule_type, is_active, created_at, updated_at, deleted_at
		FROM schedule_templates WHERE id::text = $1 AND deleted_at IS NULL
	`, id).Scan(&t.ID, &t.Name, &t.Description, &t.ScheduleType, &t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Load days
	rows, err := database.Pool.Query(ctx, `
		SELECT id, template_id, day_of_week, start_time::text, end_time::text,
			break_start::text, break_end::text, late_tolerance_minutes, early_leave_tolerance, is_active, sort_order
		FROM schedule_template_days
		WHERE template_id = $1 AND is_active = TRUE
		ORDER BY sort_order, day_of_week NULLS LAST
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d models.ScheduleTemplateDay
		if err := rows.Scan(&d.ID, &d.TemplateID, &d.DayOfWeek, &d.StartTime, &d.EndTime,
			&d.BreakStart, &d.BreakEnd, &d.LateToleranceMinutes, &d.EarlyLeaveTolerance, &d.IsActive, &d.SortOrder); err != nil {
			return nil, err
		}
		t.Days = append(t.Days, d)
	}
	if t.Days == nil {
		t.Days = []models.ScheduleTemplateDay{}
	}

	return &t, nil
}

func CreateScheduleTemplate(ctx context.Context, req *models.CreateScheduleTemplateRequest) (*models.ScheduleTemplate, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var t models.ScheduleTemplate
	err = tx.QueryRow(ctx, `
		INSERT INTO schedule_templates (name, description, schedule_type)
		VALUES ($1, $2, $3)
		RETURNING id, name, COALESCE(description, ''), schedule_type, is_active, created_at, updated_at, deleted_at
	`, req.Name, req.Description, req.ScheduleType).Scan(&t.ID, &t.Name, &t.Description, &t.ScheduleType, &t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Insert days
	for _, d := range req.Days {
		var day models.ScheduleTemplateDay
		err := tx.QueryRow(ctx, `
			INSERT INTO schedule_template_days (template_id, day_of_week, start_time, end_time, break_start, break_end, late_tolerance_minutes, early_leave_tolerance)
			VALUES ($1, $2, $3::time, $4::time, $5::time, $6::time, $7, $8)
			RETURNING id, template_id, day_of_week, start_time::text, end_time::text, break_start::text, break_end::text, late_tolerance_minutes, early_leave_tolerance, is_active, sort_order
		`, t.ID, d.DayOfWeek, d.StartTime, d.EndTime,
			d.BreakStart, d.BreakEnd, d.LateToleranceMinutes, d.EarlyLeaveTolerance).Scan(&day.ID, &day.TemplateID, &day.DayOfWeek, &day.StartTime, &day.EndTime,
			&day.BreakStart, &day.BreakEnd, &day.LateToleranceMinutes, &day.EarlyLeaveTolerance, &day.IsActive, &day.SortOrder)
		if err != nil {
			return nil, err
		}
		t.Days = append(t.Days, day)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if t.Days == nil {
		t.Days = []models.ScheduleTemplateDay{}
	}
	return &t, nil
}

func UpdateScheduleTemplate(ctx context.Context, id string, req *models.UpdateScheduleTemplateRequest) (*models.ScheduleTemplate, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update template fields
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
	}
	if req.ScheduleType != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("schedule_type = $%d", argIdx))
		args = append(args, *req.ScheduleType)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}

	if len(setClauses) > 0 {
		argIdx++
		query := fmt.Sprintf(`UPDATE schedule_templates SET %s WHERE id::text = $%d AND deleted_at IS NULL`, strings.Join(setClauses, ", "), argIdx)
		args = append(args, id)
		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	// Replace days if provided
	if req.Days != nil {
		_, err := tx.Exec(ctx, `DELETE FROM schedule_template_days WHERE template_id::text = $1`, id)
		if err != nil {
			return nil, err
		}
		for _, d := range req.Days {
			_, err := tx.Exec(ctx, `
			INSERT INTO schedule_template_days (template_id, day_of_week, start_time, end_time, break_start, break_end, late_tolerance_minutes, early_leave_tolerance)
			VALUES ($1::uuid, $2, NULLIF($3, '')::time, NULLIF($4, '')::time, NULLIF($5, '')::time, NULLIF($6, '')::time, $7, $8)
		`, id, d.DayOfWeek, d.StartTime, d.EndTime,
				d.BreakStart, d.BreakEnd, d.LateToleranceMinutes, d.EarlyLeaveTolerance)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return GetScheduleTemplateByID(ctx, id)
}

func DeleteScheduleTemplate(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE schedule_templates SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckScheduleTemplateNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM schedule_templates WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// ============================================================
// EMPLOYEE SCHEDULES (Level 2 + Level 3)
// ============================================================

func ListEmployeeSchedules(ctx context.Context, employeeID string, page, perPage int) ([]models.EmployeeScheduleSummary, int, error) {
	whereClause := ""
	args := []interface{}{}
	if employeeID != "" {
		whereClause = " WHERE es.employee_id::text = $1"
		args = append(args, employeeID)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM employee_schedules es` + whereClause
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT es.id, es.employee_id, e.full_name,
			COALESCE(st.name, '') as template_name,
			es.day_of_week, es.specific_date::text,
			es.start_time::text, es.end_time::text,
			COALESCE(es.is_remote, FALSE),
			es.effective_from::text, es.effective_until::text,
			es.priority, es.is_active,
			COALESCE((SELECT string_agg(al.name, ', ') FROM employee_schedule_locations esl
				JOIN attendance_locations al ON al.id = esl.attendance_location_id
				WHERE esl.employee_schedule_id = es.id), '') as location_names,
			es.created_at
		FROM employee_schedules es
		JOIN employees e ON e.id = es.employee_id
		LEFT JOIN schedule_templates st ON st.id = es.template_id
		%s
		ORDER BY es.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schedules []models.EmployeeScheduleSummary
	for rows.Next() {
		var s models.EmployeeScheduleSummary
		if err := rows.Scan(&s.ID, &s.EmployeeID, &s.EmployeeName, &s.TemplateName,
			&s.DayOfWeek, &s.SpecificDate, &s.StartTime, &s.EndTime,
			&s.IsRemote, &s.EffectiveFrom, &s.EffectiveUntil,
			&s.Priority, &s.IsActive, &s.LocationNames, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, s)
	}
	return schedules, total, nil
}

func GetEmployeeScheduleByID(ctx context.Context, id string) (*models.EmployeeSchedule, error) {
	var s models.EmployeeSchedule
	err := database.Pool.QueryRow(ctx, `
		SELECT es.id, es.employee_id, es.template_id, COALESCE(st.name, ''),
			es.day_of_week, es.specific_date::text,
			es.start_time::text, es.end_time::text, es.break_start::text, es.break_end::text,
			COALESCE(es.is_remote, FALSE),
			es.effective_from::text, es.effective_until::text, es.priority,
			COALESCE(es.reason, ''), es.is_active,
			es.created_at, es.updated_at
		FROM employee_schedules es
		LEFT JOIN schedule_templates st ON st.id = es.template_id
		WHERE es.id::text = $1
	`, id).Scan(&s.ID, &s.EmployeeID, &s.TemplateID, &s.TemplateName,
		&s.DayOfWeek, &s.SpecificDate, &s.StartTime, &s.EndTime, &s.BreakStart, &s.BreakEnd,
		&s.IsRemote, &s.EffectiveFrom, &s.EffectiveUntil, &s.Priority,
		&s.Reason, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Load locations
	rows, err := database.Pool.Query(ctx, `
		SELECT esl.id, esl.employee_schedule_id, esl.attendance_location_id,
			COALESCE(al.name, '') as location_name,
			esl.day_of_week, esl.is_primary, esl.sort_order
		FROM employee_schedule_locations esl
		LEFT JOIN attendance_locations al ON al.id = esl.attendance_location_id
		WHERE esl.employee_schedule_id = $1
		ORDER BY esl.sort_order, esl.day_of_week NULLS LAST
	`, s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var loc models.EmployeeScheduleLocation
		if err := rows.Scan(&loc.ID, &loc.EmployeeScheduleID, &loc.AttendanceLocationID,
			&loc.LocationName, &loc.DayOfWeek, &loc.IsPrimary, &loc.SortOrder); err != nil {
			return nil, err
		}
		s.Locations = append(s.Locations, loc)
	}
	if s.Locations == nil {
		s.Locations = []models.EmployeeScheduleLocation{}
	}

	return &s, nil
}

func CreateEmployeeSchedule(ctx context.Context, req *models.CreateEmployeeScheduleRequest, userID string) (*models.EmployeeSchedule, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	validateAndSet := func(val *string, defaultVal string) string {
		if val != nil && *val != "" {
			return *val
		}
		return defaultVal
	}

	effectiveFrom := validateAndSet(&req.EffectiveFrom, "CURRENT_DATE")
	var effectiveUntil interface{}
	if req.EffectiveUntil != nil && *req.EffectiveUntil != "" {
		effectiveUntil = *req.EffectiveUntil
	}

	isRemote := false
	if req.IsRemote != nil {
		isRemote = *req.IsRemote
	}

	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}

	var templateID interface{}
	if req.TemplateID != "" {
		templateID = req.TemplateID
	}

	var s models.EmployeeSchedule
	err = tx.QueryRow(ctx, `
		INSERT INTO employee_schedules (employee_id, template_id, day_of_week, specific_date,
			start_time, end_time, break_start, break_end, is_remote,
			effective_from, effective_until, priority, reason, created_by)
		VALUES ($1::uuid, NULLIF($2, '')::uuid, $3, NULLIF($4, '')::date,
			NULLIF($5, '')::time, NULLIF($6, '')::time, NULLIF($7, '')::time, NULLIF($8, '')::time, $9,
			$10::date, NULLIF($11, '')::date, $12, $13, NULLIF($14, '')::uuid)
		RETURNING id, employee_id, template_id,
			day_of_week, specific_date::text,
			start_time::text, end_time::text, break_start::text, break_end::text,
			is_remote, effective_from::text, effective_until::text, priority,
			COALESCE(reason, ''), is_active, created_at, updated_at
	`, req.EmployeeID, templateID, req.DayOfWeek, req.SpecificDate,
		req.StartTime, req.EndTime, req.BreakStart, req.BreakEnd, isRemote,
		effectiveFrom, effectiveUntil, priority, req.Reason, userID,
	).Scan(&s.ID, &s.EmployeeID, &s.TemplateID,
		&s.DayOfWeek, &s.SpecificDate, &s.StartTime, &s.EndTime, &s.BreakStart, &s.BreakEnd,
		&s.IsRemote, &s.EffectiveFrom, &s.EffectiveUntil, &s.Priority,
		&s.Reason, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if isPGUniqueViolation(err) {
			return nil, fmt.Errorf("karyawan sudah memiliki jadwal di tanggal tersebut")
		}
		return nil, err
	}

	// Save locations
	for _, loc := range req.Locations {
		isPrimary := true
		if loc.IsPrimary != nil {
			isPrimary = *loc.IsPrimary
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO employee_schedule_locations (employee_schedule_id, attendance_location_id, day_of_week, is_primary)
			VALUES ($1, $2::uuid, $3, $4)
		`, s.ID, loc.AttendanceLocationID, loc.DayOfWeek, isPrimary)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return GetEmployeeScheduleByID(ctx, s.ID.String())
}

func UpdateEmployeeSchedule(ctx context.Context, id string, req *models.UpdateEmployeeScheduleRequest, userID string) (*models.EmployeeSchedule, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.TemplateID != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("template_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, *req.TemplateID)
	}
	if req.StartTime != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("start_time = NULLIF($%d, '')::time", argIdx))
		args = append(args, *req.StartTime)
	}
	if req.EndTime != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("end_time = NULLIF($%d, '')::time", argIdx))
		args = append(args, *req.EndTime)
	}
	if req.BreakStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("break_start = NULLIF($%d, '')::time", argIdx))
		args = append(args, *req.BreakStart)
	}
	if req.BreakEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("break_end = NULLIF($%d, '')::time", argIdx))
		args = append(args, *req.BreakEnd)
	}
	if req.IsRemote != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_remote = $%d", argIdx))
		args = append(args, *req.IsRemote)
	}
	if req.EffectiveFrom != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("effective_from = $%d::date", argIdx))
		args = append(args, *req.EffectiveFrom)
	}
	if req.EffectiveUntil != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("effective_until = NULLIF($%d, '')::date", argIdx))
		args = append(args, *req.EffectiveUntil)
	}
	if req.Priority != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *req.Priority)
	}
	if req.Reason != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("reason = $%d", argIdx))
		args = append(args, *req.Reason)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}

	if len(setClauses) > 0 {
		argIdx++
		query := fmt.Sprintf(`UPDATE employee_schedules SET %s WHERE id::text = $%d`, strings.Join(setClauses, ", "), argIdx)
		args = append(args, id)
		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	// Replace locations if provided
	if req.Locations != nil {
		_, err := tx.Exec(ctx, `DELETE FROM employee_schedule_locations WHERE employee_schedule_id::text = $1`, id)
		if err != nil {
			return nil, err
		}
		for _, loc := range req.Locations {
			isPrimary := true
			if loc.IsPrimary != nil {
				isPrimary = *loc.IsPrimary
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO employee_schedule_locations (employee_schedule_id, attendance_location_id, day_of_week, is_primary)
				VALUES ($1::uuid, $2::uuid, $3, $4)
			`, id, loc.AttendanceLocationID, loc.DayOfWeek, isPrimary)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return GetEmployeeScheduleByID(ctx, id)
}

func DeleteEmployeeSchedule(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM employee_schedules WHERE id::text = $1`, id)
		return err
	})
}

// ============================================================
// RESOLVE SCHEDULE — Core function untuk absensi
// ============================================================

// ResolveEmployeeScheduleForDate returns the active schedule + location for an employee on a given date.
// Resolution order:
// 1. employee_schedules with specific_date match (highest priority)
// 2. employee_schedules with day_of_week + date range match
// 3. employees.work_schedule_id (old system fallback)
// 4. departments.work_schedule_id (old system fallback)
func ResolveEmployeeScheduleForDate(ctx context.Context, employeeID, date string) (*models.ResolvedSchedule, error) {
	// Step 1+2: Try new flexible schedule system
	query := `
		SELECT es.id,
			COALESCE(es.start_time::text, std.start_time::text, '08:00') as start_time,
			COALESCE(es.end_time::text, std.end_time::text, '17:00') as end_time,
			COALESCE(es.break_start::text, std.break_start::text, '12:00') as break_start,
			COALESCE(es.break_end::text, std.break_end::text, '13:00') as break_end,
			COALESCE(es.is_remote, FALSE) as is_remote,
			al.id as location_id, COALESCE(al.name, '') as location_name,
			COALESCE(al.address, '') as location_address,
			COALESCE(al.latitude, 0) as latitude, COALESCE(al.longitude, 0) as longitude,
			COALESCE(al.radius_meters, 0) as radius_meters,
			CASE WHEN es.specific_date IS NOT NULL THEN 'date_override' ELSE 'weekly_schedule' END as source
		FROM employee_schedules es
		LEFT JOIN schedule_templates st ON st.id = es.template_id AND st.is_active = TRUE AND st.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT * FROM schedule_template_days std 
			WHERE std.template_id = st.id 
			AND (std.day_of_week IS NULL OR std.day_of_week = (EXTRACT(DOW FROM $2::date)::int + 6) % 7)
			LIMIT 1
		) std ON TRUE
		LEFT JOIN LATERAL (
			SELECT * FROM employee_schedule_locations esl 
			WHERE esl.employee_schedule_id = es.id AND esl.is_primary = TRUE
			LIMIT 1
		) esl ON TRUE
		LEFT JOIN attendance_locations al ON al.id = esl.attendance_location_id AND al.is_active = TRUE AND al.deleted_at IS NULL
		WHERE es.employee_id::text = $1
			AND es.is_active = TRUE
			AND (
				(es.specific_date = $2::date)
				OR (
					es.specific_date IS NULL
					AND es.day_of_week = (EXTRACT(DOW FROM $2::date)::int + 6) % 7
					AND es.effective_from <= $2::date
					AND (es.effective_until IS NULL OR es.effective_until >= $2::date)
				)
			)
		ORDER BY 
			CASE WHEN es.specific_date IS NOT NULL THEN 0 ELSE 1 END,
			es.priority DESC
		LIMIT 1
	`

	var s models.ResolvedSchedule
	var locID *string
	var locName, locAddress string
	var lat, lng, radius float64

	err := database.Pool.QueryRow(ctx, query, employeeID, date).Scan(
		&s.ScheduleID, &s.StartTime, &s.EndTime, &s.BreakStart, &s.BreakEnd,
		&s.IsRemote, &locID, &locName, &locAddress, &lat, &lng, &radius, &s.Source,
	)
	if err == nil {
		if locID != nil {
			s.Location = &models.ResolvedLocation{
				Name:         locName,
				Address:      locAddress,
				Latitude:     lat,
				Longitude:    lng,
				RadiusMeters: int(radius),
			}
		}
		return &s, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	// Step 3: Fallback to employees.work_schedule_id (old system)
	oldQuery := `
		SELECT 
			'work_schedule' as source,
			COALESCE(ws.monday_start::text, '08:00'),
			COALESCE(ws.monday_end::text, '17:00'),
			COALESCE(ws.break_start::text, '12:00'),
			COALESCE(ws.break_end::text, '13:00'),
			COALESCE(e.is_remote, FALSE),
			al.id::text, COALESCE(al.name, ''), COALESCE(al.address, ''),
			COALESCE(al.latitude, 0), COALESCE(al.longitude, 0), COALESCE(al.radius_meters, 0)
		FROM employees e
		LEFT JOIN work_schedules ws ON ws.id = e.work_schedule_id
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN attendance_locations al ON al.is_active = TRUE AND al.deleted_at IS NULL
		WHERE e.id::text = $1 AND e.deleted_at IS NULL
		LIMIT 1
	`
	err = database.Pool.QueryRow(ctx, oldQuery, employeeID).Scan(
		&s.Source, &s.StartTime, &s.EndTime, &s.BreakStart, &s.BreakEnd,
		&s.IsRemote, &locID, &locName, &locAddress, &lat, &lng, &radius,
	)
	if err == nil {
		if locID != nil {
			s.Location = &models.ResolvedLocation{
				Name:         locName,
				Address:      locAddress,
				Latitude:     lat,
				Longitude:    lng,
				RadiusMeters: int(radius),
			}
		}
		return &s, nil
	}

	if err != pgx.ErrNoRows {
		// Step 4: Fallback to departments
		deptQuery := `
			SELECT 'department_schedule' as source,
				COALESCE(ws.monday_start::text, '08:00'), COALESCE(ws.monday_end::text, '17:00'),
				COALESCE(ws.break_start::text, '12:00'), COALESCE(ws.break_end::text, '13:00'),
				COALESCE(e.is_remote, FALSE),
				al.id::text, COALESCE(al.name, ''), COALESCE(al.address, ''),
				COALESCE(al.latitude, 0), COALESCE(al.longitude, 0), COALESCE(al.radius_meters, 0)
			FROM employees e
			JOIN departments d ON d.id = e.department_id
			LEFT JOIN work_schedules ws ON ws.id = d.work_schedule_id
			LEFT JOIN attendance_locations al ON al.is_active = TRUE AND al.deleted_at IS NULL
			WHERE e.id::text = $1 AND e.deleted_at IS NULL
			LIMIT 1
		`
		err = database.Pool.QueryRow(ctx, deptQuery, employeeID).Scan(
			&s.Source, &s.StartTime, &s.EndTime, &s.BreakStart, &s.BreakEnd,
			&s.IsRemote, &locID, &locName, &locAddress, &lat, &lng, &radius,
		)
		if err == nil && locID != nil {
			s.Location = &models.ResolvedLocation{
				Name:         locName,
				Address:      locAddress,
				Latitude:     lat,
				Longitude:    lng,
				RadiusMeters: int(radius),
			}
			return &s, nil
		}
	}

	return nil, fmt.Errorf("tidak ada jadwal yang ditemukan untuk karyawan ini")
}

// ============================================================
// Helpers
// ============================================================

func isPGUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "duplicate key")
}
