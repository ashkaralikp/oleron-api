package router

import (
	"net/http"

	"rmp-api/internal/config"
	"rmp-api/internal/middleware"
	"rmp-api/internal/modules/admin"
	"rmp-api/internal/modules/attendance"
	"rmp-api/internal/modules/auth"
	"rmp-api/internal/modules/calendar"
	"rmp-api/internal/modules/myprofile"
	"rmp-api/internal/modules/payroll"
	"rmp-api/internal/modules/recruitment"
	"rmp-api/internal/modules/reports"
	"rmp-api/internal/modules/schedule"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(cfg *config.Config, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CORS)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (mobile app with API key)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKey(cfg))
			r.Post("/auth/login", auth.NewHandler(db, cfg.JWTSecret).Login)
			r.Post("/auth/refresh", auth.NewHandler(db, cfg.JWTSecret).Refresh)
			// Public recruitment: candidates apply without a JWT
			recruitmentHandlerPublic := recruitment.NewHandler(db)
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
			recruitmentHandler := recruitment.NewHandler(db)

			// Recruitment routes (super_admin, admin, manager)
			r.Route("/recruitment", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				// Vacancies
				r.Get("/vacancies", recruitmentHandler.GetAllVacancies)
				r.Post("/vacancies", recruitmentHandler.CreateVacancy)
				r.Get("/vacancies/{id}", recruitmentHandler.GetVacancyByID)
				r.Put("/vacancies/{id}", recruitmentHandler.UpdateVacancy)
				r.Patch("/vacancies/{id}/status", recruitmentHandler.UpdateVacancyStatus)
				r.Delete("/vacancies/{id}", recruitmentHandler.DeleteVacancy)
				// Applications
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

			// Payroll routes (super_admin, admin, manager)
			r.Route("/payroll", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				r.Get("/", payrollHandler.GetAll)
				r.Post("/generate", payrollHandler.Generate)
				r.Get("/{id}", payrollHandler.GetByID)
				r.Patch("/{id}/status", payrollHandler.UpdateStatus)
				r.Delete("/{id}", payrollHandler.Delete)
			})

			// Attendance routes (all authenticated roles)
			r.Route("/attendance", func(r chi.Router) {
				r.Post("/punch", attendanceHandler.Punch)
				r.Get("/today", attendanceHandler.GetToday)
			})

			// Calendar routes (super_admin, admin, manager)
			r.Route("/calendar/branch-calendar", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				r.Get("/", calendarHandler.GetAll)
				r.Post("/", calendarHandler.Create)
				r.Get("/{id}", calendarHandler.GetByID)
				r.Put("/{id}", calendarHandler.Update)
				r.Delete("/{id}", calendarHandler.Delete)
			})

			// Schedule routes (super_admin, admin, manager)
			r.Route("/schedule/office-timings", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				r.Get("/", scheduleHandler.GetAll)
				r.Post("/", scheduleHandler.Create)
				r.Get("/{id}", scheduleHandler.GetByID)
				r.Put("/{id}", scheduleHandler.Update)
				r.Delete("/{id}", scheduleHandler.Delete)
				r.Put("/{id}/activate", scheduleHandler.Activate)
			})

			// Reports routes (super_admin, admin, manager)
			r.Route("/reports", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				r.Get("/attendance", reportsHandler.GetAttendanceReport)
			})

			// Employee routes (super_admin, admin, manager)
			r.Route("/employees", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin", "admin", "manager"))
				r.Get("/", adminHandler.GetAllEmployees)
				r.Post("/", adminHandler.CreateEmployee)
				r.Get("/{id}", adminHandler.GetEmployeeByID)
				r.Put("/{id}", adminHandler.UpdateEmployee)
				r.Delete("/{id}", adminHandler.DeleteEmployee)
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
			})

			r.Put("/profile/me", profileHandler.UpdateMyProfile)
			r.Patch("/profile/me/password", profileHandler.ChangeMyPassword)

			// Menus for current user (any authenticated role)
			r.Get("/menus/me", profileHandler.GetMyMenus)
		})
	})

	return r
}
