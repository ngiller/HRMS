package service

import (
	"context"
	"fmt"
	"sort"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type OrganizationService struct{}

func NewOrganizationService() *OrganizationService {
	return &OrganizationService{}
}

// GetTree returns the organization tree filtered by user role
func (s *OrganizationService) GetTree(ctx context.Context, roleSlug, userID string) ([]*models.OrgTreeNode, error) {
	// Build full tree
	roots, err := s.buildFullTree(ctx)
	if err != nil {
		return nil, err
	}

	// super_admin, director, hr_manager see everything
	if roleSlug == "super_admin" || roleSlug == "director" || roleSlug == "hr_manager" {
		return roots, nil
	}

	// manager, hr_staff, finance → only own department
	if roleSlug == "manager" || roleSlug == "hr_staff" || roleSlug == "finance" {
		return s.filterByDepartment(ctx, roots, userID)
	}

	// employee → only own chain
	return s.filterByEmployeeChain(ctx, roots, userID)
}

// filterByDepartment returns only the subtree of the user's department
func (s *OrganizationService) filterByDepartment(ctx context.Context, roots []*models.OrgTreeNode, userID string) ([]*models.OrgTreeNode, error) {
	// Get user's employee info
	emp, err := repository.GetEmployeeByIDRepo(ctx, userID)
	if err != nil || emp == nil || emp.DepartmentID == nil {
		// Fallback: return full tree
		return roots, nil
	}
	deptID := emp.DepartmentID.String()

	// Walk roots to find the user's department
	for _, root := range roots {
		if root.ID == deptID {
			return []*models.OrgTreeNode{root}, nil
		}
		// Check children recursively
		if found := findDeptInChildren(root, deptID); found != nil {
			// Found the department as a child — return the department node
			// stripped of sibling departments, but keep the parent chain
			parentNode := &models.OrgTreeNode{
				ID:       root.ID,
				Label:    root.Label,
				Type:     root.Type,
				Meta:     root.Meta,
				Children: []*models.OrgTreeNode{found},
			}
			return []*models.OrgTreeNode{parentNode}, nil
		}
	}

	// Department not found in tree — return full tree as fallback
	return roots, nil
}

// findDeptInChildren recursively searches for a department node by ID
func findDeptInChildren(node *models.OrgTreeNode, deptID string) *models.OrgTreeNode {
	if node.ID == deptID && node.Type == "department" {
		return node
	}
	for _, child := range node.Children {
		if found := findDeptInChildren(child, deptID); found != nil {
			return found
		}
	}
	return nil
}

// filterByEmployeeChain returns only the user's atasan chain, themselves, and their bawahans
func (s *OrganizationService) filterByEmployeeChain(ctx context.Context, roots []*models.OrgTreeNode, userID string) ([]*models.OrgTreeNode, error) {
	// Build a flat map of all nodes for quick lookup
	allNodes := make(map[string]*models.OrgTreeNode)
	collectAllNodes(roots, allNodes)

	empNode, ok := allNodes[userID]
	if !ok || empNode.Type != "employee" {
		return roots, nil // fallback
	}

	// Collect the chain: atasan → user → bawahan
	chainIDs := make(map[string]bool)

	// Add user
	chainIDs[empNode.ID] = true

	// Add all subordinates (children of the employee node)
	addSubordinates(empNode, chainIDs)

	// Add atasan chain (walk up approval_line)
	atasanIDs := collectAtasanChain(empNode, allNodes)
	for _, id := range atasanIDs {
		chainIDs[id] = true
	}

	// Build filtered tree from the chain
	return buildChainTree(roots, chainIDs), nil
}

// collectAllNodes flattens the tree into a map by ID
func collectAllNodes(nodes []*models.OrgTreeNode, result map[string]*models.OrgTreeNode) {
	for _, n := range nodes {
		result[n.ID] = n
		if n.Children != nil {
			collectAllNodes(n.Children, result)
		}
	}
}

// addSubordinates recursively adds all child employee nodes
func addSubordinates(node *models.OrgTreeNode, chainIDs map[string]bool) {
	for _, child := range node.Children {
		chainIDs[child.ID] = true
		addSubordinates(child, chainIDs)
	}
}

// collectAtasanChain walks up the approval_line to find all superiors
func collectAtasanChain(empNode *models.OrgTreeNode, allNodes map[string]*models.OrgTreeNode) []string {
	var atasanIDs []string
	visited := make(map[string]bool)

	current := empNode
	for {
		atasanID := ""
		if current.Meta != nil {
			if id, ok := current.Meta["atasan_id"]; ok {
				atasanID = fmt.Sprintf("%v", id)
			}
		}
		if atasanID == "" || visited[atasanID] {
			break
		}
		visited[atasanID] = true
		atasanIDs = append(atasanIDs, atasanID)

		atasanNode, ok := allNodes[atasanID]
		if !ok || atasanNode.Type != "employee" {
			break
		}
		current = atasanNode
	}

	return atasanIDs
}

// buildChainTree rebuilds a minimal tree containing only the chain IDs
func buildChainTree(roots []*models.OrgTreeNode, chainIDs map[string]bool) []*models.OrgTreeNode {
	var result []*models.OrgTreeNode

	for _, root := range roots {
		filtered := filterNodeChain(root, chainIDs)
		if filtered != nil {
			result = append(result, filtered)
		}
	}

	return result
}

// filterNodeChain recursively filters a tree to keep only chain IDs
func filterNodeChain(node *models.OrgTreeNode, chainIDs map[string]bool) *models.OrgTreeNode {
	// Check if this node is in the chain
	if !chainIDs[node.ID] {
		// Not in chain — check if any children are
		var keptChildren []*models.OrgTreeNode
		for _, child := range node.Children {
			if filtered := filterNodeChain(child, chainIDs); filtered != nil {
				keptChildren = append(keptChildren, filtered)
			}
		}
		if len(keptChildren) == 0 {
			return nil
		}
		// Keep this node as a container for children in the chain
		return &models.OrgTreeNode{
			ID:       node.ID,
			Label:    node.Label,
			Type:     node.Type,
			Meta:     node.Meta,
			Children: keptChildren,
		}
	}

	// This node IS in the chain
	keptChildren := make([]*models.OrgTreeNode, 0, len(node.Children))
	for _, child := range node.Children {
		if filtered := filterNodeChain(child, chainIDs); filtered != nil {
			keptChildren = append(keptChildren, filtered)
		}
	}

	return &models.OrgTreeNode{
		ID:       node.ID,
		Label:    node.Label,
		Type:     node.Type,
		Meta:     node.Meta,
		Children: keptChildren,
	}
}

// buildFullTree builds the complete organization tree without filtering
func (s *OrganizationService) buildFullTree(ctx context.Context) ([]*models.OrgTreeNode, error) {
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
			posNode.Meta["department_id"] = deptID
			roots = append(roots, posNode)
		}
	}

	// Build position map for attaching employees
	posMap := make(map[string]*models.OrgTreeNode)
	for _, d := range depts {
		for _, child := range deptMap[d.ID.String()].Children {
			if child.Type == "position" {
				posMap[child.ID] = child
			}
		}
	}

	// Build employee nodes with hierarchy info
	empNodes := make(map[string]*models.OrgTreeNode)

	for _, emp := range employees {
		departmentName := ""
		if emp.DepartmentID != nil {
			if dept, ok := deptMap[emp.DepartmentID.String()]; ok {
				departmentName = dept.Label
			}
		}

		meta := map[string]any{
			"employee_id":     emp.EmployeeID,
			"join_date":       emp.JoinDate.Format("2006-01-02"),
			"department_name": departmentName,
		}

		// Add atasan (superior) info
		if emp.ApprovalLineID != nil {
			meta["atasan_id"] = emp.ApprovalLineID.String()
			meta["atasan_nama"] = emp.ApprovalLineName
		}

		node := &models.OrgTreeNode{
			ID:    emp.ID.String(),
			Label: emp.FullName,
			Type:  "employee",
			Meta:  meta,
		}
		empNodes[emp.ID.String()] = node
	}

	// Mark department heads by ID (not by name)
	headMap := make(map[string]bool)
	for _, d := range depts {
		if d.HeadID != nil {
			headMap[d.HeadID.String()] = true
		}
	}

	// Attach employees to positions (or departments if no position)
	var attachedEmps = make(map[string]bool)
	for _, emp := range employees {
		empNode := empNodes[emp.ID.String()]
		// Mark as department head if applicable
		if headMap[emp.ID.String()] {
			empNode.Meta["is_department_head"] = true
		}

		if emp.PositionID != nil {
			posID := emp.PositionID.String()
			if pos, ok := posMap[posID]; ok {
				pos.Children = append(pos.Children, empNode)
				attachedEmps[emp.ID.String()] = true
				continue
			}
		}
		// Fallback: attach to department directly
		if emp.DepartmentID != nil {
			deptID := emp.DepartmentID.String()
			if dept, ok := deptMap[deptID]; ok {
				dept.Children = append(dept.Children, empNode)
				attachedEmps[emp.ID.String()] = true
			}
		}
	}

	// Sort children: department heads first, then by name
	sortOrgChildren(roots)

	return roots, nil
}

// sortOrgChildren recursively sorts tree node children:
// 1. Employee nodes: atasan (supervisor) comes before their bawahan
// 2. Department heads first
// 3. Alphabetically by label for same-level employees
// 4. Other nodes: alphabetically by label
func sortOrgChildren(nodes []*models.OrgTreeNode) {
	// Build atasan lookup: map of employee ID -> set of subordinate IDs within this group
	atasanMap := make(map[string]map[string]bool)
	for _, n := range nodes {
		if n.Type == "employee" && n.Meta != nil {
			if atasanID, ok := n.Meta["atasan_id"]; ok {
				id := fmt.Sprintf("%v", atasanID)
				if atasanMap[id] == nil {
					atasanMap[id] = make(map[string]bool)
				}
				atasanMap[id][n.ID] = true
			}
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		// Non-employee nodes sorted by label
		if nodes[i].Type != "employee" || nodes[j].Type != "employee" {
			return nodes[i].Label < nodes[j].Label
		}

		// Both are employees
		// Check atasan-bawahan relationship
		if atasanMap[nodes[i].ID] != nil && atasanMap[nodes[i].ID][nodes[j].ID] {
			return true // i is atasan of j, i comes first
		}
		if atasanMap[nodes[j].ID] != nil && atasanMap[nodes[j].ID][nodes[i].ID] {
			return false // j is atasan of i, j comes first
		}

		// Check department head status
		iHead := false
		if nodes[i].Meta != nil {
			if h, ok := nodes[i].Meta["is_department_head"]; ok {
				iHead = h.(bool)
			}
		}
		jHead := false
		if nodes[j].Meta != nil {
			if h, ok := nodes[j].Meta["is_department_head"]; ok {
				jHead = h.(bool)
			}
		}
		if iHead != jHead {
			return iHead // heads first
		}

		// Alphabetical by name
		return nodes[i].Label < nodes[j].Label
	})

	for _, node := range nodes {
		if len(node.Children) > 0 {
			sortOrgChildren(node.Children)
		}
	}
}
