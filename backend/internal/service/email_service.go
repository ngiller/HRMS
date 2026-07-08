package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"hrms-backend/internal/models"
	"hrms-backend/internal/repository"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	smtpFrom     string
	fromName     string
	enabled      bool
}

func NewEmailService(smtpHost, smtpPort, smtpUser, smtpPassword, smtpFrom, fromName string) *EmailService {
	enabled := smtpHost != "" && smtpUser != ""
	if enabled {
		log.Println("📧 Email Service: Enabled (SMTP " + smtpHost + ":" + smtpPort + ")")
	} else {
		log.Println("📧 Email Service: Disabled (set SMTP_HOST & SMTP_USER to enable)")
	}
	return &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		smtpFrom:     smtpFrom,
		fromName:     fromName,
		enabled:      enabled,
	}
}

func (s *EmailService) IsEnabled() bool {
	return s.enabled
}

// Send sends an email
func (s *EmailService) Send(to, subject, htmlBody string) error {
	if !s.enabled {
		log.Printf("📧 Email not sent (SMTP disabled): To=%s Subject=%s", to, subject)
		return nil
	}

	from := s.smtpFrom
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.smtpFrom)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + htmlBody

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	addr := s.smtpHost + ":" + s.smtpPort

	err := smtp.SendMail(addr, auth, s.smtpFrom, strings.Split(to, ","), []byte(msg))
	if err != nil {
		return fmt.Errorf("gagal mengirim email: %w", err)
	}

	log.Printf("📧 Email sent: To=%s Subject=%s", to, subject)
	return nil
}

// SendNotification sends an email notification based on a CreateNotificationRequest
func (s *EmailService) SendNotification(ctx context.Context, req *models.CreateNotificationRequest, toEmail string) error {
	if !s.enabled || toEmail == "" {
		return nil
	}

	// Build HTML body
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"></head>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<div style="text-align: center; margin-bottom: 20px;">
					<h2 style="color: #1A56DB; margin: 0;">HRMS</h2>
					<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 15px 0;">
				</div>
				<h3 style="color: #1f2937; margin-top: 0;">%s</h3>
				<p style="color: #4b5563; line-height: 1.6; font-size: 14px;">%s</p>
				<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">
					Ini adalah email otomatis dari sistem HRMS. Mohon tidak membalas email ini.
				</p>
			</div>
		</body>
		</html>
	`, req.Title, req.Body)

	return s.Send(toEmail, req.Title, htmlBody)
}

// SendApprovalRequest sends email notification for approval requests
func (s *EmailService) SendApprovalRequest(toEmail, entityLabel, requestorName string, linkURL string) error {
	if !s.enabled || toEmail == "" {
		return nil
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"></head>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h2 style="color: #1A56DB;">Pengajuan Baru Perlu Disetujui</h2>
				<p style="color: #4b5563;">Halo,</p>
				<p style="color: #4b5563;">%s mengajukan <strong>%s</strong> yang memerlukan persetujuan Anda.</p>
				%s
				<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">
					Ini adalah email otomatis dari sistem HRMS.
				</p>
			</div>
		</body>
		</html>
	`, requestorName, entityLabel, s.makeButtonHTML("Lihat Pengajuan", linkURL))

	return s.Send(toEmail, "Pengajuan Baru: "+entityLabel, htmlBody)
}

// SendNotificationWithLink sends an email with a link to the app
func (s *EmailService) SendNotificationWithLink(ctx context.Context, req *models.CreateNotificationRequest, toEmail, linkText, linkURL string) error {
	if !s.enabled || toEmail == "" {
		return nil
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"></head>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<div style="text-align: center; margin-bottom: 20px;">
					<h2 style="color: #1A56DB; margin: 0;">HRMS</h2>
					<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 15px 0;">
				</div>
				<h3 style="color: #1f2937; margin-top: 0;">%s</h3>
				<p style="color: #4b5563; line-height: 1.6; font-size: 14px;">%s</p>
				%s
				<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">
					Ini adalah email otomatis dari sistem HRMS.
				</p>
			</div>
		</body>
		</html>
	`, req.Title, req.Body, s.makeButtonHTML(linkText, linkURL))

	return s.Send(toEmail, req.Title, htmlBody)
}

func (s *EmailService) makeButtonHTML(text, url string) string {
	if url == "" {
		return ""
	}
	return fmt.Sprintf(`
		<div style="text-align: center; margin: 25px 0;">
			<a href="%s" style="background-color: #1A56DB; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">%s</a>
		</div>
	`, url, text)
}

// Update repository notification to mark is_email_sent = true
func MarkNotificationEmailSent(ctx context.Context, notificationID string) {
	_ = repository.UpdateNotificationEmailSent(ctx, notificationID)
}

// ─── Global Email Service ─────────────────────────────────────

var globalEmailService *EmailService

// InitGlobalEmailService sets the package-level email service instance.
// This allows other services (like ApprovalWorkflowService) to send
// emails without needing the EmailService as a direct dependency.
func InitGlobalEmailService(svc *EmailService) {
	globalEmailService = svc
	if svc != nil && svc.IsEnabled() {
		log.Println("📧 Global Email Service initialized")
	}
}

// SendEmailForUser is a convenience function that looks up a user's email
// and sends them a notification via the global email service (if configured).
// It runs asynchronously to avoid blocking the caller (e.g., database transactions).
func SendEmailForUser(ctx context.Context, userID, title, body string) {
	if globalEmailService == nil || !globalEmailService.IsEnabled() {
		return
	}

	go func() {
		emailCtx := context.Background()
		var email string
		err := repository.GetEmployeeEmailByUserID(emailCtx, userID, &email)
		if err != nil || email == "" {
			return
		}
		req := &models.CreateNotificationRequest{
			Title: title,
			Body:  body,
		}
		_ = globalEmailService.SendNotification(emailCtx, req, email)
	}()
}
