package router

import (
	"net/http"

	"rmp-api/internal/config"
	"rmp-api/internal/middleware"
	"rmp-api/internal/modules/admin"
	"rmp-api/internal/modules/appointment"
	"rmp-api/internal/modules/auth"
	"rmp-api/internal/modules/billing"
	"rmp-api/internal/modules/myprofile"
	"rmp-api/internal/modules/patient"

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
		})

		// Protected routes (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKey(cfg))
			r.Use(middleware.JWT(cfg))

			profileHandler := myprofile.NewHandler(db)

			// Patients
			r.Route("/patients", func(r chi.Router) {
				r.Get("/", patient.NewHandler(db).GetAll)
				r.Post("/", patient.NewHandler(db).Create)
				r.Get("/{id}", patient.NewHandler(db).GetByID)
				r.Put("/{id}", patient.NewHandler(db).Update)
				r.Delete("/{id}", patient.NewHandler(db).Delete)
			})

			// Billing
			r.Route("/billing", func(r chi.Router) {
				r.Get("/", billing.NewHandler(db).GetAll)
				r.Post("/", billing.NewHandler(db).Create)
				r.Get("/{id}", billing.NewHandler(db).GetByID)
				r.Put("/{id}", billing.NewHandler(db).Update)
			})

			// Appointments
			r.Route("/appointments", func(r chi.Router) {
				r.Get("/", appointment.NewHandler(db).GetAll)
				r.Post("/", appointment.NewHandler(db).Create)
				r.Put("/{id}", appointment.NewHandler(db).Update)
			})

			// Admin routes (super_admin only)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin"))

				adminHandler := admin.NewHandler(db)

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
