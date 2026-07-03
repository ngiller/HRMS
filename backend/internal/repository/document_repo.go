package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListDocuments(ctx context.Context, page, perPage int, status, employeeID, docType string) ([]models.DocumentSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM employee_documents d WHERE d.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 0

	if status != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND d.status = $%d::doc_status", argIdx)
		args = append(args, status)
	}
	if employeeID != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND d.employee_id::text = $%d", argIdx)
		args = append(args, employeeID)
	}
	if docType != "" {
		argIdx++
		countQuery += fmt.Sprintf(" AND d.doc_type = $%d::doc_type", argIdx)
		args = append(args, docType)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT d.id, d.employee_id, COALESCE(e.full_name, ''),
			d.doc_type::text, COALESCE(d.file_name, ''), COALESCE(d.title, ''),
			d.status::text,
			COALESCE(to_char(d.expiry_date, 'YYYY-MM-DD'), ''),
			d.created_at
		FROM employee_documents d
		LEFT JOIN employees e ON e.id = d.employee_id
		WHERE d.deleted_at IS NULL
	`
	searchArgs := []interface{}{}
	if status != "" {
		query += fmt.Sprintf(" AND d.status = $%d::doc_status", len(searchArgs)+1)
		searchArgs = append(searchArgs, status)
	}
	if employeeID != "" {
		query += fmt.Sprintf(" AND d.employee_id::text = $%d", len(searchArgs)+1)
		searchArgs = append(searchArgs, employeeID)
	}
	if docType != "" {
		query += fmt.Sprintf(" AND d.doc_type = $%d::doc_type", len(searchArgs)+1)
		searchArgs = append(searchArgs, docType)
	}
	query += fmt.Sprintf(" ORDER BY d.created_at DESC LIMIT $%d OFFSET $%d", len(searchArgs)+1, len(searchArgs)+2)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var documents []models.DocumentSummary
	for rows.Next() {
		var d models.DocumentSummary
		if err := rows.Scan(&d.ID, &d.EmployeeID, &d.EmployeeName,
			&d.DocType, &d.FileName, &d.Title,
			&d.Status, &d.ExpiryDate, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		documents = append(documents, d)
	}
	return documents, total, nil
}

func GetDocumentByID(ctx context.Context, id string) (*models.EmployeeDocument, error) {
	query := `
		SELECT d.id, d.employee_id, COALESCE(e.full_name, ''),
			d.doc_type::text, COALESCE(d.file_name, ''), COALESCE(d.file_url, ''),
			d.file_size, COALESCE(d.mime_type, ''), COALESCE(d.title, ''), COALESCE(d.description, ''),
			d.status::text,
			d.verified_by, COALESCE(verifier.full_name, ''),
			d.verified_at, COALESCE(d.rejection_reason, ''),
			COALESCE(to_char(d.expiry_date, 'YYYY-MM-DD'), ''),
			d.is_required,
			d.created_at, d.updated_at, d.deleted_at
		FROM employee_documents d
		LEFT JOIN employees e ON e.id = d.employee_id
		LEFT JOIN employees verifier ON verifier.id = d.verified_by
		WHERE (d.id::text = $1) AND d.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var d models.EmployeeDocument
	err := row.Scan(
		&d.ID, &d.EmployeeID, &d.EmployeeName,
		&d.DocType, &d.FileName, &d.FileURL,
		&d.FileSize, &d.MimeType, &d.Title, &d.Description,
		&d.Status,
		&d.VerifiedBy, &d.VerifiedByName,
		&d.VerifiedAt, &d.RejectionReason,
		&d.ExpiryDate,
		&d.IsRequired,
		&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func CreateDocument(ctx context.Context, req *models.CreateDocumentReq) (*models.EmployeeDocument, error) {
	query := `
		INSERT INTO employee_documents (employee_id, doc_type, file_name, file_url, file_size, mime_type, title, description, expiry_date, status)
		VALUES ($1::uuid, $2::doc_type, $3, $4, $5, $6, $7, $8,
			NULLIF($9, '')::date, 'pending'::doc_status)
		RETURNING id, employee_id, '' as employee_name,
			doc_type::text, COALESCE(file_name, ''), COALESCE(file_url, ''),
			file_size, COALESCE(mime_type, ''), COALESCE(title, ''), COALESCE(description, ''),
			status::text,
			verified_by, '' as verified_by_name,
			verified_at, COALESCE(rejection_reason, ''),
			COALESCE(to_char(expiry_date, 'YYYY-MM-DD'), ''),
			is_required,
			created_at, updated_at, deleted_at
	`
	var d models.EmployeeDocument
	err := database.WithUserContext(ctx, req.EmployeeID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.EmployeeID, req.DocType, req.FileName, req.FileURL, req.FileSize,
			req.MimeType, req.Title, req.Description, req.ExpiryDate,
		)
		return row.Scan(
			&d.ID, &d.EmployeeID, &d.EmployeeName,
			&d.DocType, &d.FileName, &d.FileURL,
			&d.FileSize, &d.MimeType, &d.Title, &d.Description,
			&d.Status,
			&d.VerifiedBy, &d.VerifiedByName,
			&d.VerifiedAt, &d.RejectionReason,
			&d.ExpiryDate,
			&d.IsRequired,
			&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func UpdateDocumentStatus(ctx context.Context, id, status, rejectionReason, verifierID string) (*models.EmployeeDocument, error) {
	query := `
		UPDATE employee_documents
		SET status = $2::doc_status,
			rejection_reason = $3,
			verified_by = $4::uuid,
			verified_at = NOW()
		WHERE id::text = $1 AND deleted_at IS NULL
		AND status = 'pending'
		RETURNING id, employee_id, '' as employee_name,
			doc_type::text, COALESCE(file_name, ''), COALESCE(file_url, ''),
			file_size, COALESCE(mime_type, ''), COALESCE(title, ''), COALESCE(description, ''),
			status::text,
			verified_by, '' as verified_by_name,
			verified_at, COALESCE(rejection_reason, ''),
			COALESCE(to_char(expiry_date, 'YYYY-MM-DD'), ''),
			is_required,
			created_at, updated_at, deleted_at
	`
	var d models.EmployeeDocument
	err := database.WithUserContext(ctx, verifierID, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query, id, status, rejectionReason, verifierID)
		return row.Scan(
			&d.ID, &d.EmployeeID, &d.EmployeeName,
			&d.DocType, &d.FileName, &d.FileURL,
			&d.FileSize, &d.MimeType, &d.Title, &d.Description,
			&d.Status,
			&d.VerifiedBy, &d.VerifiedByName,
			&d.VerifiedAt, &d.RejectionReason,
			&d.ExpiryDate,
			&d.IsRequired,
			&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		)
	})
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func DeleteDocument(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE employee_documents SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}



