package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListAttendanceLocations(ctx context.Context, page, perPage int, search string) ([]models.AttendanceLocationSummary, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM attendance_locations al
		WHERE al.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND LOWER(al.name) LIKE LOWER($%d)", argIdx)
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
		SELECT al.id, al.name, COALESCE(al.address, ''), al.latitude, al.longitude,
			al.radius_meters, al.is_active, al.created_at
		FROM attendance_locations al
		WHERE al.deleted_at IS NULL
	`)
	if search != "" {
		query += fmt.Sprintf(" AND LOWER(al.name) LIKE LOWER($%d)", argIdx-1)
	}
	query += fmt.Sprintf(" ORDER BY al.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var locations []models.AttendanceLocationSummary
	for rows.Next() {
		var loc models.AttendanceLocationSummary
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Latitude, &loc.Longitude, &loc.RadiusMeters, &loc.IsActive, &loc.CreatedAt); err != nil {
			return nil, 0, err
		}
		locations = append(locations, loc)
	}
	return locations, total, nil
}

func GetAllAttendanceLocations(ctx context.Context) ([]models.AttendanceLocationSummary, error) {
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
	return locations, nil
}

func GetAttendanceLocationByID(ctx context.Context, id string) (*models.AttendanceLocation, error) {
	query := `
		SELECT al.id, al.name, COALESCE(al.address, ''), al.latitude, al.longitude,
			al.radius_meters, al.is_active, al.created_at, al.updated_at, al.deleted_at
		FROM attendance_locations al
		WHERE (al.id::text = $1) AND al.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var loc models.AttendanceLocation
	err := row.Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Latitude, &loc.Longitude,
		&loc.RadiusMeters, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt, &loc.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func CreateAttendanceLocation(ctx context.Context, req *models.CreateAttendanceLocationRequest, userID string) (*models.AttendanceLocation, error) {
	var loc *models.AttendanceLocation
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO attendance_locations (name, address, latitude, longitude, radius_meters)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, COALESCE(address, ''), latitude, longitude,
				radius_meters, is_active, created_at, updated_at, deleted_at
		`
		row := tx.QueryRow(ctx, query, req.Name, req.Address, req.Latitude, req.Longitude, req.RadiusMeters)
		var result models.AttendanceLocation
		if err := row.Scan(&result.ID, &result.Name, &result.Address, &result.Latitude, &result.Longitude,
			&result.RadiusMeters, &result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		loc = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func UpdateAttendanceLocation(ctx context.Context, id string, req *models.UpdateAttendanceLocationRequest, userID string) (*models.AttendanceLocation, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.Address != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argIdx))
		args = append(args, *req.Address)
	}
	if req.Latitude != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("latitude = $%d", argIdx))
		args = append(args, *req.Latitude)
	}
	if req.Longitude != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", argIdx))
		args = append(args, *req.Longitude)
	}
	if req.RadiusMeters != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("radius_meters = $%d", argIdx))
		args = append(args, *req.RadiusMeters)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}

	if len(setClauses) == 0 {
		return GetAttendanceLocationByID(ctx, id)
	}

	var loc *models.AttendanceLocation
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE attendance_locations SET %s
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, name, COALESCE(address, ''), latitude, longitude,
				radius_meters, is_active, created_at, updated_at, deleted_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.AttendanceLocation
		if err := row.Scan(&result.ID, &result.Name, &result.Address, &result.Latitude, &result.Longitude,
			&result.RadiusMeters, &result.IsActive, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt); err != nil {
			return err
		}
		loc = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func DeleteAttendanceLocation(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE attendance_locations SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

func CheckAttendanceLocationNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM attendance_locations WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}
