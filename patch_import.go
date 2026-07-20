package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("backend/internal/service/employee_service.go")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	text := string(content)

	// find the start of ImportEmployees
	start := strings.Index(text, "func (s *EmployeeService) ImportEmployees(ctx context.Context, file *multipart.FileHeader, userID string) (*models.ImportResult, error) {")
	if start == -1 {
		fmt.Println("Could not find ImportEmployees")
		return
	}

	// find the end of ImportEmployees (the next func declaration)
	end := strings.Index(text[start+5:], "func (s *EmployeeService)")
	if end == -1 {
		end = len(text) - start
	} else {
		end += start + 5
	}

	replacement := `func (s *EmployeeService) ImportEmployees(ctx context.Context, file *multipart.FileHeader, userID string) (*models.ImportResult, error) {
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

	for i := range rows[1:] {
		rowNum := i + 2

		cells := make([]string, 56)
		for c := 0; c < 56; c++ {
			colName, _ := excelize.ColumnNumberToName(c + 1)
			val, _ := f.GetCellValue(sheet, fmt.Sprintf("%s%d", colName, rowNum))
			cells[c] = strings.TrimSpace(val)
		}

		employeeID := cells[0]
		fullName := cells[1]
		joinDate := cells[6]
		empStatusStr := cells[8]
		email := cells[11]
		dateOfBirth := cells[12]
		placeOfBirth := cells[14]
		nik := cells[15]
		addressKTP := cells[16]
		address := cells[17]
		npwp := cells[18]
		bankName := cells[22]
		bankAccount := cells[23]
		phone := cells[28]
		religion := cells[32]
		genderStr := cells[33]
		maritalStatus := cells[34]

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
			empStatus = "magang"
		}

		maritalStatus = strings.ToLower(maritalStatus)

		if joinDate == "" {
			joinDate = time.Now().Format("2006-01-02")
		} else if len(joinDate) >= 10 {
			joinDate = joinDate[:10] // assuming YYYY-MM-DD
		}

		if employeeID == "" {
			employeeID = fmt.Sprintf("IMP-%d", time.Now().UnixNano())
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
			Religion:         strings.ToLower(religion),
			MaritalStatus:    maritalStatus,
			Address:          address,
			AddressKTP:       addressKTP,
			NIK:              nik,
			NPWP:             npwp,
			BankName:         bankName,
			BankAccount:      bankAccount,
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

`
	newText := text[:start] + replacement + text[end:]

	err = ioutil.WriteFile("backend/internal/service/employee_service.go", []byte(newText), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Println("Successfully patched ImportEmployees")
}
