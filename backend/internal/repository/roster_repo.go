package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ============================================================
// DEPARTMENT ROSTERS
// ============================================================

func ListDepartmentRosters(ctx context.Context, page, perPage int, departmentID string) ([]models.DepartmentRosterSummary, int, error) {
	where := "WHERE dr.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 0

	if departmentID != "" {
		argIdx++
		where += fmt.Sprintf(" AND dr.department_id::text = $%d", argIdx)
		args = append(args, departmentID)
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM department_rosters dr %s`, where)
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	argIdx++
	query := fmt.Sprintf(`
		SELECT dr.id, dr.department_id, COALESCE(d.name, '') as department_name,
			dr.name, dr.month, dr.year, dr.is_published,
			COALESCE(dr.notes, ''),
			(SELECT COUNT(*) FROM roster_entries re WHERE re.roster_id = dr.id) as entry_count,
			dr.created_at
		FROM department_rosters dr
		LEFT JOIN departments d ON d.id = dr.department_id
		%s ORDER BY dr.year DESC, dr.month DESC, dr.name ASC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rosters []models.DepartmentRosterSummary
	for rows.Next() {
		var r models.DepartmentRosterSummary
		if err := rows.Scan(&r.ID, &r.DepartmentID, &r.DepartmentName, &r.Name, &r.Month, &r.Year, &r.IsPublished, &r.Notes, &r.EntryCount, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		rosters = append(rosters, r)
	}
	return rosters, total, nil
}

func GetDepartmentRosterByID(ctx context.Context, id string) (*models.DepartmentRoster, error) {
	query := `
		SELECT dr.id, dr.department_id, COALESCE(d.name, ''),
			dr.name, dr.month, dr.year, dr.is_published,
			COALESCE(dr.notes, ''), dr.created_by, dr.created_at, dr.updated_at, dr.deleted_at
		FROM department_rosters dr
		LEFT JOIN departments d ON d.id = dr.department_id
		WHERE dr.id::text = $1 AND dr.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var r models.DepartmentRoster
	err := row.Scan(&r.ID, &r.DepartmentID, &r.DepartmentName, &r.Name, &r.Month, &r.Year,
		&r.IsPublished, &r.Notes, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateDepartmentRoster(ctx context.Context, req *models.CreateDepartmentRosterRequest, userID string) (*models.DepartmentRoster, error) {
	var r *models.DepartmentRoster
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO department_rosters (department_id, name, month, year, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, department_id, name, month, year, is_published,
				COALESCE(notes, ''), created_by, created_at, updated_at, deleted_at
		`
		// Get department name
		var deptName string
		_ = tx.QueryRow(ctx, `SELECT COALESCE(name, '') FROM departments WHERE id::text = $1`, req.DepartmentID).Scan(&deptName)

		row := tx.QueryRow(ctx, query, req.DepartmentID, req.Name, req.Month, req.Year, req.Notes, userID)
		var result models.DepartmentRoster
		if err := row.Scan(&result.ID, &result.DepartmentID, &result.Name, &result.Month, &result.Year,
			&result.IsPublished, &result.Notes, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		result.DepartmentName = deptName
		r = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func UpdateDepartmentRoster(ctx context.Context, id string, req *models.UpdateDepartmentRosterRequest, userID string) (*models.DepartmentRoster, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx)); args = append(args, *req.Name)
	}
	if req.Month != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("month = $%d", argIdx)); args = append(args, *req.Month)
	}
	if req.Year != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("year = $%d", argIdx)); args = append(args, *req.Year)
	}
	if req.IsPublished != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("is_published = $%d", argIdx)); args = append(args, *req.IsPublished)
	}
	if req.Notes != nil {
		argIdx++; setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx)); args = append(args, *req.Notes)
	}

	if len(setClauses) == 0 {
		return GetDepartmentRosterByID(ctx, id)
	}

	var r *models.DepartmentRoster
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE department_rosters SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, department_id, name, month, year, is_published,
				COALESCE(notes, ''), created_by, created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.DepartmentRoster
		if err := row.Scan(&result.ID, &result.DepartmentID, &result.Name, &result.Month, &result.Year,
			&result.IsPublished, &result.Notes, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		r = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func DeleteDepartmentRoster(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE department_rosters SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckRosterNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM department_rosters WHERE name = $1 AND deleted_at IS NULL`
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
// ROSTER ENTRIES
// ============================================================

func ListRosterEntries(ctx context.Context, rosterID string) ([]models.RosterEntry, error) {
	query := `
		SELECT re.id, re.roster_id, re.employee_id,
			COALESCE(e.full_name, '') as employee_name,
			re.date::text, re.shift_id,
			COALESCE(s.name, '') as shift_name,
			COALESCE(s.code, '') as shift_code,
			COALESCE(s.color, '#3B82F6') as shift_color,
			s.start_time::text, s.end_time::text,
			COALESCE(re.notes, ''),
			re.created_at, re.updated_at
		FROM roster_entries re
		JOIN employees e ON e.id = re.employee_id
		LEFT JOIN shifts s ON s.id = re.shift_id
		WHERE re.roster_id::text = $1
		ORDER BY re.date ASC, e.full_name ASC
	`
	rows, err := database.Pool.Query(ctx, query, rosterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.RosterEntry
	for rows.Next() {
		var e models.RosterEntry
		if err := rows.Scan(&e.ID, &e.RosterID, &e.EmployeeID, &e.EmployeeName,
			&e.Date, &e.ShiftID, &e.ShiftName, &e.ShiftCode, &e.ShiftColor,
			&e.StartTime, &e.EndTime, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func GetRosterCalendar(ctx context.Context, rosterID string) ([]models.RosterCalendarEntry, error) {
	// Get all unique employees with their entries for this roster
	query := `
		SELECT e.id::text, COALESCE(e.full_name, ''),
			re.date::text,
			COALESCE(re.shift_id::text, ''),
			COALESCE(s.name, ''),
			COALESCE(s.code, ''),
			COALESCE(s.color, '#3B82F6'),
			s.start_time::text, s.end_time::text,
			re.id::text,
			COALESCE(re.notes, '')
		FROM roster_entries re
		JOIN employees e ON e.id = re.employee_id
		LEFT JOIN shifts s ON s.id = re.shift_id
		WHERE re.roster_id::text = $1
		ORDER BY e.full_name ASC, re.date ASC
	`
	rows, err := database.Pool.Query(ctx, query, rosterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	empMap := make(map[string]*models.RosterCalendarEntry)
	var order []string

	for rows.Next() {
		var empID, empName, date, shiftID, shiftName, shiftCode, shiftColor, startTime, endTime, entryID, notes string
		if err := rows.Scan(&empID, &empName, &date, &shiftID, &shiftName, &shiftCode, &shiftColor, &startTime, &endTime, &entryID, &notes); err != nil {
			return nil, err
		}

		if _, ok := empMap[empID]; !ok {
			empMap[empID] = &models.RosterCalendarEntry{
				EmployeeID:   empID,
				EmployeeName: empName,
				Days:         make(map[string]models.RosterDayInfo),
			}
			order = append(order, empID)
		}

		empMap[empID].Days[date] = models.RosterDayInfo{
			ShiftID:    shiftID,
			ShiftName:  shiftName,
			ShiftCode:  shiftCode,
			ShiftColor: shiftColor,
			StartTime:  startTime,
			EndTime:    endTime,
			EntryID:    entryID,
			Notes:      notes,
		}
	}

	result := make([]models.RosterCalendarEntry, 0, len(order))
	for _, empID := range order {
		result = append(result, *empMap[empID])
	}
	return result, nil
}

func CreateRosterEntry(ctx context.Context, req *models.CreateRosterEntryRequest, userID string) (*models.RosterEntry, error) {
	var e *models.RosterEntry
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO roster_entries (roster_id, employee_id, date, shift_id, notes)
			VALUES ($1, $2, $3::date, $4, $5)
			ON CONFLICT (roster_id, employee_id, date)
			DO UPDATE SET shift_id = EXCLUDED.shift_id, notes = EXCLUDED.notes
			RETURNING id, roster_id, employee_id, date::text, shift_id, COALESCE(notes, ''),
				created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, req.RosterID, req.EmployeeID, req.Date, req.ShiftID, req.Notes)
		var result models.RosterEntry
		if err := row.Scan(&result.ID, &result.RosterID, &result.EmployeeID, &result.Date,
			&result.ShiftID, &result.Notes, &result.CreatedAt, &result.UpdatedAt); err != nil {
			return err
		}

		// Get employee name
		_ = tx.QueryRow(ctx, `SELECT COALESCE(full_name, '') FROM employees WHERE id::text = $1`, req.EmployeeID).Scan(&result.EmployeeName)
		// Get shift info
		_ = tx.QueryRow(ctx, `SELECT COALESCE(name, ''), COALESCE(code, ''), COALESCE(color, '#3B82F6'), start_time::text, end_time::text FROM shifts WHERE id::text = $1`, req.ShiftID).Scan(
			&result.ShiftName, &result.ShiftCode, &result.ShiftColor, &result.StartTime, &result.EndTime)
		e = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func BulkCreateRosterEntries(ctx context.Context, entries []models.CreateRosterEntryRequest, clearExisting bool, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		if clearExisting && len(entries) > 0 {
			// Get unique employee IDs from entries to clear their existing roster entries
			empMap := make(map[string]bool)
			for _, entry := range entries {
				if entry.EmployeeID != "" {
					empMap[entry.EmployeeID] = true
				}
			}
			for empID := range empMap {
				_, err := tx.Exec(ctx, `DELETE FROM roster_entries WHERE roster_id::text = $1 AND employee_id::text = $2`, entries[0].RosterID, empID)
				if err != nil {
					return err
				}
			}
		}

		for _, entry := range entries {
			_, err := tx.Exec(ctx, `
				INSERT INTO roster_entries (roster_id, employee_id, date, shift_id, notes)
				VALUES ($1, $2, $3::date, $4, $5)
				ON CONFLICT (roster_id, employee_id, date)
				DO UPDATE SET shift_id = EXCLUDED.shift_id, notes = EXCLUDED.notes
			`, entry.RosterID, entry.EmployeeID, entry.Date, entry.ShiftID, entry.Notes)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func DeleteRosterEntry(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM roster_entries WHERE id::text = $1`, id)
		return err
	})
}

// CheckEmployeeRosterForToday returns true if the employee has a roster entry for today
// within an active roster for their department.
func CheckEmployeeRosterForToday(ctx context.Context, employeeID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM roster_entries re
			JOIN department_rosters dr ON dr.id = re.roster_id
			WHERE re.employee_id::text = $1
				AND re.date = CURRENT_DATE
				AND dr.year = EXTRACT(YEAR FROM CURRENT_DATE)::int
				AND dr.month = EXTRACT(MONTH FROM CURRENT_DATE)::int
				AND dr.deleted_at IS NULL
		)
	`
	var hasEntry bool
	err := database.Pool.QueryRow(ctx, query, employeeID).Scan(&hasEntry)
	if err != nil {
		return false, err
	}
	return hasEntry, nil
}

// DepartmentHasActiveRoster returns true if the employee's department has an active
// roster for the current month.
func DepartmentHasActiveRoster(ctx context.Context, employeeID string) (bool, error) {
	var exists bool
	err := database.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM department_rosters dr
			JOIN employees e ON e.department_id = dr.department_id
			WHERE e.id::text = $1
				AND dr.year = EXTRACT(YEAR FROM CURRENT_DATE)::int
				AND dr.month = EXTRACT(MONTH FROM CURRENT_DATE)::int
				AND dr.deleted_at IS NULL
		)
	`, employeeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func DeleteEmployeeRosterEntries(ctx context.Context, rosterID, employeeID, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM roster_entries WHERE roster_id::text = $1 AND employee_id::text = $2`, rosterID, employeeID)
		return err
	})
}

