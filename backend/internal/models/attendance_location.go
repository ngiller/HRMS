package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceLocation struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	RadiusMeters int        `json:"radius_meters"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type AttendanceLocationSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateAttendanceLocationRequest struct {
	Name         string  `json:"name" validate:"required"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude" validate:"required"`
	Longitude    float64 `json:"longitude" validate:"required"`
	RadiusMeters int     `json:"radius_meters"`
}

type UpdateAttendanceLocationRequest struct {
	Name         *string  `json:"name"`
	Address      *string  `json:"address"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	RadiusMeters *int     `json:"radius_meters"`
	IsActive     *bool    `json:"is_active"`
}

type AttendanceLocationListResponse struct {
	AttendanceLocations []AttendanceLocationSummary `json:"attendance_locations"`
	Total               int                         `json:"total"`
	Page                int                         `json:"page"`
	PerPage             int                         `json:"per_page"`
}
