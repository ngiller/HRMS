package repository

import (
	"context"
	"encoding/json"

	"hrms-backend/internal/database"

	"github.com/jackc/pgx/v5"
)

// GetPermissionsByRoleSlug returns the permissions JSON for a given role slug.
// Returns nil if role not found.
func GetPermissionsByRoleSlug(ctx context.Context, roleSlug string) (map[string]map[string]bool, error) {
	query := `SELECT permissions FROM roles WHERE slug = $1 AND is_active = TRUE`

	var permissionsJSON []byte
	err := database.Pool.QueryRow(ctx, query, roleSlug).Scan(&permissionsJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var permissions map[string]map[string]bool
	if err := json.Unmarshal(permissionsJSON, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// CheckPermission checks if a role has a specific action on a module.
func CheckPermission(ctx context.Context, roleSlug string, module string, action string) (bool, error) {
	permissions, err := GetPermissionsByRoleSlug(ctx, roleSlug)
	if err != nil {
		return false, err
	}
	if permissions == nil {
		return false, nil
	}

	modulePerms, ok := permissions[module]
	if !ok {
		return false, nil
	}

	hasAccess, ok := modulePerms[action]
	if !ok {
		return false, nil
	}

	return hasAccess, nil
}
