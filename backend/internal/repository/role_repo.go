package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// ListRoles returns paginated role list.
func ListRoles(ctx context.Context, page, perPage int, search string) ([]models.RoleSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM roles`
	args := []interface{}{}
	argIdx := 0

	if search != "" {
		argIdx++
		countQuery += fmt.Sprintf(" WHERE (LOWER(name) LIKE LOWER($%d) OR LOWER(slug) LIKE LOWER($%d))", argIdx, argIdx)
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
		SELECT r.id, r.name, r.slug, COALESCE(r.description, '') as description,
			r.is_system_role, r.is_active,
			(SELECT COUNT(*) FROM employees e WHERE e.role_id = r.id AND e.deleted_at IS NULL) as employee_count,
			r.created_at
		FROM roles r
	`)
	if search != "" {
		query += fmt.Sprintf(" WHERE (LOWER(r.name) LIKE LOWER($%d) OR LOWER(r.slug) LIKE LOWER($%d))", argIdx-1, argIdx-1)
	}
	query += fmt.Sprintf(" ORDER BY r.is_system_role DESC, r.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []models.RoleSummary
	for rows.Next() {
		var role models.RoleSummary
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description,
			&role.IsSystemRole, &role.IsActive, &role.EmployeeCnt, &role.CreatedAt); err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}

	return roles, total, nil
}

// GetRoleByID returns a single role by ID.
func GetRoleByID(ctx context.Context, id string) (*models.Role, error) {
	query := `
		SELECT r.id, r.name, r.slug, COALESCE(r.description, '') as description,
			r.permissions, r.is_system_role, r.is_active, r.created_at, r.updated_at
		FROM roles r
		WHERE r.id::text = $1
	`
	row := database.Pool.QueryRow(ctx, query, id)

	var role models.Role
	var permissionsJSON []byte
	err := row.Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description,
		&permissionsJSON, &role.IsSystemRole, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(permissionsJSON) > 0 {
		if err := json.Unmarshal(permissionsJSON, &role.Permissions); err != nil {
			return nil, err
		}
	} else {
		role.Permissions = make(map[string]map[string]bool)
	}

	return &role, nil
}

// CreateRole inserts a new role.
func CreateRole(ctx context.Context, req *models.CreateRoleRequest, userID string) (*models.Role, error) {
	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, err
	}
	if req.Permissions == nil {
		permissionsJSON = []byte("{}")
	}

	var role *models.Role
	err = database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO roles (name, slug, description, permissions)
			VALUES ($1, $2, NULLIF($3, ''), $4::jsonb)
			RETURNING id, name, slug, COALESCE(description, '') as description,
				permissions, is_system_role, is_active, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, req.Name, req.Slug, req.Description, string(permissionsJSON))

		var result models.Role
		var retPermissionsJSON []byte
		if err := row.Scan(
			&result.ID, &result.Name, &result.Slug, &result.Description,
			&retPermissionsJSON, &result.IsSystemRole, &result.IsActive, &result.CreatedAt, &result.UpdatedAt,
		); err != nil {
			return err
		}

		if len(retPermissionsJSON) > 0 {
			json.Unmarshal(retPermissionsJSON, &result.Permissions)
		} else {
			result.Permissions = make(map[string]map[string]bool)
		}
		role = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole updates an existing role.
func UpdateRole(ctx context.Context, id string, req *models.UpdateRoleRequest, userID string) (*models.Role, error) {
	// Build dynamic SET
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 0

	if req.Name != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
	}
	if req.Slug != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", argIdx))
		args = append(args, *req.Slug)
	}
	if req.Description != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("description = NULLIF($%d, '')", argIdx))
		args = append(args, *req.Description)
	}
	if req.IsActive != nil {
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
	}
	if req.Permissions != nil {
		argIdx++
		permJSON, err := json.Marshal(*req.Permissions)
		if err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("permissions = $%d::jsonb", argIdx))
		args = append(args, string(permJSON))
	}

	if len(setClauses) == 0 {
		return GetRoleByID(ctx, id)
	}

	var role *models.Role
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		argIdx++
		query := fmt.Sprintf(`
			UPDATE roles SET %s
			WHERE id::text = $%d
			RETURNING id, name, slug, COALESCE(description, '') as description,
				permissions, is_system_role, is_active, created_at, updated_at
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)

		var result models.Role
		var permissionsJSON []byte
		if err := row.Scan(
			&result.ID, &result.Name, &result.Slug, &result.Description,
			&permissionsJSON, &result.IsSystemRole, &result.IsActive, &result.CreatedAt, &result.UpdatedAt,
		); err != nil {
			return err
		}

		if len(permissionsJSON) > 0 {
			json.Unmarshal(permissionsJSON, &result.Permissions)
		} else {
			result.Permissions = make(map[string]map[string]bool)
		}
		role = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

// DeleteRole deletes a role (only non-system roles).
func DeleteRole(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM roles WHERE id::text = $1 AND is_system_role = FALSE`, id)
		return err
	})
}

// CheckRoleNameExists checks if a role name already exists (excluding a given ID).
func CheckRoleNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM roles WHERE name = $1`
	args := []interface{}{name}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// CheckRoleSlugExists checks if a role slug already exists (excluding a given ID).
func CheckRoleSlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM roles WHERE slug = $1`
	args := []interface{}{slug}
	if excludeID != "" {
		query += ` AND id::text != $2`
		args = append(args, excludeID)
	}
	var count int
	err := database.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// CheckRoleHasEmployees checks if a role is assigned to any employees.
func CheckRoleHasEmployees(ctx context.Context, id string) (bool, error) {
	var count int
	err := database.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM employees WHERE role_id::text = $1 AND deleted_at IS NULL`, id,
	).Scan(&count)
	return count > 0, err
}

// CheckRoleIsSystem checks if a role is a system role.
func CheckRoleIsSystem(ctx context.Context, id string) (bool, error) {
	var isSystem bool
	err := database.Pool.QueryRow(ctx,
		`SELECT is_system_role FROM roles WHERE id::text = $1`, id,
	).Scan(&isSystem)
	if err != nil {
		return false, err
	}
	return isSystem, nil
}
