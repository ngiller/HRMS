package repository

import (
	"context"
	"encoding/json"
	"errors"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func GetCompany(ctx context.Context) (*models.Company, error) {
	query := `
		SELECT id, name, COALESCE(legal_name, ''), COALESCE(address, ''),
			COALESCE(city, ''), COALESCE(province, ''), COALESCE(postal_code, ''),
			COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''),
			COALESCE(npwp, ''), COALESCE(logo_url, ''),
			COALESCE(bpjs_ks_number, ''), COALESCE(bpjs_jht_number, ''), COALESCE(bpjs_jp_number, ''),
			COALESCE(bpjs_jkk_rate, 0.54),
			hr_settings, is_active,
			created_at, updated_at
		FROM companies
		WHERE deleted_at IS NULL
		LIMIT 1
	`

	var c models.Company
	var hrSettingsBytes []byte

	err := database.Pool.QueryRow(ctx, query).Scan(
		&c.ID, &c.Name, &c.LegalName, &c.Address,
		&c.City, &c.Province, &c.PostalCode,
		&c.Phone, &c.Email, &c.Website,
		&c.NPWP, &c.LogoURL,
		&c.BPJSKSNumber, &c.BPJSJHTNumber, &c.BPJSJPNumber,
		&c.BPJSJKKRate,
		&hrSettingsBytes, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Parse hr_settings JSONB
	if len(hrSettingsBytes) > 0 {
		json.Unmarshal(hrSettingsBytes, &c.HRSettings)
	}

	return &c, nil
}

func UpdateCompanySettings(ctx context.Context, req *models.UpdateCompanySettingsRequest) (*models.Company, error) {
	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	addStringField := func(col string, val *string) {
		if val != nil {
			setClauses = append(setClauses, col+" = $"+itoa(argIdx))
			args = append(args, *val)
			argIdx++
		}
	}

	addFloatField := func(col string, val *float64) {
		if val != nil {
			setClauses = append(setClauses, col+" = $"+itoa(argIdx))
			args = append(args, *val)
			argIdx++
		}
	}

	addStringField("name", req.Name)
	addStringField("legal_name", req.LegalName)
	addStringField("address", req.Address)
	addStringField("city", req.City)
	addStringField("province", req.Province)
	addStringField("postal_code", req.PostalCode)
	addStringField("phone", req.Phone)
	addStringField("email", req.Email)
	addStringField("npwp", req.NPWP)
	addStringField("bpjs_ks_number", req.BPJSKSNumber)
	addStringField("bpjs_jht_number", req.BPJSJHTNumber)
	addStringField("bpjs_jp_number", req.BPJSJPNumber)
	addFloatField("bpjs_jkk_rate", req.BPJSJKKRate)

	// Handle hr_settings JSONB merge
	if req.HRSettings != nil {
		hrSettingsJSON, err := json.Marshal(req.HRSettings)
		if err != nil {
			return nil, err
		}
		// Use jsonb_merge: COALESCE(hr_settings, '{}'::jsonb) || $N
		setClauses = append(setClauses, "hr_settings = COALESCE(hr_settings, '{}'::jsonb) || $"+itoa(argIdx)+"::jsonb")
		args = append(args, string(hrSettingsJSON))
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("tidak ada data yang diubah")
	}

	query := `UPDATE companies SET ` + joinStringsClause(setClauses, ", ") + `, updated_at = NOW()
		WHERE deleted_at IS NULL
		RETURNING id, name, COALESCE(legal_name, ''), COALESCE(address, ''),
			COALESCE(city, ''), COALESCE(province, ''), COALESCE(postal_code, ''),
			COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''),
			COALESCE(npwp, ''), COALESCE(logo_url, ''),
			COALESCE(bpjs_ks_number, ''), COALESCE(bpjs_jht_number, ''), COALESCE(bpjs_jp_number, ''),
			COALESCE(bpjs_jkk_rate, 0.54),
			hr_settings, is_active,
			created_at, updated_at
	`

	var c models.Company
	var hrSettingsBytes []byte

	err := database.Pool.QueryRow(ctx, query, args...).Scan(
		&c.ID, &c.Name, &c.LegalName, &c.Address,
		&c.City, &c.Province, &c.PostalCode,
		&c.Phone, &c.Email, &c.Website,
		&c.NPWP, &c.LogoURL,
		&c.BPJSKSNumber, &c.BPJSJHTNumber, &c.BPJSJPNumber,
		&c.BPJSJKKRate,
		&hrSettingsBytes, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("company tidak ditemukan")
		}
		return nil, err
	}

	if len(hrSettingsBytes) > 0 {
		json.Unmarshal(hrSettingsBytes, &c.HRSettings)
	}

	return &c, nil
}

// Helper: integer to string (replacement for strconv.Itoa)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

// Helper: join strings with separator
func joinStringsClause(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

type EmployeeTHRInfo struct {
	EmployeeID   string
	EmployeeName string
	JoinDate     string
	BaseSalary   float64
}

func GetAllEmployeesForTHR(ctx context.Context) ([]EmployeeTHRInfo, error) {
	query := `
		SELECT e.id::text, e.full_name, e.join_date::text,
			COALESCE((SELECT base_salary FROM employee_salary_histories
				WHERE employee_id = e.id AND deleted_at IS NULL
				ORDER BY effective_date DESC NULLS LAST, created_at DESC LIMIT 1), 0)
		FROM employees e
		WHERE e.deleted_at IS NULL AND e.is_active = TRUE
		ORDER BY e.full_name
	`

	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []EmployeeTHRInfo
	for rows.Next() {
		var emp EmployeeTHRInfo
		if err := rows.Scan(&emp.EmployeeID, &emp.EmployeeName, &emp.JoinDate, &emp.BaseSalary); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	return employees, nil
}

type EmployeeBPJSConfig struct {
	EmployeeID string           `json:"employee_id"`
	BPJSConfig *models.BPJSConfig `json:"bpjs_config"`
}

func GetEmployeeBPJSConfig(ctx context.Context, employeeID string) (*EmployeeBPJSConfig, error) {
	query := `
		SELECT id::text, bpjs_config
		FROM employees
		WHERE id::text = $1 AND deleted_at IS NULL
	`

	var cfg EmployeeBPJSConfig
	var bpjsBytes []byte

	err := database.Pool.QueryRow(ctx, query, employeeID).Scan(&cfg.EmployeeID, &bpjsBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		return nil, err
	}

	if len(bpjsBytes) > 0 {
		json.Unmarshal(bpjsBytes, &cfg.BPJSConfig)
	}

	return &cfg, nil
}

func UpdateEmployeeBPJSConfig(ctx context.Context, employeeID string, bpjsConfig *models.BPJSConfig) error {
	var query string
	var args []interface{}

	if bpjsConfig == nil {
		// Set to NULL to clear BPJS config (use company defaults)
		query = `UPDATE employees SET bpjs_config = NULL, updated_at = NOW()
			WHERE id::text = $1 AND deleted_at IS NULL`
		args = []interface{}{employeeID}
	} else {
		bpjsJSON, err := json.Marshal(bpjsConfig)
		if err != nil {
			return err
		}
		query = `UPDATE employees SET bpjs_config = $1::jsonb, updated_at = NOW()
			WHERE id::text = $2 AND deleted_at IS NULL`
		args = []interface{}{string(bpjsJSON), employeeID}
	}

	tag, err := database.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("karyawan tidak ditemukan")
	}

	return nil
}
