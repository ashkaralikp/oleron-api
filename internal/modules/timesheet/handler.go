package timesheet

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func parseYearMonth(r *http.Request) (year, month int, ok bool) {
	y, errY := strconv.Atoi(r.URL.Query().Get("year"))
	m, errM := strconv.Atoi(r.URL.Query().Get("month"))
	if errY != nil || errM != nil || y < 2000 || m < 1 || m > 12 {
		return 0, 0, false
	}
	return y, m, true
}

type Handler struct {
	repo *Repository
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{repo: NewRepository(db)}
}

func ctx(r *http.Request) context.Context { return r.Context() }

func (h *Handler) attachEstimate(c context.Context, ts *TimesheetResponse) {
	cfg, err := h.repo.FetchConfigByEmployeeID(c, ts.EmployeeID)
	if err != nil {
		return
	}
	result := Compute(EstimateInput{
		Year:               ts.Year,
		Month:              ts.Month,
		SupportHours:       ts.SupportHours,
		OvertimeHours:      ts.OvertimeHours,
		FixedMonthlySalary: cfg.FixedMonthlySalary,
		OTRate:             cfg.OTRate,
	})
	ts.EstimatedPay = result.EstimatedPay
	ts.Currency = cfg.Currency
}

// Estimate computes the pay estimate for the currently logged-in user.
func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request) {
	callerUserID := r.Context().Value(middleware.UserIDKey).(string)

	var req EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	cfg, err := h.repo.FetchConfigByUserID(ctx(r), callerUserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "employee record not found for your account")
		return
	}

	result := Compute(EstimateInput{
		Year:               req.Year,
		Month:              req.Month,
		SupportHours:       req.SupportHours,
		OvertimeHours:      req.OvertimeHours,
		FixedMonthlySalary: cfg.FixedMonthlySalary,
		OTRate:             cfg.OTRate,
	})

	response.Success(w, EstimateResponse{
		Year:               req.Year,
		Month:              req.Month,
		WholeMonthHours:    result.WholeMonthHours,
		SupportHours:       req.SupportHours,
		OvertimeHours:      req.OvertimeHours,
		FixedMonthlySalary: cfg.FixedMonthlySalary,
		OTRate:             cfg.OTRate,
		HourlyRate:         result.HourlyRate,
		Scenario:           result.Scenario,
		EstimatedPay:       result.EstimatedPay,
		Currency:           cfg.Currency,
	})
}

// Submit allows a consultant to submit (or resubmit) their monthly timesheet.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	callerUserID := r.Context().Value(middleware.UserIDKey).(string)

	cfg, err := h.repo.FetchConfigByUserID(ctx(r), callerUserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "employee record not found for your account")
		return
	}

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ts, err := h.repo.Submit(ctx(r), cfg.EmployeeID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to submit timesheet")
		return
	}
	response.Success(w, ts)
}

// GetMine returns the timesheet for the given year+month for the logged-in consultant.
func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	callerUserID := r.Context().Value(middleware.UserIDKey).(string)

	year, month, ok := parseYearMonth(r)
	if !ok {
		response.Error(w, http.StatusBadRequest, "year and month query parameters are required (e.g. ?year=2026&month=4)")
		return
	}

	cfg, err := h.repo.FetchConfigByUserID(ctx(r), callerUserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "employee record not found for your account")
		return
	}

	ts, err := h.repo.GetMine(ctx(r), cfg.EmployeeID, year, month)
	if err != nil {
		response.Error(w, http.StatusNotFound, "timesheet not found for the given month")
		return
	}
	response.Success(w, ts)
}

// GetAll returns timesheets for a given month (branch-scoped for admin/manager).
// year and month are optional query params; defaults to the current month.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(middleware.UserRoleKey).(string)
	branchID := r.Context().Value(middleware.UserBranchIDKey).(string)

	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if y, m, ok := parseYearMonth(r); ok {
		year, month = y, m
	}

	list, err := h.repo.GetAll(ctx(r), role, branchID, year, month)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch timesheets")
		return
	}
	for i := range list {
		h.attachEstimate(ctx(r), &list[i])
	}
	response.Success(w, list)
}

// GetByID returns a single timesheet; admin/manager are restricted to their branch.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := r.Context().Value(middleware.UserRoleKey).(string)
	branchID := r.Context().Value(middleware.UserBranchIDKey).(string)

	if role != "super_admin" {
		ok, err := h.repo.IsTimesheetInBranch(ctx(r), id, branchID)
		if err != nil || !ok {
			response.Error(w, http.StatusNotFound, "timesheet not found")
			return
		}
	}

	ts, err := h.repo.GetByID(ctx(r), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "timesheet not found")
		return
	}
	h.attachEstimate(ctx(r), ts)
	response.Success(w, ts)
}

// Review lets an admin or manager approve or reject a timesheet.
func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role := r.Context().Value(middleware.UserRoleKey).(string)
	branchID := r.Context().Value(middleware.UserBranchIDKey).(string)
	reviewerUserID := r.Context().Value(middleware.UserIDKey).(string)

	if role != "super_admin" {
		ok, err := h.repo.IsTimesheetInBranch(ctx(r), id, branchID)
		if err != nil || !ok {
			response.Error(w, http.StatusNotFound, "timesheet not found")
			return
		}
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ts, err := h.repo.Review(ctx(r), id, req.Status, req.ReviewNote, reviewerUserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update timesheet")
		return
	}
	response.Success(w, ts)
}
