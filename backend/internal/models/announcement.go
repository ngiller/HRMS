package models

import (
	"time"

	"github.com/google/uuid"
)

// Announcement — full announcement object
type Announcement struct {
	ID                    uuid.UUID  `json:"id"`
	Title                 string     `json:"title"`
	Content               string     `json:"content"`
	AnnouncementType      string     `json:"announcement_type"`
	TargetDepartmentID    *uuid.UUID `json:"target_department_id"`
	TargetDepartmentName  string     `json:"target_department_name,omitempty"`
	TargetPositionGradeID *uuid.UUID `json:"target_position_grade_id"`
	TargetAll             bool       `json:"target_all"`
	AttachmentURLs        []string   `json:"attachment_urls"`
	IsPinned              bool       `json:"is_pinned"`
	PinPriority           int        `json:"pin_priority"`
	PublishedAt           time.Time  `json:"published_at"`
	ExpiredAt             *time.Time `json:"expired_at"`
	CreatedBy             uuid.UUID  `json:"created_by"`
	CreatedByName         string     `json:"created_by_name"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty"`
	ReadCount             int        `json:"read_count,omitempty"`
}

// AnnouncementSummary — for list views
type AnnouncementSummary struct {
	ID               uuid.UUID  `json:"id"`
	Title            string     `json:"title"`
	AnnouncementType string     `json:"announcement_type"`
	TargetAll        bool       `json:"target_all"`
	IsPinned         bool       `json:"is_pinned"`
	CreatedByName    string     `json:"created_by_name"`
	PublishedAt      time.Time  `json:"published_at"`
	ExpiredAt        *time.Time `json:"expired_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type CreateAnnouncementReq struct {
	Title              string   `json:"title"`
	Content            string   `json:"content"`
	AnnouncementType   string   `json:"announcement_type"`
	TargetDepartmentID string   `json:"target_department_id,omitempty"`
	TargetAll          bool     `json:"target_all"`
	AttachmentURLs     []string `json:"attachment_urls,omitempty"`
	IsPinned           bool     `json:"is_pinned"`
	PinPriority        int      `json:"pin_priority,omitempty"`
	ExpiredAt          string   `json:"expired_at,omitempty"`
}

type UpdateAnnouncementReq struct {
	Title              string   `json:"title,omitempty"`
	Content            string   `json:"content,omitempty"`
	AnnouncementType   string   `json:"announcement_type,omitempty"`
	TargetDepartmentID string   `json:"target_department_id,omitempty"`
	TargetAll          *bool    `json:"target_all,omitempty"`
	AttachmentURLs     []string `json:"attachment_urls,omitempty"`
	IsPinned           *bool    `json:"is_pinned,omitempty"`
	PinPriority        *int     `json:"pin_priority,omitempty"`
	ExpiredAt          string   `json:"expired_at,omitempty"`
}

type AnnouncementListResponse struct {
	Announcements []AnnouncementSummary `json:"announcements"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
}

// AnnouncementType labels
var AnnouncementTypeLabels = map[string]string{
	"general":   "Umum",
	"important": "Penting",
	"emergency": "Darurat",
}
