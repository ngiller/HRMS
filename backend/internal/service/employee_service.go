package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hrms-backend/internal/config"
	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type EmployeeService struct{}

func NewEmployeeService() *EmployeeService {
	return &EmployeeService{}
}

func (s *EmployeeService) ListEmployees(ctx context.Context, page, perPage int, search, departmentID, status string, includeDeleted bool) (*models.EmployeeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	employees, total, err := repository.ListEmployees(ctx, page, perPage, search, departmentID, status, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data karyawan: %w", err)
	}

	if employees == nil {
		employees = []models.EmployeeSummary{}
	}

	return &models.EmployeeListResponse{
		Employees: employees,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *EmployeeService) GetEmployee(ctx context.Context, id string) (*models.Employee, error) {
	employee, err := repository.GetEmployeeByIDRepo(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memuat data karyawan")
	}
	if employee == nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	return employee, nil
}

func (s *EmployeeService) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	stats, err := repository.GetDashboardStats(ctx)
	if err != nil {
		return nil, errors.New("gagal memuat data dashboard")
	}
	return stats, nil
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, req *models.CreateEmployeeRequest, userID string) (*models.Employee, error) {
	// Validasi
	if req.FullName == "" {
		return nil, errors.New("nama lengkap harus diisi")
	}
	if req.Email == "" {
		return nil, errors.New("email harus diisi")
	}
	if req.Password == "" {
		return nil, errors.New("password harus diisi")
	}
	if req.Gender == "" {
		return nil, errors.New("jenis kelamin harus diisi")
	}
	if req.EmploymentStatus == "" {
		return nil, errors.New("status karyawan harus diisi")
	}
	if req.JoinDate == "" {
		return nil, errors.New("tanggal bergabung harus diisi")
	}

	// Cek duplikasi email
	exists, err := repository.CheckEmailExists(ctx, req.Email, "")
	if err != nil {
		return nil, errors.New("gagal memvalidasi email")
	}
	if exists {
		return nil, errors.New("email sudah digunakan")
	}

	// Cek duplikasi employee_id
	if req.EmployeeID != "" {
		exists, err = repository.CheckEmployeeIDExists(ctx, req.EmployeeID, "")
		if err != nil {
			return nil, errors.New("gagal memvalidasi NIK")
		}
		if exists {
			return nil, errors.New("NIK karyawan sudah digunakan")
		}
	}

	employee, err := repository.CreateEmployee(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat karyawan: %w", err)
	}
	return employee, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id string, req *models.UpdateEmployeeRequest, userID string) (*models.Employee, error) {
	// Cek apakah karyawan ada
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// Cek duplikasi email (jika diubah)
	if req.Email != "" {
		exists, err = repository.CheckEmailExists(ctx, req.Email, id)
		if err != nil {
			return nil, errors.New("gagal memvalidasi email")
		}
		if exists {
			return nil, errors.New("email sudah digunakan")
		}
	}

	employee, err := repository.UpdateEmployee(ctx, id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate karyawan: %w", err)
	}
	return employee, nil
}

func (s *EmployeeService) GetEmployeeHistory(ctx context.Context, id string, page, perPage int) (*models.EmployeeHistoryListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	histories, total, err := repository.GetEmployeeHistory(ctx, id, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat riwayat karyawan: %w", err)
	}

	if histories == nil {
		histories = []models.EmployeeHistory{}
	}

	return &models.EmployeeHistoryListResponse{
		Histories: histories,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

func (s *EmployeeService) LogHistory(ctx context.Context, employeeID, changeType string, oldValue, newValue map[string]any, reason, changedBy string) error {
	return repository.CreateEmployeeHistory(ctx, employeeID, changeType, oldValue, newValue, reason, changedBy)
}

func (s *EmployeeService) ExportEmployees(ctx context.Context) ([]byte, error) {
	employees, err := repository.ListEmployeesForExportFull(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data karyawan: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Data Karyawan"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Employee ID", "Full Name", "Barcode",
		"Organization", "Job Position", "Job Level",
		"Join Date", "Resign Date", "Status Employee",
		"End Date", "Sign Date",
		"Email", "Birth Date", "Age",
		"Birth Place", "Citizen ID Address", "Residential Address",
		"NPWP", "PTKP Status", "Employee Tax Status",
		"Tax Config",
		"Bank Name", "Bank Account", "Bank Account Holder",
		"BPJS Ketenagakerjaan", "BPJS Kesehatan",
		"NIK (NPWP 16 Digit)",
		"Mobile Phone", "Phone",
		"Branch Name", "Parent Branch Name",
		"Religion", "Gender", "Marital Status",
		"Blood Type", "Nationality Code", "Currency",
		"Length Of Service", "Payment Schedule",
		"Approval Line", "Manager", "Grade",
		"Class", "Profile Picture",
		"Cost Center", "Cost Center Category", "SBU",
		"NPWP 16 digit (new)",
		"Passport", "Passport Expiration Date",
		"Jenis Dok. Referensi Bukti Potong",
		"Nomor Dok. Referensi Bukti Potong",
		"Tanggal Dok. Referensi Bukti Potong",
		"TIN (Taxpayer Identification Number)",
		"Ukuran Baju",
	}

	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	f.SetRowStyle(sheet, 1, 1, boldStyle)

	for i, emp := range employees {
		row := i + 2
		vals := []interface{}{
			emp.EmployeeID, emp.FullName, emp.Barcode,
			emp.Organization, emp.JobPosition, emp.JobLevel,
			emp.JoinDate, emp.ResignDate, emp.StatusEmployee,
			emp.EndDate, emp.SignDate,
			emp.Email, emp.BirthDate, emp.Age,
			emp.BirthPlace, emp.CitizenIDAddress, emp.ResidentialAddress,
			emp.NPWP, emp.PTKPStatus, emp.EmployeeTaxStatus,
			emp.TaxConfig,
			emp.BankName, emp.BankAccount, emp.BankAccountHolder,
			emp.BPJSTK, emp.BPJSKesehatan,
			emp.NIK,
			emp.MobilePhone, emp.Phone,
			emp.BranchName, emp.ParentBranchName,
			emp.Religion, emp.Gender, emp.MaritalStatus,
			emp.BloodType, emp.NationalityCode, emp.Currency,
			emp.LengthOfService, emp.PaymentSchedule,
			emp.ApprovalLine, emp.Manager, emp.Grade,
			emp.Class, emp.ProfilePicture,
			emp.CostCenter, emp.CostCenterCategory, emp.SBU,
			emp.NPWPBaru,
			emp.Passport, emp.PassportExpirationDate,
			emp.JenisDokReferensi,
			emp.NomorDokReferensi,
			emp.TanggalDokReferensi,
			emp.TIN,
			emp.UkuranBaju,
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

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

func (s *EmployeeService) ImportEmployees(ctx context.Context, file *multipart.FileHeader, userID string) (*models.ImportResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file Excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet: %w", err)
	}

	result := &models.ImportResult{}
	if len(rows) < 2 {
		result.Message = "Tidak ada data untuk diimpor"
		return result, nil
	}

	// Pre-load lookup data for resolving names to IDs
	deptList, _ := repository.GetAllDepartments(ctx)
	posList, _ := repository.GetAllPositions(ctx)
	gradeList, _ := repository.GetAllPositionGrades(ctx)

	deptMap := make(map[string]string)
	for _, d := range deptList {
		deptMap[strings.ToLower(d.Name)] = d.ID.String()
	}
	posMap := make(map[string]string)
	for _, p := range posList {
		posMap[strings.ToLower(p.Name)] = p.ID.String()
	}
	gradeMap := make(map[string]string)
	for _, g := range gradeList {
		gradeMap[strings.ToLower(g.Name)] = g.ID.String()
	}

	for i := range rows[1:] {
		rowNum := i + 2

		cells := make([]string, 56)
		for c := 0; c < 56; c++ {
			colName, _ := excelize.ColumnNumberToName(c + 1)
			val, _ := f.GetCellValue(sheet, fmt.Sprintf("%s%d", colName, rowNum))
			cells[c] = strings.TrimSpace(val)
		}

		// 55-column mapping (indices match employee_template.xlsx)
		// 0=EmployeeID, 1=FullName, 2=Barcode, 3=Organization, 4=Job Position, 5=Job Level
		// 6=JoinDate, 7=ResignDate, 8=Status Employee, 9=EndDate
		// 11=Email, 12=BirthDate, 14=BirthPlace
		// 15=CitizenIDAddr, 16=ResidentialAddr, 17=NPWP, 18=PTKPStatus
		// 21=BankName, 22=BankAccount, 23=BankAccountHolder
		// 26=NIK, 27=MobilePhone, 28=Phone
		// 31=Religion, 32=Gender, 33=MaritalStatus, 34=BloodType
		// 39=ApprovalLine
		employeeID := cells[0]
		fullName := cells[1]
		orgName := cells[3]         // Organization (department)
		jobPositionName := cells[4]  // Job Position
		jobLevelName := cells[5]     // Job Level (grade)
		joinDate := cells[6]
		empStatusStr := cells[8]
		endDateStr := cells[9]
		email := cells[11]
		dateOfBirth := cells[12]
		placeOfBirth := cells[14]
		nik := cells[26]          // NIK (NPWP 16 Digit)
		addressKTP := cells[15]    // Citizen ID Address
		address := cells[16]       // Residential Address
		npwp := cells[17]          // NPWP
		ptkpStr := cells[18]       // PTKP Status
		bankName := cells[21]      // Bank Name
		bankAccount := cells[22]   // Bank Account
		phone := cells[27]         // Mobile Phone
		religion := cells[31]      // Religion
		genderStr := cells[32]     // Gender
		maritalStatus := cells[33] // Marital Status
		bloodType := cells[34]     // Blood Type
		approvalLineName := cells[39] // Approval Line

		if fullName == "" || email == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: nama dan email wajib diisi", rowNum))
			continue
		}

		gender := "laki_laki"
		g := strings.ToLower(genderStr)
		if g == "perempuan" || g == "p" || g == "female" || g == "f" {
			gender = "perempuan"
		}

		empStatus := "percobaan"
		s := strings.ToLower(empStatusStr)
		if strings.Contains(s, "tetap") || strings.Contains(s, "permanent") {
			empStatus = "tetap"
		} else if strings.Contains(s, "kontrak") || strings.Contains(s, "contract") {
			empStatus = "kontrak"
		} else if strings.Contains(s, "magang") || strings.Contains(s, "intern") {
			empStatus = "percobaan"
		}

		// Normalize marital status: DB enum values are: lajang, menikah, cerai_hidup, cerai_mati
		maritalStatus = strings.ToLower(maritalStatus)
		if strings.Contains(maritalStatus, "single") || strings.Contains(maritalStatus, "belum menikah") || strings.Contains(maritalStatus, "lajang") {
			maritalStatus = "lajang"
		} else if strings.Contains(maritalStatus, "married") || strings.Contains(maritalStatus, "menikah") {
			maritalStatus = "menikah"
		} else if strings.Contains(maritalStatus, "widow") || strings.Contains(maritalStatus, "cerai") {
			maritalStatus = "cerai_hidup"
		} else {
			maritalStatus = "lajang"
		}

		religion = strings.ToLower(religion)
		if strings.Contains(religion, "christian") || strings.Contains(religion, "kristen") {
			religion = "kristen"
		} else if strings.Contains(religion, "catholic") || strings.Contains(religion, "katolik") {
			religion = "katolik"
		} else if strings.Contains(religion, "moslem") || strings.Contains(religion, "muslim") || strings.Contains(religion, "islam") {
			religion = "islam"
		} else if strings.Contains(religion, "hindu") {
			religion = "hindu"
		} else if strings.Contains(religion, "buddh") {
			religion = "buddha"
		} else if strings.Contains(religion, "kong") {
			religion = "konghucu"
		} else {
			religion = "lainnya"
		}
		maritalStatus = strings.ReplaceAll(maritalStatus, " ", "_")
		maritalStatus = strings.ReplaceAll(maritalStatus, "-", "_")
		switch maritalStatus {
		case "belum_menikah", "single", "lajang":
			maritalStatus = "lajang"
		case "menikah", "married", "kawin":
			maritalStatus = "menikah"
		case "cerai_hidup", "cerai", "divorced":
			maritalStatus = "cerai_hidup"
		case "cerai_mati", "widow", "widower", "janda", "duda":
			maritalStatus = "cerai_mati"
		}

		// Normalize PTKP Status (case-insensitive)
		ptkpStatus := strings.ToUpper(ptkpStr)
		if ptkpStatus != "" {
			ptkpStatus = strings.ReplaceAll(ptkpStatus, " ", "")
			ptkpStatus = strings.ReplaceAll(ptkpStatus, "/", "") // TK/0 -> TK0
		}

		if joinDate == "" {
			joinDate = time.Now().Format("2006-01-02")
		} else if len(joinDate) >= 10 {
			joinDate = joinDate[:10]
		}

		// Normalize end date
		if endDateStr != "" && len(endDateStr) >= 10 {
			endDateStr = endDateStr[:10]
		}

		if employeeID == "" {
			employeeID = fmt.Sprintf("IMP-%d", time.Now().UnixNano())
		}

		// Resolve department, position, grade by name using pre-loaded maps
		// Also tries English→Indonesian translation and substring matching
		deptID := resolveDepartmentID(deptMap, orgName)
		if deptID == "" && orgName != "" {
			req := &models.CreateDepartmentRequest{
				Name: orgName,
				Code: fmt.Sprintf("DPT-%d", time.Now().UnixMilli()%1000000),
			}
			if newDept, err := repository.CreateDepartment(ctx, req, userID); err == nil && newDept != nil {
				deptID = newDept.ID.String()
				deptMap[strings.ToLower(orgName)] = deptID
			}
		}

		posID := resolvePositionID(posMap, jobPositionName)
		if posID == "" && jobPositionName != "" {
			req := &models.CreatePositionRequest{
				Name: jobPositionName,
			}
			if newPos, err := repository.CreatePosition(ctx, req, userID); err == nil && newPos != nil {
				posID = newPos.ID.String()
				posMap[strings.ToLower(jobPositionName)] = posID
			}
		}

		gradeID := resolveGradeID(gradeMap, jobLevelName)
		if gradeID == "" && jobLevelName != "" {
			req := &models.CreatePositionGradeRequest{
				Name: jobLevelName,
			}
			if newGrade, err := repository.CreatePositionGrade(ctx, req, userID); err == nil && newGrade != nil {
				gradeID = newGrade.ID.String()
				gradeMap[strings.ToLower(jobLevelName)] = gradeID
			}
		}

		// Resolve approval line by employee name (lookup on-the-fly)
		approvalLineID := ""
		if approvalLineName != "" {
			if id, err := repository.GetEmployeeByName(ctx, approvalLineName); err == nil && id != nil {
				approvalLineID = *id
			}
		}

		req := &models.CreateEmployeeRequest{
			EmployeeID:       employeeID,
			FullName:         fullName,
			Email:            email,
			Password:         "password123",
			Gender:           gender,
			EmploymentStatus: empStatus,
			JoinDate:         joinDate,
			Phone:            phone,
			PlaceOfBirth:     placeOfBirth,
			DateOfBirth:      dateOfBirth,
			Religion:         normalizeReligion(religion),
			MaritalStatus:    maritalStatus,
			Address:          address,
			AddressKTP:       addressKTP,
			NIK:              nik,
			NPWP:             npwp,
			BankName:         bankName,
			BankAccount:      bankAccount,
			DepartmentID:     deptID,
			PositionID:       posID,
			PositionGradeID:  gradeID,
			BloodType:        bloodType,
			PTKPStatus:       ptkpStatus,
			EndDate:          endDateStr,
			ApprovalLineID:   approvalLineID,
		}

		_, err := repository.CreateEmployee(ctx, req, userID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: %v", rowNum, err))
			continue
		}
		result.Success++
	}

	result.Message = fmt.Sprintf("Berhasil mengimpor %d dari %d data", result.Success, len(rows[1:]))
	return result, nil
}

// resolveDepartmentID tries exact match, then translation map, then substring match
func resolveDepartmentID(deptMap map[string]string, name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(name))
	
	// 1. Exact match
	if id, ok := deptMap[key]; ok {
		return id
	}
	
	// 2. English → Indonesian translation map
	translate := map[string]string{
		"accounting":            "keuangan",
		"sales":                 "penjualan",
		"marketing":             "pemasaran",
		"hrd":                   "sumber daya manusia",
		"human resources":       "sumber daya manusia",
		"human resources & general": "sumber daya manusia",
		"it":                    "teknologi informasi",
		"information technology": "teknologi informasi",
		"technical":             "teknologi informasi",
		"development":           "teknologi informasi",
		"engineering":           "teknologi informasi",
		"director":              "direksi",
		"finance":               "keuangan",
		"financial":             "keuangan",
		"logistic":              "",
		"general affair":        "sumber daya manusia",
	}
	if translated, ok := translate[key]; ok && translated != "" {
		if id, ok := deptMap[translated]; ok {
			return id
		}
	}
	
	// 3. Substring: check if any DB dept name contains the key or vice versa
	for dbName, id := range deptMap {
		if strings.Contains(dbName, key) || strings.Contains(key, dbName) {
			return id
		}
	}
	
	return ""
}

// resolvePositionID tries exact match, then substring match
func resolvePositionID(posMap map[string]string, name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(name))
	
	// 1. Exact match
	if id, ok := posMap[key]; ok {
		return id
	}
	
	// 2. English → Indonesian translation
	translate := map[string]string{
		"accounting":              "accounting",
		"pre sales":               "sales",
		"inbound sales":           "sales",
		"in house sales":          "sales",
		"sales admin":             "sales",
		"sales manager":           "sales manager",
		"web programmer":          "staff it",
		"engineer":                "staff it",
		"office boy":              "",
		"driver":                  "",
		"purchasing":              "",
		"store officer":           "",
		"hrd":                     "hr staff",
		"human resources & accounting": "hr staff",
		"marketing":               "marketing",
		"project manager":         "manager it",
		"business development manager": "marketing manager",
		"project coordinator":     "",
		"project admin":           "",
		"project sales leader":    "sales manager",
		"retail sales coordinator": "sales",
		"assistant head of technical": "staff it",
		"head of technical and development": "it director",
		"director":                "",
	}
	if translated, ok := translate[key]; ok && translated != "" {
		if id, ok := posMap[translated]; ok {
			return id
		}
	}
	
	// 3. Substring match
	for dbName, id := range posMap {
		if strings.Contains(dbName, key) || strings.Contains(key, dbName) {
			return id
		}
	}
	
	return ""
}

// resolveGradeID tries exact match, then translation, then substring
func resolveGradeID(gradeMap map[string]string, name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(name))
	
	// 1. Exact match
	if id, ok := gradeMap[key]; ok {
		return id
	}
	
	// 2. English → Indonesian translation
	translate := map[string]string{
		"owner":       "director",
		"direktur":    "director",
		"vp":          "senior manager",
		"head":        "manager",
	}
	if translated, ok := translate[key]; ok && translated != "" {
		if id, ok := gradeMap[translated]; ok {
			return id
		}
	}
	
	// 3. Substring match
	for dbName, id := range gradeMap {
		if strings.Contains(dbName, key) || strings.Contains(key, dbName) {
			return id
		}
	}
	
	return ""
}

func (s *EmployeeService) UpdateWorkSchedule(ctx context.Context, id, workScheduleID, userID string) (*models.Employee, error) {
	// Validate employee exists
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return nil, errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// Validate work schedule exists (if non-empty)
	if workScheduleID != "" {
		ws, err := repository.GetWorkScheduleByID(ctx, workScheduleID)
		if err != nil || ws == nil {
			return nil, errors.New("jadwal kerja tidak ditemukan")
		}
	}

	employee, err := repository.UpdateEmployeeWorkSchedule(ctx, id, workScheduleID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengupdate jadwal kerja: %w", err)
	}
	return employee, nil
}

// RegisterFaceDescriptor stores a face descriptor for an employee
func (s *EmployeeService) RegisterFaceDescriptor(ctx context.Context, id, descriptorJSON, userID string) error {
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return errors.New("karyawan tidak ditemukan")
	}

	return repository.UpdateFaceDescriptor(ctx, id, descriptorJSON, userID)
}

func (s *EmployeeService) UploadPhoto(ctx context.Context, id string, file *multipart.FileHeader, userID string) (string, error) {
	// Check employee exists
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return "", errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return "", errors.New("karyawan tidak ditemukan")
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	photoFilename := "employee_" + id + ext
	uploadDir := config.Load().UploadDir

	// Save file to disk
	savePath := filepath.Join(uploadDir, photoFilename)
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("gagal menulis file: %w", err)
	}

	photoURL := "/uploads/" + photoFilename

	// Update database
	if err := repository.UpdateEmployeePhoto(ctx, id, photoURL, userID); err != nil {
		return "", fmt.Errorf("gagal memperbarui foto: %w", err)
	}

	return photoURL, nil
}
func (s *EmployeeService) RestoreEmployee(ctx context.Context, id string, userID string) error {
	return repository.RestoreEmployee(ctx, id, userID)
}

func (s *EmployeeService) GetManagerDashboard(ctx context.Context, userID string) (*models.ManagerDashboardResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("ID pengguna tidak valid")
	}
	stats, err := repository.GetManagerDashboardStats(ctx, uid)
	if err != nil {
		return nil, errors.New("gagal memuat dashboard manajer")
	}
	return stats, nil
}

func (s *EmployeeService) GetHRDashboard(ctx context.Context) (*models.HRDashboardResponse, error) {
	stats, err := repository.GetHRDashboardStats(ctx)
	if err != nil {
		return nil, errors.New("gagal memuat dashboard HR")
	}
	return stats, nil
}

func (s *EmployeeService) DeleteEmployee(ctx context.Context, id string, userID string) error {
	exists, err := repository.CheckEmployeeExists(ctx, id)
	if err != nil {
		return errors.New("gagal memvalidasi karyawan")
	}
	if !exists {
		return errors.New("karyawan tidak ditemukan")
	}

	return repository.DeleteEmployee(ctx, id, userID)
}

// normalizeReligion maps various religion inputs to the DB enum values
func normalizeReligion(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "islam", "moslem", "muslim":
		return "islam"
	case "kristen", "christian", "protestan":
		return "kristen"
	case "katolik", "catholic":
		return "katolik"
	case "hindu":
		return "hindu"
	case "buddha", "buddhist", "budha":
		return "buddha"
	case "konghucu", "confucius", "confucian":
		return "konghucu"
	default:
		return v
	}
}
