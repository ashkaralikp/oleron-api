package timesheet

import (
	"context"
	"encoding/json"
	"net/http"

	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{repo: NewRepository(db)}
}

func ctx(r *http.Request) context.Context { return r.Context() }

// Estimate computes the pay estimate for the currently logged-in employee.
func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request) {
	callerUserID := r.Context().Value(middleware.UserIDKey).(string)
	_ = r.Context().Value(middleware.UserRoleKey).(string)

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
