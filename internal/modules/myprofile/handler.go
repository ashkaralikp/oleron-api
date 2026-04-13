package myprofile

import (
	"encoding/json"
	"net/http"
	"strings"

	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Handler{service: svc}
}

func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusForbidden, "unable to determine user")
		return
	}

	var req UpdateMyProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateMyProfile(r.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "first_name cannot be empty", "last_name cannot be empty", "email cannot be empty":
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to update profile: "+err.Error())
		}
		return
	}

	response.Success(w, user)
}

func (h *Handler) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusForbidden, "unable to determine user")
		return
	}

	var req ChangeMyPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.ChangeMyPassword(r.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "old_password is required", "new_password is required", "old_password cannot be empty", "new_password cannot be empty", "old_password is incorrect":
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to change password: "+err.Error())
		}
		return
	}

	response.Success(w, map[string]string{"message": "password updated successfully"})
}

func (h *Handler) GetMyMenus(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || role == "" {
		response.Error(w, http.StatusForbidden, "unable to determine user role")
		return
	}

	menus, err := h.service.GetMenusByRole(r.Context(), role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch menus")
		return
	}

	response.Success(w, menus)
}
