package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

type AttendanceRecordService struct{}

func NewAttendanceRecordService() *AttendanceRecordService {
	return &AttendanceRecordService{}
}

func (s *AttendanceRecordService) GetTodayStatus(ctx context.Context, employeeID string) (*models.TodayAttendanceStatus, error) {
	record, err := repository.GetTodayAttendanceByEmployee(ctx, employeeID)
	if err != nil {
		return nil, errors.New("gagal memuat status absensi hari ini")
	}

	_, scheduleName, startTime, endTime, err := repository.GetEmployeeScheduleInfo(ctx, employeeID)
	if err != nil {
		scheduleName = "Tidak ada jadwal"
	}

	status := &models.TodayAttendanceStatus{
		ScheduleName:  scheduleName,
		ScheduleStart: startTime,
		ScheduleEnd:   endTime,
		HasCheckedIn:  false,
		HasCheckedOut: false,
	}

	if record != nil {
		status.HasCheckedIn = record.CheckInTime != nil
		status.HasCheckedOut = record.CheckOutTime != nil
		status.Record = &models.AttendanceRecordSummary{
			ID:                  record.ID,
			Date:                record.Date,
			CheckInTime:         record.CheckInTime,
			CheckOutTime:        record.CheckOutTime,
			Status:              record.Status,
			IsLate:              record.IsLate,
			LateMinutes:         record.LateMinutes,
			TotalWorkHours:      record.TotalWorkHours,
			CheckInLocationName: record.CheckInLocationName,
		}
	}

	return status, nil
}

func (s *AttendanceRecordService) CheckIn(ctx context.Context, employeeID string, req *models.CheckInRequest) (*models.AttendanceRecord, error) {
	existing, err := repository.GetTodayAttendanceByEmployee(ctx, employeeID)
	if err != nil {
		return nil, errors.New("gagal memvalidasi absensi hari ini")
	}
	if existing != nil && existing.CheckInTime != nil {
		return nil, errors.New("sudah melakukan check-in hari ini")
	}

	scheduleID, _, startTime, _, err := repository.GetEmployeeScheduleInfo(ctx, employeeID)
	if err != nil {
		return nil, errors.New("gagal memuat jadwal kerja")
	}
	if scheduleID == nil {
		return nil, errors.New("tidak memiliki jadwal kerja, hubungi HR")
	}

	locID, locName := s.validateGPS(ctx, req.Lat, req.Lng)
	if req.LocationName != nil && *req.LocationName != "" {
		locName = *req.LocationName
	}
	if locID == nil && req.LocationID != nil && *req.LocationID != "" {
		locID = req.LocationID
	}
	var locNamePtr *string
	if locName != "" {
		locNamePtr = &locName
	}

	isLate := false
	lateMinutes := 0
	if startTime != "" {
		schedTime, parseErr := time.Parse("15:04", startTime)
		if parseErr == nil {
			now := time.Now()
			checkTime := time.Date(now.Year(), now.Month(), now.Day(), schedTime.Hour(), schedTime.Minute(), 0, 0, now.Location())
			if checkTime.Before(now) {
				lateMinutes = int(now.Sub(checkTime).Minutes())
				if lateMinutes > 0 {
					isLate = true
				}
			}
		}
	}

	record, err := repository.CreateCheckIn(ctx, employeeID, *scheduleID, req.Lat, req.Lng, locID, locNamePtr, isLate, lateMinutes)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan check-in: %w", err)
	}

	return record, nil
}

func (s *AttendanceRecordService) CheckOut(ctx context.Context, employeeID string, req *models.CheckOutRequest) (*models.AttendanceRecord, error) {
	record, err := repository.GetTodayAttendanceByEmployee(ctx, employeeID)
	if err != nil {
		return nil, errors.New("gagal memvalidasi absensi hari ini")
	}
	if record == nil {
		return nil, errors.New("belum melakukan check-in hari ini")
	}
	if record.CheckOutTime != nil {
		return nil, errors.New("sudah melakukan check-out hari ini")
	}

	locID, locName := s.validateGPS(ctx, req.Lat, req.Lng)
	if req.LocationName != nil && *req.LocationName != "" {
		locName = *req.LocationName
	}
	if locID == nil && req.LocationID != nil && *req.LocationID != "" {
		locID = req.LocationID
	}
	var locNamePtr *string
	if locName != "" {
		locNamePtr = &locName
	}

	var totalWorkHours *float64
	if record.CheckInTime != nil {
		now := time.Now()
		duration := now.Sub(*record.CheckInTime).Hours()
		duration = math.Round(duration*100) / 100
		totalWorkHours = &duration
	}

	updated, err := repository.UpdateCheckOut(ctx, record.ID.String(), req.Lat, req.Lng, locID, locNamePtr, totalWorkHours)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan check-out: %w", err)
	}

	return updated, nil
}

func (s *AttendanceRecordService) ListMyHistory(ctx context.Context, employeeID string, page, perPage int) (*models.AttendanceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	records, total, err := repository.ListMyAttendance(ctx, employeeID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat riwayat absensi: %w", err)
	}

	return &models.AttendanceListResponse{
		Records: records,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *AttendanceRecordService) ListReport(ctx context.Context, page, perPage int, deptID, status, dateFrom, dateTo string) (*models.AttendanceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	records, total, err := repository.ListAttendanceReport(ctx, page, perPage, deptID, status, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat laporan absensi: %w", err)
	}

	return &models.AttendanceListResponse{
		Records: records,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s *AttendanceRecordService) ExportReport(ctx context.Context, deptID, status, dateFrom, dateTo string) ([]byte, error) {
	// Get all records without pagination
	records, _, err := repository.ListAttendanceReport(ctx, 1, 100000, deptID, status, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data laporan absensi: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Laporan Absensi"
	f.SetSheetName("Sheet1", sheet)

	// Headers
	headers := []string{"No", "Tanggal", "Hari", "Nama Karyawan", "Departemen", "Check In", "Check Out", "Status", "Terlambat (menit)", "Total Jam Kerja", "Lokasi Check In"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Bold headers
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	for i, r := range records {
		row := i + 2
		checkIn := ""
		if r.CheckInTime != nil {
			checkIn = r.CheckInTime.Format("15:04")
		}
		checkOut := ""
		if r.CheckOutTime != nil {
			checkOut = r.CheckOutTime.Format("15:04")
		}
		totalHours := 0.0
		if r.TotalWorkHours != nil {
			totalHours = *r.TotalWorkHours
		}
		lateMin := 0
		if r.IsLate {
			lateMin = r.LateMinutes
		}
		locName := ""
		if r.CheckInLocationName != nil {
			locName = *r.CheckInLocationName
		}

		vals := []interface{}{
			i + 1,
			r.Date.Format("2006-01-02"),
			r.DayName,
			r.EmployeeName,
			r.DepartmentName,
			checkIn,
			checkOut,
			r.Status,
			lateMin,
			totalHours,
			locName,
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// Auto-width columns
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 20)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal menulis file Excel: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *AttendanceRecordService) validateGPS(ctx context.Context, lat, lng *float64) (locationID *string, locationName string) {
	if lat == nil || lng == nil {
		return nil, ""
	}

	locations, err := repository.GetActiveAttendanceLocations(ctx)
	if err != nil || len(locations) == 0 {
		return nil, ""
	}

	for _, loc := range locations {
		distance := haversine(*lat, *lng, loc.Latitude, loc.Longitude)
		distanceMeters := distance * 1000
		if distanceMeters <= float64(loc.RadiusMeters) {
			id := loc.ID.String()
			return &id, loc.Name
		}
	}

	return nil, "Luar area absensi"
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	r := 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return r * c
}
