package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListWorkSchedules(ctx context.Context, page, perPage int, search string) ([]models.WorkScheduleSummary, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM work_schedules ws
		WHERE ws.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND LOWER(ws.name) LIKE LOWER($%d)", argIdx)
		args = append(args, "%"+search+"%")
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT ws.id, ws.name, ws.schedule_type::text,
			COALESCE(ws.description, '') as description,
			ws.weekly_hours, ws.is_active, ws.created_at
		FROM work_schedules ws
		WHERE ws.deleted_at IS NULL
	`)
	if search != "" {
		query += fmt.Sprintf(" AND LOWER(ws.name) LIKE LOWER($%d)", argIdx-1)
	}
	query += fmt.Sprintf(" ORDER BY ws.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schedules []models.WorkScheduleSummary
	for rows.Next() {
		var ws models.WorkScheduleSummary
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.ScheduleType, &ws.Description, &ws.WeeklyHours, &ws.IsActive, &ws.CreatedAt); err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, ws)
	}
	return schedules, total, nil
}

func GetWorkScheduleByID(ctx context.Context, id string) (*models.WorkSchedule, error) {
	query := `
		SELECT ws.id, ws.name, ws.schedule_type::text,
			COALESCE(ws.description, ''),
			ws.monday_start::text, ws.monday_end::text,
			ws.tuesday_start::text, ws.tuesday_end::text,
			ws.wednesday_start::text, ws.wednesday_end::text,
			ws.thursday_start::text, ws.thursday_end::text,
			ws.friday_start::text, ws.friday_end::text,
			ws.saturday_start::text, ws.saturday_end::text,
			ws.sunday_start::text, ws.sunday_end::text,
			ws.break_start::text, ws.break_end::text,
			ws.late_tolerance_minutes, ws.early_leave_tolerance,
			ws.weekly_hours, ws.is_active,
			ws.created_at, ws.updated_at, ws.deleted_at
		FROM work_schedules ws
		WHERE (ws.id::text = $1) AND ws.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var ws models.WorkSchedule
	err := row.Scan(
		&ws.ID, &ws.Name, &ws.ScheduleType, &ws.Description,
		&ws.MondayStart, &ws.MondayEnd,
		&ws.TuesdayStart, &ws.TuesdayEnd,
		&ws.WednesdayStart, &ws.WednesdayEnd,
		&ws.ThursdayStart, &ws.ThursdayEnd,
		&ws.FridayStart, &ws.FridayEnd,
		&ws.SaturdayStart, &ws.SaturdayEnd,
		&ws.SundayStart, &ws.SundayEnd,
		&ws.BreakStart, &ws.BreakEnd,
		&ws.LateToleranceMinutes, &ws.EarlyLeaveTolerance,
		&ws.WeeklyHours, &ws.IsActive,
		&ws.CreatedAt, &ws.UpdatedAt, &ws.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func CreateWorkSchedule(ctx context.Context, req *models.CreateWorkScheduleRequest, userID string) (*models.WorkSchedule, error) {
	var ws *models.WorkSchedule
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO work_schedules (name, schedule_type, description,
				monday_start, monday_end,
				tuesday_start, tuesday_end,
				wednesday_start, wednesday_end,
				thursday_start, thursday_end,
				friday_start, friday_end,
				saturday_start, saturday_end,
				sunday_start, sunday_end,
				break_start, break_end,
				late_tolerance_minutes, early_leave_tolerance, weekly_hours)
			VALUES ($1, $2::work_schedule_type, $3,
				$4::time, $5::time,
				$6::time, $7::time,
				$8::time, $9::time,
				$10::time, $11::time,
				$12::time, $13::time,
				NULLIF(NULLIF($14, ''), 'null')::time, NULLIF(NULLIF($15, ''), 'null')::time,
				NULLIF(NULLIF($16, ''), 'null')::time, NULLIF(NULLIF($17, ''), 'null')::time,
				$18::time, $19::time,
				$20, $21, $22)
			RETURNING id, name, schedule_type::text,
				COALESCE(description, ''),
				monday_start::text, monday_end::text,
				tuesday_start::text, tuesday_end::text,
				wednesday_start::text, wednesday_end::text,
				thursday_start::text, thursday_end::text,
				friday_start::text, friday_end::text,
				saturday_start::text, saturday_end::text,
				sunday_start::text, sunday_end::text,
				break_start::text, break_end::text,
				late_tolerance_minutes, early_leave_tolerance,
				weekly_hours, is_active,
				created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query,
			req.Name, req.ScheduleType, req.Description,
			req.MondayStart, req.MondayEnd,
			req.TuesdayStart, req.TuesdayEnd,
			req.WednesdayStart, req.WednesdayEnd,
			req.ThursdayStart, req.ThursdayEnd,
			req.FridayStart, req.FridayEnd,
			coalesceStr(req.SaturdayStart), coalesceStr(req.SaturdayEnd),
			coalesceStr(req.SundayStart), coalesceStr(req.SundayEnd),
			req.BreakStart, req.BreakEnd,
			req.LateToleranceMinutes, req.EarlyLeaveTolerance, req.WeeklyHours,
		)
		var result models.WorkSchedule
		if err := row.Scan(
			&result.ID, &result.Name, &result.ScheduleType, &result.Description,
			&result.MondayStart, &result.MondayEnd,
			&result.TuesdayStart, &result.TuesdayEnd,
			&result.WednesdayStart, &result.WednesdayEnd,
			&result.ThursdayStart, &result.ThursdayEnd,
			&result.FridayStart, &result.FridayEnd,
			&result.SaturdayStart, &result.SaturdayEnd,
			&result.SundayStart, &result.SundayEnd,
			&result.BreakStart, &result.BreakEnd,
			&result.LateToleranceMinutes, &result.EarlyLeaveTolerance,
			&result.WeeklyHours, &result.IsActive,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		ws = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func UpdateWorkSchedule(ctx context.Context, id string, req *models.UpdateWorkScheduleRequest, userID string) (*models.WorkSchedule, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.ScheduleType != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("schedule_type = $%d::work_schedule_type", argIdx))
		args = append(args, *req.ScheduleType)
	}
	if req.Description != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
	}
	if req.MondayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("monday_start = $%d::time", argIdx))
		args = append(args, *req.MondayStart)
	}
	if req.MondayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("monday_end = $%d::time", argIdx))
		args = append(args, *req.MondayEnd)
	}
	if req.TuesdayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("tuesday_start = $%d::time", argIdx))
		args = append(args, *req.TuesdayStart)
	}
	if req.TuesdayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("tuesday_end = $%d::time", argIdx))
		args = append(args, *req.TuesdayEnd)
	}
	if req.WednesdayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("wednesday_start = $%d::time", argIdx))
		args = append(args, *req.WednesdayStart)
	}
	if req.WednesdayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("wednesday_end = $%d::time", argIdx))
		args = append(args, *req.WednesdayEnd)
	}
	if req.ThursdayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("thursday_start = $%d::time", argIdx))
		args = append(args, *req.ThursdayStart)
	}
	if req.ThursdayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("thursday_end = $%d::time", argIdx))
		args = append(args, *req.ThursdayEnd)
	}
	if req.FridayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("friday_start = $%d::time", argIdx))
		args = append(args, *req.FridayStart)
	}
	if req.FridayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("friday_end = $%d::time", argIdx))
		args = append(args, *req.FridayEnd)
	}
	if req.SaturdayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("saturday_start = NULLIF(NULLIF($%d, ''), 'null')::time", argIdx))
		args = append(args, *req.SaturdayStart)
	}
	if req.SaturdayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("saturday_end = NULLIF(NULLIF($%d, ''), 'null')::time", argIdx))
		args = append(args, *req.SaturdayEnd)
	}
	if req.SundayStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("sunday_start = NULLIF(NULLIF($%d, ''), 'null')::time", argIdx))
		args = append(args, *req.SundayStart)
	}
	if req.SundayEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("sunday_end = NULLIF(NULLIF($%d, ''), 'null')::time", argIdx))
		args = append(args, *req.SundayEnd)
	}
	if req.BreakStart != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("break_start = $%d::time", argIdx))
		args = append(args, *req.BreakStart)
	}
	if req.BreakEnd != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("break_end = $%d::time", argIdx))
		args = append(args, *req.BreakEnd)
	}
	if req.LateToleranceMinutes != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("late_tolerance_minutes = $%d", argIdx))
		args = append(args, *req.LateToleranceMinutes)
	}
	if req.EarlyLeaveTolerance != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("early_leave_tolerance = $%d", argIdx))
		args = append(args, *req.EarlyLeaveTolerance)
	}
	if req.WeeklyHours != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("weekly_hours = $%d", argIdx))
		args = append(args, *req.WeeklyHours)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}

	if len(setClauses) == 0 {
		return GetWorkScheduleByID(ctx, id)
	}

	var ws *models.WorkSchedule
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE work_schedules SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, name, schedule_type::text,
				COALESCE(description, ''),
				monday_start::text, monday_end::text,
				tuesday_start::text, tuesday_end::text,
				wednesday_start::text, wednesday_end::text,
				thursday_start::text, thursday_end::text,
				friday_start::text, friday_end::text,
				saturday_start::text, saturday_end::text,
				sunday_start::text, sunday_end::text,
				break_start::text, break_end::text,
				late_tolerance_minutes, early_leave_tolerance,
				weekly_hours, is_active,
				created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.WorkSchedule
		if err := row.Scan(
			&result.ID, &result.Name, &result.ScheduleType, &result.Description,
			&result.MondayStart, &result.MondayEnd,
			&result.TuesdayStart, &result.TuesdayEnd,
			&result.WednesdayStart, &result.WednesdayEnd,
			&result.ThursdayStart, &result.ThursdayEnd,
			&result.FridayStart, &result.FridayEnd,
			&result.SaturdayStart, &result.SaturdayEnd,
			&result.SundayStart, &result.SundayEnd,
			&result.BreakStart, &result.BreakEnd,
			&result.LateToleranceMinutes, &result.EarlyLeaveTolerance,
			&result.WeeklyHours, &result.IsActive,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
		); err != nil {
			return err
		}
		ws = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func DeleteWorkSchedule(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE work_schedules SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckWorkScheduleNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM work_schedules WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func CheckWorkScheduleUsedByDepartments(ctx context.Context, id string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM departments WHERE work_schedule_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}
