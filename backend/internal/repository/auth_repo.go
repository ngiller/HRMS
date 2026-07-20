package repository

import (
	"context"
	"time"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetEmployeeByEmail(ctx context.Context, email string) (*models.Employee, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email, e.password_hash,
			e.gender::text, COALESCE(e.birth_place, ''), e.birth_date, COALESCE(e.religion::text, ''),
			COALESCE(e.marital_status::text, ''), e.join_date, e.employment_status::text, e.is_active,
			e.role_id, COALESCE(r.slug, ''), COALESCE(r.name, ''),
			e.position_id, COALESCE(p.name, ''),
			e.department_id, COALESCE(d.name, ''),
			e.work_schedule_id, COALESCE(ws.name, ''),
			e.approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = e.approval_line_id), '') as approval_line_name,
			COALESCE(e.phone, ''), COALESCE(e.address_domicile, ''), COALESCE(e.photo_url, ''),
			COALESCE(decrypt_sensitive(e.encrypted_nik), ''),
			COALESCE(decrypt_sensitive(e.encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(e.encrypted_address_ktp), ''),
			COALESCE(e.ptkp_status::text, ''),
			e.is_pregnant,
			e.base_salary,
			e.daily_wage,
			e.last_login_at, e.is_locked, e.locked_until,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
	LEFT JOIN positions p ON e.position_id = p.id
	LEFT JOIN departments d ON e.department_id = d.id
	LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
	WHERE e.email = $1 AND e.deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, email)
	return scanEmployee(row)
}

func GetEmployeeByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	query := `
		SELECT e.id, e.employee_id, e.full_name, e.email, e.password_hash,
			e.gender::text, COALESCE(e.birth_place, ''), e.birth_date, COALESCE(e.religion::text, ''),
			COALESCE(e.marital_status::text, ''), e.join_date, e.employment_status::text, e.is_active,
			e.role_id, COALESCE(r.slug, ''), COALESCE(r.name, ''),
			e.position_id, COALESCE(p.name, ''),
			e.department_id, COALESCE(d.name, ''),
			e.work_schedule_id, COALESCE(ws.name, ''),
			e.approval_line_id,
			COALESCE((SELECT full_name FROM employees WHERE id = e.approval_line_id), '') as approval_line_name,
			COALESCE(e.phone, ''), COALESCE(e.address_domicile, ''), COALESCE(e.photo_url, ''),
			COALESCE(decrypt_sensitive(e.encrypted_nik), ''),
			COALESCE(decrypt_sensitive(e.encrypted_npwp), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_name), ''),
			COALESCE(decrypt_sensitive(e.encrypted_bank_account), ''),
			COALESCE(decrypt_sensitive(e.encrypted_address_ktp), ''),
			COALESCE(e.ptkp_status::text, ''),
			e.is_pregnant,
			e.base_salary,
			e.daily_wage,
			e.last_login_at, e.is_locked, e.locked_until,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN roles r ON e.role_id = r.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN departments d ON e.department_id = d.id
		LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
		WHERE e.id = $1 AND e.deleted_at IS NULL
	`

	row := database.Pool.QueryRow(ctx, query, id)
	return scanEmployee(row)
}

func UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE employees SET last_login_at = NOW() WHERE id = $1`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func CreatePasswordResetToken(ctx context.Context, employeeID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (employee_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, employeeID, token, expiresAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func ValidateResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	query := `
		SELECT employee_id FROM password_reset_tokens
		WHERE token = $1 AND is_used = FALSE AND expires_at > NOW()
	`
	var employeeID uuid.UUID
	err := database.Pool.QueryRow(ctx, query, token).Scan(&employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &employeeID, nil
}

func MarkResetTokenUsed(ctx context.Context, token string) error {
	query := `UPDATE password_reset_tokens SET is_used = TRUE WHERE token = $1`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, token)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func UpdatePassword(ctx context.Context, employeeID uuid.UUID, passwordHash string) error {
	query := `UPDATE employees SET password_hash = $1 WHERE id = $2`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, passwordHash, employeeID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func RecordLoginAttempt(ctx context.Context, employeeID uuid.UUID, ipAddress string, isSuccessful bool) error {
	query := `
		INSERT INTO login_attempts (employee_id, ip_address, is_successful)
		VALUES ($1, $2, $3)
	`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, employeeID, ipAddress, isSuccessful)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func GetRecentFailedAttempts(ctx context.Context, employeeID uuid.UUID, within time.Duration) (int, error) {
	query := `
		SELECT COUNT(*) FROM login_attempts
		WHERE employee_id = $1 AND is_successful = FALSE
		AND attempted_at > NOW() - $2::interval
	`
	var count int
	err := database.Pool.QueryRow(ctx, query, employeeID, within.String()).Scan(&count)
	return count, err
}

func LockEmployee(ctx context.Context, employeeID uuid.UUID, until time.Time) error {
	query := `UPDATE employees SET is_locked = TRUE, locked_until = $1 WHERE id = $2`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, until, employeeID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func StoreRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string, expiresAt time.Time) error {
	query := `
		INSERT INTO user_sessions (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
	`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, userID, refreshToken, expiresAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, is_active, expires_at
		FROM user_sessions
		WHERE refresh_token = $1 AND is_active = TRUE AND expires_at > NOW()
	`
	var s models.Session
	err := database.Pool.QueryRow(ctx, query, refreshToken).Scan(
		&s.ID, &s.UserID, &s.RefreshToken, &s.IsActive, &s.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func InvalidateSession(ctx context.Context, refreshToken string) error {
	query := `UPDATE user_sessions SET is_active = FALSE WHERE refresh_token = $1`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, refreshToken)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE user_sessions SET is_active = FALSE WHERE user_id = $1 AND is_active = TRUE`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func scanEmployee(row pgx.Row) (*models.Employee, error) {
	var e models.Employee
	err := row.Scan(
		&e.ID, &e.EmployeeID, &e.FullName, &e.Email, &e.PasswordHash,
		&e.Gender, &e.PlaceOfBirth, &e.DateOfBirth, &e.Religion,
		&e.MaritalStatus, &e.JoinDate, &e.EmploymentStatus, &e.IsActive,
		&e.RoleID, &e.RoleSlug, &e.RoleName,
		&e.PositionID, &e.PositionName,
		&e.DepartmentID, &e.DepartmentName,
		&e.WorkScheduleID, &e.WorkScheduleName,
		&e.ApprovalLineID, &e.ApprovalLineName,
		&e.Phone, &e.Address, &e.PhotoURL,
		&e.NIK, &e.NPWP, &e.BankName, &e.BankAccount, &e.AddressKTP, &e.PTKPStatus,
		&e.IsPregnant,
		&e.BaseSalary, &e.DailyWage,
		&e.LastLoginAt, &e.IsLocked, &e.LockedUntil,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}
