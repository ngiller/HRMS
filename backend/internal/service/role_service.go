package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

func (s *RoleService) ListRoles(ctx context.Context, page, perPage int, search string) (*models.RoleListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	roles, total, err := repository.ListRoles(ctx, page, perPage, search)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data role: %w", err)
	}
	if roles == nil {
		roles = []models.RoleSummary{}
	}

	return &models.RoleListResponse{
		Roles:   roles,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *RoleService) GetRole(ctx context.Context, id string) (*models.Role, error) {
	role, err := repository.GetRoleByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data role")
	}
	if role == nil {
		return nil, errors.New("role tidak ditemukan")
	}
	return role, nil
}

func (s *RoleService) CreateRole(ctx context.Context, req *models.CreateRoleRequest, userID string) (*models.Role, error) {
	// Validasi
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))

	if req.Name == "" {
		return nil, errors.New("nama role harus diisi")
	}
	if req.Slug == "" {
		return nil, errors.New("slug role harus diisi")
	}

	// Cek duplikasi
	exists, err := repository.CheckRoleNameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi role: %w", err)
	}
	if exists {
		return nil, errors.New("nama role sudah digunakan")
	}

	slugExists, err := repository.CheckRoleSlugExists(ctx, req.Slug, "")
	if err != nil {
		return nil, fmt.Errorf("gagal validasi role: %w", err)
	}
	if slugExists {
		return nil, errors.New("slug role sudah digunakan")
	}

	// Normalize permissions
	if req.Permissions == nil {
		req.Permissions = make(map[string]map[string]bool)
	}

	role, err := repository.CreateRole(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat role: %w", err)
	}
	return role, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id string, req *models.UpdateRoleRequest, userID string) (*models.Role, error) {
	existing, err := repository.GetRoleByID(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data role")
	}
	if existing == nil {
		return nil, errors.New("role tidak ditemukan")
	}

	// Cek duplikasi name
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.New("nama role tidak boleh kosong")
		}
		req.Name = &name
		if name != existing.Name {
			exists, err := repository.CheckRoleNameExists(ctx, name, id)
			if err != nil {
				return nil, fmt.Errorf("gagal validasi role: %w", err)
			}
			if exists {
				return nil, errors.New("nama role sudah digunakan")
			}
		}
	}

	// Cek duplikasi slug
	if req.Slug != nil {
		slug := strings.TrimSpace(strings.ToLower(*req.Slug))
		if slug == "" {
			return nil, errors.New("slug role tidak boleh kosong")
		}
		req.Slug = &slug
		if slug != existing.Slug {
			slugExists, err := repository.CheckRoleSlugExists(ctx, slug, id)
			if err != nil {
				return nil, fmt.Errorf("gagal validasi role: %w", err)
			}
			if slugExists {
				return nil, errors.New("slug role sudah digunakan")
			}
		}
	}

	role, err := repository.UpdateRole(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui role: %w", err)
	}
	return role, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, id, userID string) error {
	existing, err := repository.GetRoleByID(ctx, id)
	if err != nil {
		return errors.New("gagal memuat data role")
	}
	if existing == nil {
		return errors.New("role tidak ditemukan")
	}

	// Cek system role
	if existing.IsSystemRole {
		return errors.New("role sistem tidak dapat dihapus")
	}

	// Cek apakah masih dipakai
	hasEmployees, err := repository.CheckRoleHasEmployees(ctx, id)
	if err != nil {
		return fmt.Errorf("gagal validasi role: %w", err)
	}
	if hasEmployees {
		return errors.New("role masih digunakan oleh karyawan, tidak dapat dihapus")
	}

	err = repository.DeleteRole(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus role: %w", err)
	}
	return nil
}

func (s *RoleService) GetPermissionTemplate(ctx context.Context) []models.PermissionModule {
	return models.DefaultPermissionTemplate()
}
