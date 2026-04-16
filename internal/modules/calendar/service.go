package calendar

import (
	"context"
	"errors"

	"rmp-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetAll returns entries scoped by role + optional filters.
func (s *Service) GetAll(ctx context.Context, role, branchID string, f CalendarRangeFilter) ([]models.BranchCalendar, error) {
	if role != "super_admin" {
		f.BranchID = branchID
	}
	return s.repo.FindAll(ctx, f)
}

// GetByID returns a single entry, enforcing branch ownership for non-super_admin.
func (s *Service) GetByID(ctx context.Context, id, role, branchID string) (*models.BranchCalendar, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "super_admin" && e.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return e, nil
}

// Create adds a calendar entry for the caller's branch.
func (s *Service) Create(ctx context.Context, role, branchID string, req CreateCalendarEntryRequest) (*models.BranchCalendar, error) {
	return s.repo.Create(ctx, branchID, req.Date, req.Type, req.Name)
}

// Update modifies an existing entry, enforcing branch ownership for non-super_admin.
func (s *Service) Update(ctx context.Context, id, role, branchID string, req UpdateCalendarEntryRequest) (*models.BranchCalendar, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "super_admin" && existing.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return s.repo.Update(ctx, id, req.Type, req.Name)
}

// Delete removes an entry, enforcing branch ownership for non-super_admin.
func (s *Service) Delete(ctx context.Context, id, role, branchID string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role != "super_admin" && existing.BranchID != branchID {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}
