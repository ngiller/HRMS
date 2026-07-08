package routes

import (
	"hrms-backend/internal/handlers"
	"hrms-backend/internal/middleware"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// Setup registers all API routes on the given Fiber app.
//
// Global middleware (recover, requestid, logger, helmet, CORS) is expected
// to be applied in main.go before calling this function.
func Setup(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	employeeHandler *handlers.EmployeeHandler,
	departmentHandler *handlers.DepartmentHandler,
	roleHandler *handlers.RoleHandler,
	positionGradeHandler *handlers.PositionGradeHandler,
	positionHandler *handlers.PositionHandler,
	workScheduleHandler *handlers.WorkScheduleHandler,
	attendanceLocationHandler *handlers.AttendanceLocationHandler,
	organizationHandler *handlers.OrganizationHandler,
	salaryComponentHandler *handlers.SalaryComponentHandler,
	shiftChangeHandler *handlers.ShiftChangeHandler,
	overtimeHandler *handlers.OvertimeHandler,
	reimbursementHandler *handlers.ReimbursementHandler,
	attendanceRecordHandler *handlers.AttendanceRecordHandler,
	scheduleHandler *handlers.ScheduleHandler,
	payrollHandler *handlers.PayrollHandler,
	leaveHandler *handlers.LeaveHandler,
	documentHandler *handlers.DocumentHandler,
	announcementHandler *handlers.AnnouncementHandler,
	holidayHandler *handlers.HolidayHandler,
	loanHandler *handlers.LoanHandler,
	kpiHandler *handlers.KPIHandler,
	reprimandHandler *handlers.ReprimandHandler,
	dailyJournalHandler *handlers.DailyJournalHandler,
	reportHandler *handlers.ReportHandler,
	notificationHandler *handlers.NotificationHandler,
	activityLogHandler *handlers.ActivityLogHandler,
	companyHandler *handlers.CompanyHandler,
	manualAttendanceHandler *handlers.ManualAttendanceHandler,
	resignHandler *handlers.ResignHandler,
	approvalWorkflowHandler *handlers.ApprovalWorkflowHandler,
	authService *service.AuthService,
) {
	// Health check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":  "healthy",
				"version": "1.0.0",
			},
		})
	})

	// Public routes (no auth required)
	api := app.Group("/api")

	// ==================== Auth Routes (public) ====================
	auth := api.Group("/auth")
	auth.Post("/login", middleware.RateLimitConfig(middleware.TierCritical), authHandler.Login)
	auth.Post("/forgot-password", middleware.ForgotPasswordRateLimit(), authHandler.ForgotPassword)
	auth.Post("/reset-password", middleware.RateLimitConfig(middleware.TierHigh), authHandler.ResetPassword)
	auth.Post("/refresh", middleware.RateLimitConfig(middleware.TierHigh), authHandler.RefreshToken)

	// ==================== Protected Routes ====================
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	protected.Use(middleware.RateLimitConfig(middleware.TierLow))

	// User info
	protected.Get("/auth/me", authHandler.Me)

	// Logout & Change Password
	protected.Post("/auth/logout", authHandler.Logout)
	protected.Put("/auth/change-password", authHandler.ChangePassword)

	// Dashboard
	protected.Get("/dashboard", employeeHandler.Dashboard)
	protected.Get("/dashboard/manager", middleware.RBAC("employee", "read"), employeeHandler.ManagerDashboard)
	protected.Get("/dashboard/hr", middleware.RBAC("employee", "read"), employeeHandler.HRDashboard)

	// ==================== Role Routes ====================
	roles := protected.Group("/roles")
	roles.Get("/permissions/template", roleHandler.GetPermissionTemplate)
	roles.Get("/", middleware.RBAC("user_management", "read"), roleHandler.ListRoles)
	roles.Get("/:id", middleware.RBAC("user_management", "read"), roleHandler.GetRole)
	roles.Post("/", middleware.RBAC("user_management", "create"), roleHandler.CreateRole)
	roles.Put("/:id", middleware.RBAC("user_management", "update"), roleHandler.UpdateRole)
	roles.Delete("/:id", middleware.RBAC("user_management", "delete"), roleHandler.DeleteRole)

	// ==================== Employee Routes ====================
	employees := protected.Group("/employees")
	employees.Get("/", middleware.RBAC("employee", "read"), employeeHandler.ListEmployees)
	employees.Get("/:id/history", middleware.RBAC("employee", "read"), employeeHandler.GetEmployeeHistory)
	employees.Get("/:id/salary-components", middleware.RBAC("payroll", "read"), salaryComponentHandler.ListComponents)
	employees.Post("/:id/salary-components", middleware.RBAC("payroll", "update"), salaryComponentHandler.CreateComponent)
	employees.Put("/:id/salary-components/:componentId", middleware.RBAC("payroll", "update"), salaryComponentHandler.UpdateComponent)
	employees.Delete("/:id/salary-components/:componentId", middleware.RBAC("payroll", "update"), salaryComponentHandler.DeleteComponent)
	employees.Get("/export", middleware.RBAC("employee", "read"), employeeHandler.ExportEmployees)
	employees.Post("/import", middleware.RBAC("employee", "create"), employeeHandler.ImportEmployees)
	employees.Get("/:id", middleware.RBAC("employee", "read"), employeeHandler.GetEmployee)
	employees.Post("/:id/photo", middleware.RBAC("employee", "update"), employeeHandler.UploadPhoto)
	employees.Put("/:id/work-schedule", middleware.RBAC("employee", "update"), employeeHandler.UpdateWorkSchedule)
	employees.Put("/:id/restore", middleware.RBAC("employee", "update"), employeeHandler.RestoreEmployee)
	employees.Post("/", middleware.RBAC("employee", "create"), employeeHandler.CreateEmployee)
	employees.Put("/:id", middleware.RBAC("employee", "update"), employeeHandler.UpdateEmployee)
	employees.Delete("/:id", middleware.RBAC("employee", "delete"), employeeHandler.DeleteEmployee)

	// ==================== Department Routes ====================
	departments := protected.Group("/departments")
	departments.Get("/work-schedules", middleware.RBAC("department", "read"), departmentHandler.GetWorkSchedules)
	departments.Get("/all", middleware.RBAC("department", "read"), departmentHandler.GetAllDepartments)
	departments.Get("/", middleware.RBAC("department", "read"), departmentHandler.ListDepartments)
	departments.Get("/:id", middleware.RBAC("department", "read"), departmentHandler.GetDepartment)
	departments.Post("/", middleware.RBAC("department", "create"), departmentHandler.CreateDepartment)
	departments.Put("/:id", middleware.RBAC("department", "update"), departmentHandler.UpdateDepartment)
	departments.Put("/:id/work-schedule", middleware.RBAC("department", "update"), departmentHandler.UpdateWorkSchedule)
	departments.Delete("/:id", middleware.RBAC("department", "delete"), departmentHandler.DeleteDepartment)

	// ==================== Position Grade Routes ====================
	positionGrades := protected.Group("/position-grades")
	positionGrades.Get("/all", middleware.RBAC("position_grade", "read"), positionGradeHandler.GetAllPositionGrades)
	positionGrades.Get("/", middleware.RBAC("position_grade", "read"), positionGradeHandler.ListPositionGrades)
	positionGrades.Get("/:id", middleware.RBAC("position_grade", "read"), positionGradeHandler.GetPositionGrade)
	positionGrades.Post("/", middleware.RBAC("position_grade", "create"), positionGradeHandler.CreatePositionGrade)
	positionGrades.Put("/:id", middleware.RBAC("position_grade", "update"), positionGradeHandler.UpdatePositionGrade)
	positionGrades.Delete("/:id", middleware.RBAC("position_grade", "delete"), positionGradeHandler.DeletePositionGrade)

	// ==================== Position Routes ====================
	positions := protected.Group("/positions")
	positions.Get("/all", middleware.RBAC("position", "read"), positionHandler.GetAllPositions)
	positions.Get("/", middleware.RBAC("position", "read"), positionHandler.ListPositions)
	positions.Get("/:id", middleware.RBAC("position", "read"), positionHandler.GetPosition)
	positions.Post("/", middleware.RBAC("position", "create"), positionHandler.CreatePosition)
	positions.Put("/:id", middleware.RBAC("position", "update"), positionHandler.UpdatePosition)
	positions.Delete("/:id", middleware.RBAC("position", "delete"), positionHandler.DeletePosition)

	// ==================== Work Schedule Routes ====================
	workSchedules := protected.Group("/work-schedules")
	workSchedules.Get("/all", middleware.RBAC("work_schedule", "read"), workScheduleHandler.GetAllWorkSchedules)
	workSchedules.Get("/", middleware.RBAC("work_schedule", "read"), workScheduleHandler.ListWorkSchedules)
	workSchedules.Get("/:id", middleware.RBAC("work_schedule", "read"), workScheduleHandler.GetWorkSchedule)
	workSchedules.Post("/", middleware.RBAC("work_schedule", "create"), workScheduleHandler.CreateWorkSchedule)
	workSchedules.Put("/:id", middleware.RBAC("work_schedule", "update"), workScheduleHandler.UpdateWorkSchedule)
	workSchedules.Delete("/:id", middleware.RBAC("work_schedule", "delete"), workScheduleHandler.DeleteWorkSchedule)

	// ==================== Attendance Location Routes ====================
	attendanceLocations := protected.Group("/attendance-locations")
	attendanceLocations.Get("/all", middleware.RBAC("attendance_location", "read"), attendanceLocationHandler.GetAllAttendanceLocations)
	attendanceLocations.Get("/", middleware.RBAC("attendance_location", "read"), attendanceLocationHandler.ListAttendanceLocations)
	attendanceLocations.Get("/:id", middleware.RBAC("attendance_location", "read"), attendanceLocationHandler.GetAttendanceLocation)
	attendanceLocations.Post("/", middleware.RBAC("attendance_location", "create"), attendanceLocationHandler.CreateAttendanceLocation)
	attendanceLocations.Put("/:id", middleware.RBAC("attendance_location", "update"), attendanceLocationHandler.UpdateAttendanceLocation)
	attendanceLocations.Delete("/:id", middleware.RBAC("attendance_location", "delete"), attendanceLocationHandler.DeleteAttendanceLocation)

	// ==================== Overtime Routes ====================
	overtimeRequests := protected.Group("/overtime-requests")
	overtimeRequests.Get("/", middleware.RBAC("overtime", "read"), overtimeHandler.ListOvertimeRequests)
	overtimeRequests.Get("/:id", middleware.RBAC("overtime", "read"), overtimeHandler.GetOvertimeRequest)
	overtimeRequests.Get("/:id/calculation", middleware.RBAC("overtime", "read"), overtimeHandler.GetOvertimeCalculation)
	overtimeRequests.Post("/", middleware.RBAC("overtime", "create"), overtimeHandler.CreateOvertimeRequest)
	overtimeRequests.Put("/:id/approve", middleware.RBAC("overtime", "approve"), overtimeHandler.ApproveOvertimeRequest)
	overtimeRequests.Put("/:id/reject", middleware.RBAC("overtime", "update"), overtimeHandler.RejectOvertimeRequest)
	overtimeRequests.Put("/:id/cancel", middleware.RBAC("overtime", "create"), overtimeHandler.CancelOvertimeRequest)

	// ==================== Reimbursement Routes ====================
	reimbursements := protected.Group("/reimbursements")
	reimbursements.Get("/", middleware.RBAC("reimbursement", "read"), reimbursementHandler.ListReimbursements)
	reimbursements.Get("/:id", middleware.RBAC("reimbursement", "read"), reimbursementHandler.GetReimbursement)
	reimbursements.Post("/upload", middleware.RBAC("reimbursement", "create"), reimbursementHandler.UploadReceipt)
	reimbursements.Post("/", middleware.RBAC("reimbursement", "create"), reimbursementHandler.CreateReimbursement)
	reimbursements.Put("/:id/approve", middleware.RBAC("reimbursement", "approve"), reimbursementHandler.ApproveReimbursement)
	reimbursements.Put("/:id/reject", middleware.RBAC("reimbursement", "update"), reimbursementHandler.RejectReimbursement)
	reimbursements.Put("/:id/pay", middleware.RBAC("reimbursement", "update"), reimbursementHandler.PayReimbursement)
	reimbursements.Put("/:id/cancel", middleware.RBAC("reimbursement", "create"), reimbursementHandler.CancelReimbursement)

	// ==================== Shift Change Request Routes ====================
	shiftChanges := protected.Group("/shift-change-requests")
	shiftChanges.Get("/", middleware.RBAC("shift_change", "read"), shiftChangeHandler.ListShiftChangeRequests)
	shiftChanges.Get("/:id", middleware.RBAC("shift_change", "read"), shiftChangeHandler.GetShiftChangeRequest)
	shiftChanges.Post("/", middleware.RBAC("shift_change", "create"), shiftChangeHandler.CreateShiftChangeRequest)
	shiftChanges.Put("/:id/approve", middleware.RBAC("shift_change", "update"), shiftChangeHandler.ApproveShiftChangeRequest)
	shiftChanges.Put("/:id/reject", middleware.RBAC("shift_change", "update"), shiftChangeHandler.RejectShiftChangeRequest)
	shiftChanges.Put("/:id/confirm-swap", middleware.RBAC("shift_change", "update"), shiftChangeHandler.ConfirmSwapShiftChangeRequest)
	shiftChanges.Put("/:id/cancel", middleware.RBAC("shift_change", "create"), shiftChangeHandler.CancelShiftChangeRequest)

	// ==================== Attendance Record Routes ====================
	attendance := protected.Group("/attendance")
	attendance.Get("/today", attendanceRecordHandler.GetTodayStatus)
	attendance.Post("/check-in", middleware.RBAC("attendance", "create"), attendanceRecordHandler.CheckIn)
	attendance.Put("/check-out", middleware.RBAC("attendance", "update"), attendanceRecordHandler.CheckOut)
	attendance.Get("/my-history", attendanceRecordHandler.ListMyAttendance)
	attendance.Get("/report", middleware.RBAC("attendance", "read"), attendanceRecordHandler.ListAttendanceReport)
	attendance.Get("/report/export", middleware.RBAC("attendance", "read"), attendanceRecordHandler.ExportAttendanceReport)

	// ==================== Manual Attendance Routes ====================
	manualAttendance := protected.Group("/manual-attendance")
	manualAttendance.Get("/", middleware.RBAC("attendance", "read"), manualAttendanceHandler.ListManualAttendance)
	manualAttendance.Get("/:id", middleware.RBAC("attendance", "read"), manualAttendanceHandler.GetManualAttendance)
	manualAttendance.Post("/", middleware.RBAC("attendance", "create"), manualAttendanceHandler.CreateManualAttendance)
	manualAttendance.Put("/:id/approve", middleware.RBAC("attendance", "approve"), manualAttendanceHandler.ApproveManualAttendance)
	manualAttendance.Put("/:id/reject", middleware.RBAC("attendance", "update"), manualAttendanceHandler.RejectManualAttendance)
	manualAttendance.Put("/:id/cancel", middleware.RBAC("attendance", "create"), manualAttendanceHandler.CancelManualAttendance)

	// ==================== Resign & Exit Management Routes ====================
	resign := protected.Group("/resign")
	resign.Get("/", middleware.RBAC("employee", "read"), resignHandler.ListResigns)
	resign.Get("/:id", middleware.RBAC("employee", "read"), resignHandler.GetResign)
	resign.Post("/", middleware.RBAC("employee", "create"), resignHandler.CreateResign)
	resign.Put("/:id/approve", middleware.RBAC("employee", "update"), resignHandler.ApproveResign)
	resign.Put("/:id/reject", middleware.RBAC("employee", "update"), resignHandler.RejectResign)
	resign.Get("/:id/clearance", middleware.RBAC("employee", "read"), resignHandler.ListClearanceItems)
	resign.Put("/clearance/:itemId", middleware.RBAC("employee", "update"), resignHandler.UpdateClearanceItem)

	// ==================== Organization Tree Routes ====================
	protected.Get("/organization/tree", organizationHandler.GetTree)

	// ==================== Payroll Routes ====================
	payroll := protected.Group("/payroll")
	payroll.Get("/periods", middleware.RBAC("payroll", "read"), payrollHandler.ListPeriods)
	payroll.Post("/periods", middleware.RBAC("payroll", "create"), payrollHandler.CreatePeriod)
	payroll.Get("/periods/:id", middleware.RBAC("payroll", "read"), payrollHandler.GetPeriod)
	payroll.Post("/periods/:id/calculate", middleware.RBAC("payroll", "update"), payrollHandler.CalculatePayroll)
	payroll.Get("/periods/:id/items", middleware.RBAC("payroll", "read"), payrollHandler.ListPayrollItems)
	payroll.Put("/periods/:id/approve", middleware.RBAC("payroll", "update"), payrollHandler.ApprovePeriod)
	payroll.Put("/periods/:id/pay", middleware.RBAC("payroll", "update"), payrollHandler.PayPeriod)
	payroll.Get("/payslips/:periodId/:employeeId", middleware.RBAC("payroll", "read"), payrollHandler.GetPayslip)
	payroll.Get("/my-payslips", middleware.RBAC("payslip", "read"), payrollHandler.ListMyPayslips)
	payroll.Get("/my-payslips/:periodId", middleware.RBAC("payslip", "read"), payrollHandler.GetMyPayslip)

	// ==================== Leave Routes ====================
	leaves := protected.Group("/leaves")
	leaves.Get("/types", leaveHandler.GetLeaveTypes)
	leaves.Get("/my-balances", leaveHandler.GetMyLeaveBalances)
	leaves.Get("/balances", middleware.RBAC("leave", "read"), leaveHandler.GetAllLeaveBalances)
	leaves.Get("/", middleware.RBAC("leave", "read"), leaveHandler.ListLeaveRequests)
	leaves.Get("/:id", middleware.RBAC("leave", "read"), leaveHandler.GetLeaveRequest)
	leaves.Post("/", middleware.RBAC("leave", "create"), leaveHandler.CreateLeaveRequest)
	leaves.Put("/:id/approve", middleware.RBAC("leave", "approve"), leaveHandler.ApproveLeaveRequest)
	leaves.Put("/:id/reject", middleware.RBAC("leave", "update"), leaveHandler.RejectLeaveRequest)
	leaves.Put("/:id/cancel", middleware.RBAC("leave", "create"), leaveHandler.CancelLeaveRequest)

	// ==================== Schedule Template Routes ====================
	scheduleTemplates := protected.Group("/schedule-templates")
	scheduleTemplates.Get("/all", middleware.RBAC("work_schedule", "read"), scheduleHandler.GetAllTemplates)
	scheduleTemplates.Get("/", middleware.RBAC("work_schedule", "read"), scheduleHandler.ListTemplates)
	scheduleTemplates.Get("/:id", middleware.RBAC("work_schedule", "read"), scheduleHandler.GetTemplate)
	scheduleTemplates.Post("/", middleware.RBAC("work_schedule", "create"), scheduleHandler.CreateTemplate)
	scheduleTemplates.Put("/:id", middleware.RBAC("work_schedule", "update"), scheduleHandler.UpdateTemplate)
	scheduleTemplates.Delete("/:id", middleware.RBAC("work_schedule", "delete"), scheduleHandler.DeleteTemplate)

	// ==================== Employee Schedule Routes ====================
	employeeSchedules := protected.Group("/employee-schedules")
	employeeSchedules.Get("/resolve", middleware.RBAC("work_schedule", "read"), scheduleHandler.ResolveSchedule)
	employeeSchedules.Get("/", middleware.RBAC("work_schedule", "read"), scheduleHandler.ListEmployeeSchedules)
	employeeSchedules.Get("/:id", middleware.RBAC("work_schedule", "read"), scheduleHandler.GetEmployeeSchedule)
	employeeSchedules.Post("/", middleware.RBAC("work_schedule", "create"), scheduleHandler.CreateEmployeeSchedule)
	employeeSchedules.Put("/:id", middleware.RBAC("work_schedule", "update"), scheduleHandler.UpdateEmployeeSchedule)
	employeeSchedules.Delete("/:id", middleware.RBAC("work_schedule", "delete"), scheduleHandler.DeleteEmployeeSchedule)

	// ==================== Document Routes ====================
	documents := protected.Group("/documents")
	documents.Get("/", middleware.RBAC("document", "read"), documentHandler.ListDocuments)
	documents.Get("/:id", middleware.RBAC("document", "read"), documentHandler.GetDocument)
	documents.Post("/", middleware.RBAC("document", "create"), documentHandler.CreateDocument)
	documents.Put("/:id/verify", middleware.RBAC("document", "update"), documentHandler.VerifyDocument)
	documents.Put("/:id/reject", middleware.RBAC("document", "update"), documentHandler.RejectDocument)
	documents.Delete("/:id", middleware.RBAC("document", "delete"), documentHandler.DeleteDocument)

	// ==================== Announcement Routes ====================
	announcements := protected.Group("/announcements")
	announcements.Get("/", middleware.RBAC("announcement", "read"), announcementHandler.ListAnnouncements)
	announcements.Get("/:id", middleware.RBAC("announcement", "read"), announcementHandler.GetAnnouncement)
	announcements.Post("/", middleware.RBAC("announcement", "create"), announcementHandler.CreateAnnouncement)
	announcements.Put("/:id", middleware.RBAC("announcement", "update"), announcementHandler.UpdateAnnouncement)
	announcements.Delete("/:id", middleware.RBAC("announcement", "delete"), announcementHandler.DeleteAnnouncement)
	announcements.Post("/:id/read", announcementHandler.MarkAnnouncementRead)

	// ==================== Holiday Routes ====================
	holidays := protected.Group("/holidays")
	holidays.Get("/year/:year", holidayHandler.GetHolidaysByYear)
	holidays.Get("/", middleware.RBAC("announcement", "read"), holidayHandler.ListHolidays)
	holidays.Get("/:id", middleware.RBAC("announcement", "read"), holidayHandler.GetHoliday)
	holidays.Post("/", middleware.RBAC("announcement", "create"), holidayHandler.CreateHoliday)
	holidays.Put("/:id", middleware.RBAC("announcement", "update"), holidayHandler.UpdateHoliday)
	holidays.Delete("/:id", middleware.RBAC("announcement", "delete"), holidayHandler.DeleteHoliday)

	// ==================== Company Settings Routes ====================
	company := protected.Group("/company")
	company.Get("/settings", middleware.RBAC("company_settings", "read"), companyHandler.GetSettings)
	company.Put("/settings", middleware.RBAC("company_settings", "update"), companyHandler.UpdateSettings)

	// ==================== Per-Employee BPJS Config ====================
	protected.Get("/employees/:id/bpjs-config", middleware.RBAC("employee", "read"), companyHandler.GetEmployeeBPJSConfig)
	protected.Put("/employees/:id/bpjs-config", middleware.RBAC("employee", "update"), companyHandler.UpdateEmployeeBPJSConfig)

	// ==================== Payroll THR Calculation ====================
	payroll.Get("/periods/:id/calculate-thr", middleware.RBAC("payroll", "read"), payrollHandler.CalculateTHR)

	// ==================== Loan Routes (Pinjaman) ====================
	loans := protected.Group("/loans")
	loans.Get("/stats", middleware.RBAC("loan", "read"), loanHandler.GetLoanStats)
	loans.Get("/", middleware.RBAC("loan", "read"), loanHandler.ListLoans)
	loans.Get("/:id", middleware.RBAC("loan", "read"), loanHandler.GetLoan)
	loans.Post("/", middleware.RBAC("loan", "create"), loanHandler.CreateLoan)
	loans.Put("/:id/approve", middleware.RBAC("loan", "approve"), loanHandler.ApproveLoan)
	loans.Put("/:id/reject", middleware.RBAC("loan", "update"), loanHandler.RejectLoan)
	loans.Put("/:id/cancel", middleware.RBAC("loan", "create"), loanHandler.CancelLoan)
	loans.Put("/:id/disburse", middleware.RBAC("loan", "update"), loanHandler.DisburseLoan)

	// ==================== KPI Routes ====================
	kpi := protected.Group("/kpi")
	kpi.Get("/templates", middleware.RBAC("kpi", "read"), kpiHandler.ListKPITemplates)
	kpi.Get("/templates/:id", middleware.RBAC("kpi", "read"), kpiHandler.GetKPITemplate)
	kpi.Post("/templates", middleware.RBAC("kpi", "create"), kpiHandler.CreateKPITemplate)
	kpi.Put("/templates/:id", middleware.RBAC("kpi", "update"), kpiHandler.UpdateKPITemplate)
	kpi.Delete("/templates/:id", middleware.RBAC("kpi", "delete"), kpiHandler.DeleteKPITemplate)
	kpi.Get("/reviews", middleware.RBAC("kpi", "read"), kpiHandler.ListKPIReviews)
	kpi.Get("/reviews/:id", middleware.RBAC("kpi", "read"), kpiHandler.GetKPIReview)
	kpi.Post("/reviews", middleware.RBAC("kpi", "create"), kpiHandler.CreateKPIReview)

	// ==================== Reprimand Routes ====================
	reprimands := protected.Group("/reprimands")
	reprimands.Get("/", middleware.RBAC("reprimand", "read"), reprimandHandler.ListReprimands)
	reprimands.Get("/:id", middleware.RBAC("reprimand", "read"), reprimandHandler.GetReprimand)
	reprimands.Post("/", middleware.RBAC("reprimand", "create"), reprimandHandler.CreateReprimand)
	reprimands.Put("/:id/acknowledge", middleware.RBAC("reprimand", "update"), reprimandHandler.AcknowledgeReprimand)

	// ==================== Daily Journal Routes ====================
	dailyJournals := protected.Group("/daily-journals")
	dailyJournals.Get("/", middleware.RBAC("daily_journal", "read"), dailyJournalHandler.ListJournals)
	dailyJournals.Get("/:id", middleware.RBAC("daily_journal", "read"), dailyJournalHandler.GetJournal)
	dailyJournals.Post("/", middleware.RBAC("daily_journal", "create"), dailyJournalHandler.CreateJournal)
	dailyJournals.Put("/:id/acknowledge", middleware.RBAC("daily_journal", "update"), dailyJournalHandler.AcknowledgeJournal)

	// ==================== Report Routes ====================
	reports := protected.Group("/reports")
	reports.Get("/headcount", middleware.RBAC("report", "read"), reportHandler.Headcount)
	reports.Get("/payroll-summary", middleware.RBAC("report", "read"), reportHandler.PayrollSummary)
	reports.Get("/attendance-summary", middleware.RBAC("report", "read"), reportHandler.AttendanceSummary)
	reports.Get("/loan-summary", middleware.RBAC("report", "read"), reportHandler.LoanSummary)
	reports.Get("/leave-summary", middleware.RBAC("report", "read"), reportHandler.LeaveSummary)
	reports.Get("/overtime-summary", middleware.RBAC("report", "read"), reportHandler.OvertimeSummary)

	// ==================== Notification Routes ====================
	notifications := protected.Group("/notifications")
	notifications.Get("/", notificationHandler.ListNotifications)
	notifications.Get("/unread-count", notificationHandler.GetUnreadCount)
	notifications.Put("/mark-read", notificationHandler.MarkAsRead)
	notifications.Post("/", middleware.RBAC("employee", "create"), notificationHandler.CreateNotification)

	// ==================== Activity Log Routes ====================
	activityLogs := protected.Group("/activity-logs")
	activityLogs.Get("/", middleware.RBAC("employee", "read"), activityLogHandler.ListActivityLogs)
	activityLogs.Get("/entity-types", middleware.RBAC("employee", "read"), activityLogHandler.GetEntityTypes)
	activityLogs.Get("/actions", middleware.RBAC("employee", "read"), activityLogHandler.GetActions)
	activityLogs.Get("/:id", middleware.RBAC("employee", "read"), activityLogHandler.GetActivityLog)

	// ==================== Approval Workflow Routes ====================
	// Workflow Configuration (Admin)
	approvalWorkflows := protected.Group("/approval-workflows")
	approvalWorkflows.Get("/", middleware.RBAC("company_settings", "read"), approvalWorkflowHandler.ListWorkflows)
	approvalWorkflows.Get("/:id", middleware.RBAC("company_settings", "read"), approvalWorkflowHandler.GetWorkflowDetail)
	approvalWorkflows.Post("/", middleware.RBAC("company_settings", "update"), approvalWorkflowHandler.CreateWorkflow)
	approvalWorkflows.Delete("/:id", middleware.RBAC("company_settings", "update"), approvalWorkflowHandler.DeleteWorkflow)

	// Workflow Steps (Admin)
	approvalWorkflows.Post("/:id/steps", middleware.RBAC("company_settings", "update"), approvalWorkflowHandler.AddWorkflowStep)
	approvalWorkflowSteps := protected.Group("/approval-workflow-steps")
	approvalWorkflowSteps.Put("/:id", middleware.RBAC("company_settings", "update"), approvalWorkflowHandler.UpdateWorkflowStep)
	approvalWorkflowSteps.Delete("/:id", middleware.RBAC("company_settings", "update"), approvalWorkflowHandler.DeleteWorkflowStep)

	// Pending Approvals (all authenticated users)
	approvals := protected.Group("/approvals")
	approvals.Get("/pending", approvalWorkflowHandler.GetPendingApprovals)
	approvals.Post("/:entityType/:entityId/init", approvalWorkflowHandler.InitializeTracking)
	approvals.Put("/:entityType/:entityId/process", approvalWorkflowHandler.ProcessApproval)
}
