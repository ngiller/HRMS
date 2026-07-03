package service

import (
	"context"
	"fmt"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type OrganizationService struct{}

func NewOrganizationService() *OrganizationService {
	return &OrganizationService{}
}

func (s *OrganizationService) GetTree(ctx context.Context) ([]*models.OrgTreeNode, error) {
	depts, err := repository.GetAllDeptsForTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat departemen: %w", err)
	}

	positions, err := repository.GetAllPositionsForTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat posisi: %w", err)
	}

	employees, err := repository.GetAllEmployeesForTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat karyawan: %w", err)
	}

	// Build department map
	deptMap := make(map[string]*models.OrgTreeNode)
	for _, d := range depts {
		id := d.ID.String()
		node := &models.OrgTreeNode{
			ID:    id,
			Label: d.Name,
			Type:  "department",
			Meta: map[string]any{
				"head_name": d.HeadName,
			},
		}
		deptMap[id] = node
	}

	// Attach child departments and positions to parents
	var roots []*models.OrgTreeNode
	for _, d := range depts {
		node := deptMap[d.ID.String()]
		if d.ParentID != nil {
			parentID := d.ParentID.String()
			if parent, ok := deptMap[parentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	// Attach positions to departments
	for _, p := range positions {
		posNode := &models.OrgTreeNode{
			ID:    p.ID.String(),
			Label: p.Name,
			Type:  "position",
			Meta: map[string]any{
				"grade_name": p.GradeName,
			},
		}
		deptID := p.DepartmentID.String()
		if dept, ok := deptMap[deptID]; ok {
			dept.Children = append(dept.Children, posNode)
		} else {
			// Position without valid department — add as root (shouldn't happen normally)
			posNode.Meta["department_id"] = deptID
			roots = append(roots, posNode)
		}
	}

	// Attach employees to positions (or departments if no position)
	posMap := make(map[string]*models.OrgTreeNode)
	for _, d := range depts {
		for _, child := range deptMap[d.ID.String()].Children {
			if child.Type == "position" {
				posMap[child.ID] = child
			}
		}
	}

	for _, emp := range employees {
		empNode := &models.OrgTreeNode{
			ID:    emp.ID.String(),
			Label: emp.FullName,
			Type:  "employee",
			Meta: map[string]any{
				"employee_id": emp.EmployeeID,
				"join_date":   emp.JoinDate.Format("2006-01-02"),
			},
		}
		if emp.PositionID != nil {
			posID := emp.PositionID.String()
			if pos, ok := posMap[posID]; ok {
				pos.Children = append(pos.Children, empNode)
				continue
			}
		}
		// Fallback: attach to department directly
		if emp.DepartmentID != nil {
			deptID := emp.DepartmentID.String()
			if dept, ok := deptMap[deptID]; ok {
				dept.Children = append(dept.Children, empNode)
			}
		}
	}

	return roots, nil
}
