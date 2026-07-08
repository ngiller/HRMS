package models

import (
	"time"

	"github.com/google/uuid"
)

type BPJSConfig struct {
	Kesehatan *BPJSComponentConfig `json:"kesehatan,omitempty"`
	JHT       *BPJSComponentConfig `json:"jht,omitempty"`
	JP        *BPJSComponentConfig `json:"jp,omitempty"`
	JKK       *BPJSComponentConfig `json:"jkk,omitempty"`
	JKM       *BPJSComponentConfig `json:"jkm,omitempty"`
}

type BPJSComponentConfig struct {
	Enabled      *bool    `json:"enabled,omitempty"`
	EmployeeRate *float64 `json:"employee_rate,omitempty"`
	CompanyRate  *float64 `json:"company_rate,omitempty"`
	Ceiling      *float64 `json:"ceiling,omitempty"`
}

type Company struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	LegalName      string    `json:"legal_name,omitempty"`
	Address        string    `json:"address,omitempty"`
	City           string    `json:"city,omitempty"`
	Province       string    `json:"province,omitempty"`
	PostalCode     string    `json:"postal_code,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	Website        string    `json:"website,omitempty"`
	NPWP           string    `json:"npwp,omitempty"`
	LogoURL        string    `json:"logo_url,omitempty"`
	BPJSKSNumber   string    `json:"bpjs_ks_number,omitempty"`
	BPJSJHTNumber  string    `json:"bpjs_jht_number,omitempty"`
	BPJSJPNumber   string    `json:"bpjs_jp_number,omitempty"`
	BPJSJKKRate    float64   `json:"bpjs_jkk_rate"`
	HRSettings     HRSettings `json:"hr_settings"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type HRSettings struct {
	BPJS *BPJSConfig `json:"bpjs,omitempty"`
}

type UpdateCompanySettingsRequest struct {
	Name          *string    `json:"name,omitempty"`
	LegalName     *string    `json:"legal_name,omitempty"`
	Address       *string    `json:"address,omitempty"`
	City          *string    `json:"city,omitempty"`
	Province      *string    `json:"province,omitempty"`
	PostalCode    *string    `json:"postal_code,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Email         *string    `json:"email,omitempty"`
	NPWP          *string    `json:"npwp,omitempty"`
	BPJSKSNumber  *string    `json:"bpjs_ks_number,omitempty"`
	BPJSJHTNumber *string    `json:"bpjs_jht_number,omitempty"`
	BPJSJPNumber  *string    `json:"bpjs_jp_number,omitempty"`
	BPJSJKKRate   *float64   `json:"bpjs_jkk_rate,omitempty"`
	HRSettings    *HRSettings `json:"hr_settings,omitempty"`
}
