package models

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeDocument — full document object
type EmployeeDocument struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeName    string     `json:"employee_name"`
	DocType         string     `json:"doc_type"`
	FileName        string     `json:"file_name"`
	FileURL         string     `json:"file_url"`
	FileSize        *int       `json:"file_size"`
	MimeType        string     `json:"mime_type"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	VerifiedBy      *uuid.UUID `json:"verified_by"`
	VerifiedByName  string     `json:"verified_by_name"`
	VerifiedAt      *time.Time `json:"verified_at"`
	RejectionReason string     `json:"rejection_reason"`
	ExpiryDate      *string    `json:"expiry_date"`
	IsRequired      bool       `json:"is_required"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// DocumentSummary — for list views
type DocumentSummary struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	DocType      string    `json:"doc_type"`
	FileName     string    `json:"file_name"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	ExpiryDate   *string   `json:"expiry_date"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateDocumentReq struct {
	EmployeeID  string `json:"employee_id"`
	DocType     string `json:"doc_type"`
	FileName    string `json:"file_name"`
	FileURL     string `json:"file_url"`
	FileSize    *int   `json:"file_size,omitempty"`
	MimeType    string `json:"mime_type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ExpiryDate  string `json:"expiry_date,omitempty"`
}

type UpdateDocumentStatusReq struct {
	RejectionReason string `json:"rejection_reason"`
}

type DocumentListResponse struct {
	Documents []DocumentSummary `json:"documents"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PerPage   int               `json:"per_page"`
}

// DocType labels for frontend
var DocTypeLabels = map[string]string{
	"ktp":        "KTP",
	"kk":         "Kartu Keluarga",
	"ijazah":     "Ijazah",
	"sertifikat": "Sertifikat",
	"kontrak":    "Kontrak Kerja",
	"npwp":       "NPWP",
	"bpjs":       "BPJS",
	"photo":      "Foto",
	"other":      "Lainnya",
}
