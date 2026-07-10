package repository

import (
	"context"
	"fmt"

	"hrms-backend/internal/database"
	"hrms-backend/internal/models"

	"github.com/jackc/pgx/v5"
)

func ListAnnouncements(ctx context.Context, page, perPage int, announcementType string) ([]models.AnnouncementSummary, int, error) {
	countQuery := `SELECT COUNT(*) FROM announcements a WHERE a.deleted_at IS NULL`
	args := []interface{}{}
	if announcementType != "" {
		countQuery += fmt.Sprintf(" AND a.announcement_type = $1::announcement_type")
		args = append(args, announcementType)
	}

	var total int
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT a.id, a.title, a.content, a.announcement_type::text,
			a.target_all, a.is_pinned,
			COALESCE(e.full_name, ''),
			a.published_at, a.expired_at,
			a.created_at
		FROM announcements a
		LEFT JOIN employees e ON e.id = a.created_by
		WHERE a.deleted_at IS NULL
	`
	searchArgs := []interface{}{}
	if announcementType != "" {
		query += fmt.Sprintf(" AND a.announcement_type = $%d::announcement_type", len(searchArgs)+1)
		searchArgs = append(searchArgs, announcementType)
	}
	query += fmt.Sprintf(" ORDER BY a.is_pinned DESC, a.pin_priority DESC, a.published_at DESC LIMIT $%d OFFSET $%d", len(searchArgs)+1, len(searchArgs)+2)
	allArgs := append(searchArgs, perPage, offset)

	rows, err := database.Pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var announcements []models.AnnouncementSummary
	for rows.Next() {
		var a models.AnnouncementSummary
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.AnnouncementType,
			&a.TargetAll, &a.IsPinned,
			&a.CreatedByName,
			&a.PublishedAt, &a.ExpiredAt,
			&a.CreatedAt); err != nil {
			return nil, 0, err
		}
		announcements = append(announcements, a)
	}
	return announcements, total, nil
}

func GetAnnouncementByID(ctx context.Context, id string) (*models.Announcement, error) {
	query := `
		SELECT a.id, a.title, a.content, a.announcement_type::text,
			a.target_department_id, COALESCE(d.name, ''),
			a.target_position_grade_id,
			a.target_all, COALESCE(a.attachment_urls, '{}'),
			a.is_pinned, a.pin_priority,
			a.published_at, a.expired_at,
			a.created_by, COALESCE(e.full_name, ''),
			a.created_at, a.updated_at, a.deleted_at,
			(SELECT COUNT(*) FROM announcement_reads ar WHERE ar.announcement_id = a.id) as read_count
		FROM announcements a
		LEFT JOIN employees e ON e.id = a.created_by
		LEFT JOIN departments d ON d.id = a.target_department_id
		WHERE (a.id::text = $1) AND a.deleted_at IS NULL
	`
	row := database.Pool.QueryRow(ctx, query, id)
	var a models.Announcement
	err := row.Scan(
		&a.ID, &a.Title, &a.Content, &a.AnnouncementType,
		&a.TargetDepartmentID, &a.TargetDepartmentName,
		&a.TargetPositionGradeID,
		&a.TargetAll, &a.AttachmentURLs,
		&a.IsPinned, &a.PinPriority,
		&a.PublishedAt, &a.ExpiredAt,
		&a.CreatedBy, &a.CreatedByName,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		&a.ReadCount,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateAnnouncement(ctx context.Context, createdBy string, req *models.CreateAnnouncementReq) (*models.Announcement, error) {
	query := `
		INSERT INTO announcements (title, content, announcement_type, target_department_id, target_all, attachment_urls, is_pinned, pin_priority, published_at, expired_at, created_by)
		VALUES ($1, $2, $3::announcement_type,
			NULLIF($4, '')::uuid, $5, $6, $7, $8, NOW(), NULLIF($9, '')::timestamptz, $10::uuid)
		RETURNING id, title, content, announcement_type::text,
			target_department_id, '' as target_department_name,
			target_position_grade_id,
			target_all, COALESCE(attachment_urls, '{}'),
			is_pinned, pin_priority,
			published_at, expired_at,
			created_by, '' as created_by_name,
			created_at, updated_at, deleted_at,
			0 as read_count
	`
	var a models.Announcement
	err := database.WithUserContext(ctx, createdBy, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, query,
			req.Title, req.Content, req.AnnouncementType, req.TargetDepartmentID,
			req.TargetAll, req.AttachmentURLs, req.IsPinned, req.PinPriority,
			req.ExpiredAt, createdBy,
		)
		return row.Scan(
			&a.ID, &a.Title, &a.Content, &a.AnnouncementType,
			&a.TargetDepartmentID, &a.TargetDepartmentName,
			&a.TargetPositionGradeID,
			&a.TargetAll, &a.AttachmentURLs,
			&a.IsPinned, &a.PinPriority,
			&a.PublishedAt, &a.ExpiredAt,
			&a.CreatedBy, &a.CreatedByName,
			&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
			&a.ReadCount,
		)
	})
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func UpdateAnnouncement(ctx context.Context, id, userID string, req *models.UpdateAnnouncementReq) (*models.Announcement, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, req.Title)
		argIdx++
	}
	if req.Content != "" {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argIdx))
		args = append(args, req.Content)
		argIdx++
	}
	if req.AnnouncementType != "" {
		setClauses = append(setClauses, fmt.Sprintf("announcement_type = $%d::announcement_type", argIdx))
		args = append(args, req.AnnouncementType)
		argIdx++
	}
	if req.TargetAll != nil {
		setClauses = append(setClauses, fmt.Sprintf("target_all = $%d", argIdx))
		args = append(args, *req.TargetAll)
		argIdx++
	}
	if req.TargetDepartmentID != "" {
		setClauses = append(setClauses, fmt.Sprintf("target_department_id = NULLIF($%d, '')::uuid", argIdx))
		args = append(args, req.TargetDepartmentID)
		argIdx++
	}
	if req.AttachmentURLs != nil {
		setClauses = append(setClauses, fmt.Sprintf("attachment_urls = $%d", argIdx))
		args = append(args, req.AttachmentURLs)
		argIdx++
	}
	if req.IsPinned != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_pinned = $%d", argIdx))
		args = append(args, *req.IsPinned)
		argIdx++
	}
	if req.PinPriority != nil {
		setClauses = append(setClauses, fmt.Sprintf("pin_priority = $%d", argIdx))
		args = append(args, *req.PinPriority)
		argIdx++
	}
	if req.ExpiredAt != "" {
		setClauses = append(setClauses, fmt.Sprintf("expired_at = NULLIF($%d, '')::timestamptz", argIdx))
		args = append(args, req.ExpiredAt)
		argIdx++
	}

	if len(setClauses) == 0 {
		return GetAnnouncementByID(ctx, id)
	}

	var a *models.Announcement
	err := database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		query := fmt.Sprintf(`
			UPDATE announcements SET %s, updated_at = NOW()
			WHERE id::text = $%d AND deleted_at IS NULL
			RETURNING id, title, content, announcement_type::text,
				target_department_id, '' as target_department_name,
				target_position_grade_id,
				target_all, COALESCE(attachment_urls, '{}'),
				is_pinned, pin_priority,
				published_at, expired_at,
				created_by, '' as created_by_name,
				created_at, updated_at, deleted_at,
				0 as read_count
		`, joinStrings(setClauses, ", "), argIdx)
		args = append(args, id)

		row := tx.QueryRow(ctx, query, args...)
		var result models.Announcement
		if err := row.Scan(
			&result.ID, &result.Title, &result.Content, &result.AnnouncementType,
			&result.TargetDepartmentID, &result.TargetDepartmentName,
			&result.TargetPositionGradeID,
			&result.TargetAll, &result.AttachmentURLs,
			&result.IsPinned, &result.PinPriority,
			&result.PublishedAt, &result.ExpiredAt,
			&result.CreatedBy, &result.CreatedByName,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
			&result.ReadCount,
		); err != nil {
			return err
		}
		a = &result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func DeleteAnnouncement(ctx context.Context, id, userID string) error {
	return database.WithUserContext(ctx, userID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE announcements SET deleted_at = NOW() WHERE id::text = $1 AND deleted_at IS NULL`, id)
		return err
	})
}

// MarkAnnouncementRead marks an announcement as read by an employee.
func MarkAnnouncementRead(ctx context.Context, announcementID, employeeID string) error {
	query := `
		INSERT INTO announcement_reads (announcement_id, employee_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (announcement_id, employee_id) DO NOTHING
	`
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query, announcementID, employeeID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
