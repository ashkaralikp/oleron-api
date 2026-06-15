package router

import (
	"net/http"
	"time"

	"rmp-api/internal/config"
	"rmp-api/internal/middleware"
	"rmp-api/internal/modules/admin"
	"rmp-api/internal/modules/attendance"
	"rmp-api/internal/modules/auth"
	"rmp-api/internal/modules/calendar"
	"rmp-api/internal/modules/contact"
	"rmp-api/internal/modules/myprofile"
	"rmp-api/internal/modules/payroll"
	"rmp-api/internal/modules/recruitment"
	"rmp-api/internal/modules/reports"
	"rmp-api/internal/modules/schedule"
	"rmp-api/internal/modules/timesheet"
	"rmp-api/internal/modules/workorder"
	"rmp-api/pkg/email"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(cfg *config.Config, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CORS)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static file server for uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		var mailer *email.Sender
		if cfg.SMTPUser != "" && cfg.SMTPPass != "" {
			mailer = email.NewSender(email.Config{
				Host:     cfg.SMTPHost,
				Port:     cfg.SMTPPort,
				Username: cfg.SMTPUser,
				Password: cfg.SMTPPass,
				From:     cfg.SMTPFrom,
			})
		}
		contactHandler := contact.NewHandler(db, mailer, cfg.SMTPTo)
		r.With(middleware.RateLimit(5, time.Minute)).Post("/contact-submissions", contactHandler.CreateSubmission)

		// Public routes (mobile app with API key)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKey(cfg))
			r.Post("/auth/login", auth.NewHandler(db, cfg.JWTSecret).Login)
			r.Post("/auth/refresh", auth.NewHandler(db, cfg.JWTSecret).Refresh)
			// Public recruitment: no JWT required
			recruitmentHandlerPublic := recruitment.NewHandler(db, cfg.UploadDir, cfg.BaseURL)
			r.Post("/recruitment/upload/cv", recruitmentHandlerPublic.UploadCV)
			r.Get("/recruitment/vacancies/public", recruitmentHandlerPublic.GetPublicVacancies)
			r.Post("/recruitment/vacancies/{id}/apply", recruitmentHandlerPublic.Apply)
		})

		// Protected routes (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKey(cfg))
			r.Use(middleware.JWT(cfg))

			profileHandler := myprofile.NewHandler(db)
			adminHandler := admin.NewHandler(db)
			reportsHandler := reports.NewHandler(db)
			scheduleHandler := schedule.NewHandler(db)
			calendarHandler := calendar.NewHandler(db)
			attendanceHandler := attendance.NewHandler(db)
			payrollHandler := payroll.NewHandler(db)
			recruitmentHandler := recruitment.NewHandler(db, cfg.UploadDir, cfg.BaseURL)
			timesheetHandler := timesheet.NewHandler(db)
			workOrderHandler := workorder.NewHandler(db, cfg.UploadDir, cfg.BaseURL)

			// Timesheet estimate (all authenticated roles)
			r.Post("/timesheets/estimate", timesheetHandler.Estimate)

			// Timesheet submission (consultant only)
			r.With(middleware.RequireRole("consultant")).Post("/timesheets", timesheetHandler.Submit)
			r.With(middleware.RequireRole("consultant")).Get("/timesheets/me", timesheetHandler.GetMine)

			// Timesheet review (super_admin, admin, manager)
			r.With(middleware.RequireRole("super_admin", "regional_manager", "manager")).Get("/timesheets", timesheetHandler.GetAll)
			r.With(middleware.RequireRole("super_admin", "regional_manager", "manager")).Get("/timesheets/{id}", timesheetHandler.GetByID)
			r.With(middleware.RequireRole("super_admin", "regional_manager", "manager")).Patch("/timesheets/{id}/review", timesheetHandler.Review)

			// Recruitment routes (super_admin, regional_manager, manager)
			r.Route("/recruitment", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				// Vacancies
				r.Get("/vacancies", recruitmentHandler.GetAllVacancies)
				r.Post("/vacancies", recruitmentHandler.CreateVacancy)
				r.Get("/vacancies/{id}", recruitmentHandler.GetVacancyByID)
				r.Put("/vacancies/{id}", recruitmentHandler.UpdateVacancy)
				r.Patch("/vacancies/{id}/status", recruitmentHandler.UpdateVacancyStatus)
				r.Delete("/vacancies/{id}", recruitmentHandler.DeleteVacancy)
				// Applications
				r.Post("/vacancies/{id}/apply/bulk", recruitmentHandler.BulkApply)
				r.Get("/vacancies/{id}/applications", recruitmentHandler.GetApplicationsByVacancy)
				r.Get("/applications/{id}", recruitmentHandler.GetApplicationByID)
				r.Patch("/applications/{id}/status", recruitmentHandler.UpdateApplicationStatus)
				r.Delete("/applications/{id}", recruitmentHandler.DeleteApplication)
				r.Post("/applications/{id}/interviews", recruitmentHandler.CreateInterview)
				r.Post("/applications/{id}/hire", recruitmentHandler.Hire)
				// Interviews
				r.Put("/interviews/{id}", recruitmentHandler.UpdateInterview)
				r.Delete("/interviews/{id}", recruitmentHandler.DeleteInterview)
			})

			// Payroll routes (super_admin, regional_manager, manager)
			r.Route("/payroll", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/", payrollHandler.GetAll)
				r.Post("/generate", payrollHandler.Generate)
				r.Get("/{id}", payrollHandler.GetByID)
				r.Patch("/{id}/status", payrollHandler.UpdateStatus)
				r.Delete("/{id}", payrollHandler.Delete)
			})

			// Work order routes
			r.Route("/work-orders", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.With(middleware.RequireRole("regional_manager")).Post("/assets", workOrderHandler.UpsertAsset)
				r.With(middleware.RequireRole("regional_manager")).Get("/assets/me", workOrderHandler.GetMyAsset)
				r.Get("/invoices/{invoiceID}", workOrderHandler.GetInvoice)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Patch("/invoices/{invoiceID}/status", workOrderHandler.UpdateInvoiceStatus)
				r.Get("/invoices/{invoiceID}/payments", workOrderHandler.ListInvoicePayments)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Post("/invoices/{invoiceID}/payments", workOrderHandler.AddInvoicePayment)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Patch("/invoice-payments/{paymentID}/status", workOrderHandler.UpdateInvoicePaymentStatus)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Post("/invoice-payments/{paymentID}/statements", workOrderHandler.UploadInvoicePaymentStatement)
				r.Get("/", workOrderHandler.GetAll)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Post("/", workOrderHandler.Create)
				r.Get("/{id}/invoices", workOrderHandler.ListInvoices)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Post("/{id}/invoices", workOrderHandler.CreateInvoice)
				r.Get("/{id}", workOrderHandler.GetByID)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Put("/{id}", workOrderHandler.Update)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Patch("/{id}/status", workOrderHandler.UpdateStatus)
				r.With(middleware.RequireRole("super_admin", "regional_manager")).Delete("/{id}", workOrderHandler.Delete)
			})

			// Attendance routes (all authenticated roles)
			r.Route("/attendance", func(r chi.Router) {
				r.Post("/punch", attendanceHandler.Punch)
				r.Get("/today", attendanceHandler.GetToday)
			})

			// Calendar routes (super_admin, regional_manager, manager)
			r.Route("/calendar/branch-calendar", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/", calendarHandler.GetAll)
				r.Post("/", calendarHandler.Create)
				r.Get("/{id}", calendarHandler.GetByID)
				r.Put("/{id}", calendarHandler.Update)
				r.Delete("/{id}", calendarHandler.Delete)
			})

			// Schedule routes (super_admin, regional_manager, manager)
			r.Route("/schedule/office-timings", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/", scheduleHandler.GetAll)
				r.Post("/", scheduleHandler.Create)
				r.Get("/{id}", scheduleHandler.GetByID)
				r.Put("/{id}", scheduleHandler.Update)
				r.Delete("/{id}", scheduleHandler.Delete)
				r.Put("/{id}/activate", scheduleHandler.Activate)
			})

			// Reports routes (super_admin, regional_manager, manager)
			r.Route("/reports", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/attendance", reportsHandler.GetAttendanceReport)
			})

			// Employee routes (super_admin, regional_manager, manager)
			r.Route("/employees", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/", adminHandler.GetAllEmployees)
				r.Post("/", adminHandler.CreateEmployee)
				r.Get("/{id}", adminHandler.GetEmployeeByID)
				r.Put("/{id}", adminHandler.UpdateEmployee)
				r.Patch("/{id}/password", adminHandler.ResetEmployeePassword)
				r.Delete("/{id}", adminHandler.DeleteEmployee)
			})

			// Contact submissions (super_admin, regional_manager, manager)
			r.Route("/admin/contact-submissions", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "regional_manager", "manager"))
				r.Get("/", contactHandler.GetAll)
				r.Get("/{id}", contactHandler.GetByID)
				r.Patch("/{id}/status", contactHandler.UpdateStatus)
				r.Delete("/{id}", contactHandler.Delete)
			})

			// Admin routes (super_admin only)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin"))

				// Branches

				r.Route("/branches", func(r chi.Router) {
					r.Get("/", adminHandler.GetAllBranches)
					r.Post("/", adminHandler.CreateBranch)
					r.Get("/{id}", adminHandler.GetBranchByID)
					r.Put("/{id}", adminHandler.UpdateBranch)
					r.Delete("/{id}", adminHandler.DeleteBranch)
				})

				// Users
				r.Route("/users", func(r chi.Router) {
					r.Get("/", adminHandler.GetAllUsers)
					r.Post("/", adminHandler.CreateUser)
					r.Get("/{id}", adminHandler.GetUserByID)
					r.Put("/{id}", adminHandler.UpdateUser)
					r.Patch("/{id}/password", adminHandler.ResetUserPassword)
					r.Delete("/{id}", adminHandler.DeleteUser)
				})

				// Menus (CRUD for super_admin)
				r.Route("/menus", func(r chi.Router) {
					r.Get("/", adminHandler.GetAllMenus)
					r.Get("/tree", adminHandler.GetMenusTree)
					r.Post("/", adminHandler.CreateMenu)
					r.Get("/{id}", adminHandler.GetMenuByID)
					r.Put("/{id}", adminHandler.UpdateMenu)
					r.Delete("/{id}", adminHandler.DeleteMenu)
				})

				// Role permissions
				r.Route("/role-permissions", func(r chi.Router) {
					r.Get("/", adminHandler.GetAllRolePermissions)
					r.Post("/", adminHandler.CreateRolePermission)
					r.Get("/{id}", adminHandler.GetRolePermissionByID)
					r.Put("/{id}", adminHandler.UpdateRolePermission)
					r.Delete("/{id}", adminHandler.DeleteRolePermission)
				})

				// Regional manager branch assignments
				r.Get("/regional-managers/{id}/branches", adminHandler.GetRegionalManagerBranches)
				r.Put("/regional-managers/{id}/branches", adminHandler.SetRegionalManagerBranches)
			})

			r.Put("/profile/me", profileHandler.UpdateMyProfile)
			r.Patch("/profile/me/password", profileHandler.ChangeMyPassword)

			// Menus for current user (any authenticated role)
			r.Get("/menus/me", profileHandler.GetMyMenus)
		})
	})

	return r
}
