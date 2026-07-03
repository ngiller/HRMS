package repository

import (
	"context"
	"time"

	"hrms-backend/internal/database"

	"github.com/google/uuid"
)

type DeptTreeNode struct {
	ID       uuid.UUID
	Name     string
	ParentID *uuid.UUID
	HeadName string
}

type PosTreeNode struct {
	ID           uuid.UUID
	Name         string
	DepartmentID uuid.UUID
	GradeName    string
}

type EmpTreeNode struct {
	ID           uuid.UUID
	FullName     string
	EmployeeID   string
	PositionID   *uuid.UUID
	DepartmentID *uuid.UUID
	JoinDate     time.Time
}

func GetAllDeptsForTree(ctx context.Context) ([]DeptTreeNode, error) {
	query := `
		SELECT d.id, d.name, d.parent_id,
			COALESCE(e.full_name, '') as head_name
		FROM departments d
		LEFT JOIN employees e ON d.head_id = e.id
		WHERE d.deleted_at IS NULL AND d.is_active = TRUE
		ORDER BY d.sort_order ASC, d.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []DeptTreeNode
	for rows.Next() {
		var d DeptTreeNode
		if err := rows.Scan(&d.ID, &d.Name, &d.ParentID, &d.HeadName); err != nil {
			return nil, err
		}
		depts = append(depts, d)
	}
	return depts, nil
}

func GetAllPositionsForTree(ctx context.Context) ([]PosTreeNode, error) {
	query := `
		SELECT p.id, p.name, p.department_id,
			COALESCE(pg.name, '') as grade_name
		FROM positions p
		LEFT JOIN position_grades pg ON p.grade_id = pg.id
		WHERE p.deleted_at IS NULL AND p.is_active = TRUE
		ORDER BY p.name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []PosTreeNode
	for rows.Next() {
		var p PosTreeNode
		if err := rows.Scan(&p.ID, &p.Name, &p.DepartmentID, &p.GradeName); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, nil
}

func GetAllEmployeesForTree(ctx context.Context) ([]EmpTreeNode, error) {
	query := `
		SELECT e.id, e.full_name, e.employee_id, e.position_id, e.department_id, e.join_date
		FROM employees e
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		ORDER BY e.full_name ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []EmpTreeNode
	for rows.Next() {
		var emp EmpTreeNode
		if err := rows.Scan(&emp.ID, &emp.FullName, &emp.EmployeeID, &emp.PositionID, &emp.DepartmentID, &emp.JoinDate); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}
