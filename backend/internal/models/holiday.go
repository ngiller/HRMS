package models

import (
	"time"

	"github.com/google/uuid"
)

// CompanyHoliday — full holiday object
type CompanyHoliday struct {
	ID                uuid.UUID  `json:"id"`
	Date              string     `json:"date"`
	Name              string     `json:"name"`
	HolidayType       string     `json:"holiday_type"`
	IsRecurringYearly bool       `json:"is_recurring_yearly"`
	Description       string     `json:"description"`
	IsActive          bool       `json:"is_active"`
	CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
	CreatedByName     string     `json:"created_by_name,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateHolidayReq struct {
	Date              string `json:"date"`
	Name              string `json:"name"`
	HolidayType       string `json:"holiday_type"`
	IsRecurringYearly bool   `json:"is_recurring_yearly"`
	Description       string `json:"description,omitempty"`
}

type UpdateHolidayReq struct {
	Date              *string `json:"date,omitempty"`
	Name              *string `json:"name,omitempty"`
	HolidayType       *string `json:"holiday_type,omitempty"`
	IsRecurringYearly *bool   `json:"is_recurring_yearly,omitempty"`
	Description       *string `json:"description,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

type HolidayListResponse struct {
	Holidays []CompanyHoliday `json:"holidays"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PerPage  int              `json:"per_page"`
}

// HolidayYearResponse — list by year for calendar view
type HolidayYearResponse struct {
	Year     int              `json:"year"`
	Holidays []CompanyHoliday `json:"holidays"`
}

// HolidayType labels
var HolidayTypeLabels = map[string]string{
	"national": "Libur Nasional",
	"joint":    "Cuti Bersama",
	"company":  "Libur Perusahaan",
}
