package models

type OrgTreeNode struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Type     string         `json:"type"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []*OrgTreeNode `json:"children,omitempty"`
}

type OrgTreeResponse struct {
	Tree []*OrgTreeNode `json:"tree"`
}
