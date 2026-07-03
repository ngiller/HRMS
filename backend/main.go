package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hrms-backend/internal/config"
	"hrms-backend/internal/database"
	"hrms-backend/internal/handlers"
	"hrms-backend/internal/middleware"
	"hrms-backend/internal/routes"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL(), cfg.EncryptionKey); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize services
	authService := service.NewAuthService(cfg)
	employeeService := service.NewEmployeeService()
	departmentService := service.NewDepartmentService()
	roleService := service.NewRoleService()
	positionGradeService := service.NewPositionGradeService()
	positionService := service.NewPositionService()
	workScheduleService := service.NewWorkScheduleService()
	attendanceLocationService := service.NewAttendanceLocationService()
	organizationService := service.NewOrganizationService()
	salaryComponentService := service.NewSalaryComponentService()
	shiftChangeService := service.NewShiftChangeService()
	overtimeService := service.NewOvertimeService()
	reimbursementService := service.NewReimbursementService()
	payrollService := service.NewPayrollService()
	attendanceRecordService := service.NewAttendanceRecordService()
	scheduleService := service.NewScheduleService()
	leaveService := service.NewLeaveService()
	documentService := service.NewDocumentService()
	announcementService := service.NewAnnouncementService()
	holidayService := service.NewHolidayService()
	loanService := service.NewLoanService()
	kpiService := service.NewKPIService()
	reprimandService := service.NewReprimandService()
	dailyJournalService := service.NewDailyJournalService()
	reportService := service.NewReportService()
	notificationService := service.NewNotificationService()
	activityLogService := service.NewActivityLogService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	roleHandler := handlers.NewRoleHandler(roleService)
	positionGradeHandler := handlers.NewPositionGradeHandler(positionGradeService)
	positionHandler := handlers.NewPositionHandler(positionService)
	workScheduleHandler := handlers.NewWorkScheduleHandler(workScheduleService)
	attendanceLocationHandler := handlers.NewAttendanceLocationHandler(attendanceLocationService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	salaryComponentHandler := handlers.NewSalaryComponentHandler(salaryComponentService)
	shiftChangeHandler := handlers.NewShiftChangeHandler(shiftChangeService)
	overtimeHandler := handlers.NewOvertimeHandler(overtimeService)
	reimbursementHandler := handlers.NewReimbursementHandler(reimbursementService)
	payrollHandler := handlers.NewPayrollHandler(payrollService)
	attendanceRecordHandler := handlers.NewAttendanceRecordHandler(attendanceRecordService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	leaveHandler := handlers.NewLeaveHandler(leaveService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	announcementHandler := handlers.NewAnnouncementHandler(announcementService)
	holidayHandler := handlers.NewHolidayHandler(holidayService)
	loanHandler := handlers.NewLoanHandler(loanService)
	kpiHandler := handlers.NewKPIHandler(kpiService)
	reprimandHandler := handlers.NewReprimandHandler(reprimandService)
	dailyJournalHandler := handlers.NewDailyJournalHandler(dailyJournalService)
	reportHandler := handlers.NewReportHandler(reportService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	activityLogHandler := handlers.NewActivityLogHandler(activityLogService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:       "HRMS API",
		CaseSensitive: true,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\\n",
	}))
	app.Use(helmet.New())
	app.Use(middleware.CORSConfig(cfg))

	// Serve uploaded files
	app.Static("/uploads", cfg.UploadDir)

	// Register all routes
	routes.Setup(app, authHandler, employeeHandler, departmentHandler, roleHandler,
		positionGradeHandler, positionHandler, workScheduleHandler, attendanceLocationHandler,
		organizationHandler, salaryComponentHandler, shiftChangeHandler, overtimeHandler, reimbursementHandler,
		attendanceRecordHandler, scheduleHandler, payrollHandler, leaveHandler,
		documentHandler, announcementHandler, holidayHandler,
		loanHandler, kpiHandler, reprimandHandler, dailyJournalHandler, reportHandler,
		notificationHandler, activityLogHandler,
		authService)

	// Start server
	go func() {
		addr := cfg.ServerHost + ":" + cfg.ServerPort
		log.Printf("Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
