package admin

import (
	"encoding/json"
	"net/http"

	"clinic-api/pkg/response"

	"github.com/go-chi/chi/v5"
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

// =============================================
// BRANCH HANDLERS
// =============================================

func (h *Handler) GetAllBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := h.service.GetAllBranches(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch branches")
		return
	}

	response.Success(w, branches)
}

func (h *Handler) GetBranchByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	branch, err := h.service.GetBranchByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "branch not found")
		return
	}

	response.Success(w, branch)
}

func (h *Handler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Code == "" {
		response.Error(w, http.StatusBadRequest, "name and code are required")
		return
	}

	branch, err := h.service.CreateBranch(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create branch: "+err.Error())
		return
	}

	response.Created(w, branch)
}

func (h *Handler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	branch, err := h.service.UpdateBranch(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update branch: "+err.Error())
		return
	}

	response.Success(w, branch)
}

func (h *Handler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteBranch(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete branch: "+err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "branch deleted"})
}

// =============================================
// USER HANDLERS
// =============================================

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	response.Success(w, users)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	response.Success(w, user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" || req.BranchID == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "first_name, last_name, email, password, branch_id, and role are required")
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	response.Created(w, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update user: "+err.Error())
		return
	}

	response.Success(w, user)
}

func (h *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Password == "" {
		response.Error(w, http.StatusBadRequest, "password is required")
		return
	}

	if err := h.service.ResetUserPassword(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to reset password: "+err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "password reset successful"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete user: "+err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "user deleted"})
}

// =============================================
// MENU HANDLERS
// =============================================

func (h *Handler) GetAllMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := h.service.GetAllMenus(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch menus")
		return
	}

	response.Success(w, menus)
}

func (h *Handler) GetMenusTree(w http.ResponseWriter, r *http.Request) {
	menus, err := h.service.GetAllMenusTree(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch menus")
		return
	}

	response.Success(w, menus)
}

func (h *Handler) GetMenuByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	menu, err := h.service.GetMenuByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "menu not found")
		return
	}

	response.Success(w, menu)
}

func (h *Handler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	var req CreateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Label == "" {
		response.Error(w, http.StatusBadRequest, "label is required")
		return
	}

	menu, err := h.service.CreateMenu(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create menu: "+err.Error())
		return
	}

	response.Created(w, menu)
}

func (h *Handler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	menu, err := h.service.UpdateMenu(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update menu: "+err.Error())
		return
	}

	response.Success(w, menu)
}

func (h *Handler) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteMenu(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete menu: "+err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "menu deleted"})
}

// =============================================
// ROLE PERMISSION HANDLERS
// =============================================

func (h *Handler) GetAllRolePermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.GetAllRolePermissions(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch role permissions")
		return
	}

	response.Success(w, permissions)
}

func (h *Handler) GetRolePermissionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	permission, err := h.service.GetRolePermissionByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "role permission not found")
		return
	}

	response.Success(w, permission)
}

func (h *Handler) CreateRolePermission(w http.ResponseWriter, r *http.Request) {
	var req CreateRolePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Role == "" || req.Resource == "" {
		response.Error(w, http.StatusBadRequest, "role and resource are required")
		return
	}

	permission, err := h.service.CreateRolePermission(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create role permission: "+err.Error())
		return
	}

	response.Created(w, permission)
}

func (h *Handler) UpdateRolePermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateRolePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	permission, err := h.service.UpdateRolePermission(r.Context(), id, req)
	if err != nil {
		if err.Error() == "role permission not found" {
			response.Error(w, http.StatusNotFound, "role permission not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update role permission: "+err.Error())
		return
	}

	response.Success(w, permission)
}

func (h *Handler) DeleteRolePermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteRolePermission(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete role permission: "+err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "role permission deleted"})
}
