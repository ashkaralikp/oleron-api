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

// GetAll returns entries visible to the caller, with optional date/type filters.
// nil branchIDs = super_admin = all branches.
func (s *Service) GetAll(ctx context.Context, branchIDs []string, f CalendarRangeFilter) ([]models.BranchCalendar, error) {
	f.BranchIDs = branchIDs
	return s.repo.FindAll(ctx, f)
}

// GetByID returns a single entry, enforcing branch membership.
func (s *Service) GetByID(ctx context.Context, id string, branchIDs []string) (*models.BranchCalendar, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !containsBranch(branchIDs, e.BranchID) {
		return nil, errors.New("forbidden")
	}
	return e, nil
}

// Create adds a calendar entry for the given branch.
func (s *Service) Create(ctx context.Context, branchID string, req CreateCalendarEntryRequest) (*models.BranchCalendar, error) {
	return s.repo.Create(ctx, branchID, req.Date, req.Type, req.Name)
}

// Update modifies an existing entry, enforcing branch membership.
func (s *Service) Update(ctx context.Context, id string, branchIDs []string, req UpdateCalendarEntryRequest) (*models.BranchCalendar, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !containsBranch(branchIDs, existing.BranchID) {
		return nil, errors.New("forbidden")
	}
	return s.repo.Update(ctx, id, req.Type, req.Name)
}

// Delete removes an entry, enforcing branch membership.
func (s *Service) Delete(ctx context.Context, id string, branchIDs []string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !containsBranch(branchIDs, existing.BranchID) {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func containsBranch(branchIDs []string, id string) bool {
	if branchIDs == nil {
		return true
	}
	for _, b := range branchIDs {
		if b == id {
			return true
		}
	}
	return false
}
