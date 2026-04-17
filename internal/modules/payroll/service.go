package payroll

import (
	"context"
	"errors"
	"math"

	"rmp-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Generate computes and stores a payroll run for the branch covering the given period.
func (s *Service) Generate(ctx context.Context, branchID, generatedBy string, req GeneratePayrollRequest) (*models.PayrollRun, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Count expected working days in the period for this branch
	workingDays, err := s.repo.CountWorkingDays(ctx, branchID, req.PeriodFrom, req.PeriodTo)
	if err != nil || workingDays == 0 {
		workingDays = 0 // no timing configured — still proceed, working_days will be 0
	}

	// Fetch attendance + employee data for all active employees in the branch
	payData, err := s.repo.FetchEmployeePayData(ctx, branchID, req.PeriodFrom, req.PeriodTo)
	if err != nil {
		return nil, err
	}
	if len(payData) == 0 {
		return nil, errors.New("no active employees found in this branch")
	}

	// Build payroll items
	var items []models.PayrollItem
	var totalAmount float64

	for _, d := range payData {
		grossPay := round2(d.TotalHours * d.HourlyRate)
		deductions := 0.0 // extend here for tax/other deductions
		netPay := round2(grossPay - deductions)

		items = append(items, models.PayrollItem{
			EmployeeID:   d.EmployeeID,
			UserID:       d.UserID,
			FirstName:    d.FirstName,
			LastName:     d.LastName,
			Email:        d.Email,
			EmployeeCode: d.EmployeeCode,
			WorkingDays:  workingDays,
			PresentDays:  d.PresentDays,
			AbsentDays:   d.AbsentDays,
			LeaveDays:    d.LeaveDays,
			TotalHours:   round2(d.TotalHours),
			HourlyRate:   d.HourlyRate,
			Currency:     d.Currency,
			GrossPay:     grossPay,
			Deductions:   deductions,
			NetPay:       netPay,
		})
		totalAmount += netPay
	}

	return s.repo.CreatePayrollRun(
		ctx, branchID, generatedBy,
		req.PeriodFrom, req.PeriodTo,
		currency, req.Notes,
		round2(totalAmount), items,
	)
}

// GetAll returns payroll runs scoped by role.
func (s *Service) GetAll(ctx context.Context, role, branchID string) ([]models.PayrollRun, error) {
	if role == "super_admin" {
		return s.repo.FindAllRuns(ctx, "")
	}
	return s.repo.FindAllRuns(ctx, branchID)
}

// GetByID returns a run with items, enforcing branch ownership for non-super_admin.
func (s *Service) GetByID(ctx context.Context, id, role, branchID string) (*models.PayrollRun, error) {
	run, err := s.repo.FindRunByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "super_admin" && run.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return run, nil
}

// UpdateStatus moves a run to approved or paid, enforcing branch ownership.
func (s *Service) UpdateStatus(ctx context.Context, id, role, branchID string, req UpdateStatusRequest) (*models.PayrollRun, error) {
	run, err := s.repo.FindRunByID(ctx, id)
	if err != nil {
		return nil, errors.New("payroll run not found")
	}
	if role != "super_admin" && run.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	// Status transitions: draft → approved → paid
	if run.Status == "paid" {
		return nil, errors.New("payroll run is already marked as paid")
	}
	if req.Status == "approved" && run.Status != "draft" {
		return nil, errors.New("only draft runs can be approved")
	}
	if req.Status == "paid" && run.Status != "approved" {
		return nil, errors.New("only approved runs can be marked as paid")
	}
	return s.repo.UpdateStatus(ctx, id, req.Status)
}

// Delete removes a draft run. Approved/paid runs cannot be deleted.
func (s *Service) Delete(ctx context.Context, id, role, branchID string) error {
	run, err := s.repo.FindRunByID(ctx, id)
	if err != nil {
		return errors.New("payroll run not found")
	}
	if role != "super_admin" && run.BranchID != branchID {
		return errors.New("forbidden")
	}
	if run.Status != "draft" {
		return errors.New("only draft payroll runs can be deleted")
	}
	return s.repo.DeleteRun(ctx, id)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
