package schedule

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

// GetAll returns timings visible to the caller (nil branchIDs = super_admin = all).
func (s *Service) GetAll(ctx context.Context, branchIDs []string) ([]models.OfficeTiming, error) {
	return s.repo.FindAll(ctx, branchIDs)
}

// GetByID returns a single timing with days, enforcing branch membership.
func (s *Service) GetByID(ctx context.Context, id string, branchIDs []string) (*models.OfficeTiming, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !containsBranch(branchIDs, t.BranchID) {
		return nil, errors.New("forbidden")
	}
	return t, nil
}

// Create creates a new timing for the given branch.
func (s *Service) Create(ctx context.Context, branchID string, req CreateOfficeTimingRequest) (*models.OfficeTiming, error) {
	return s.repo.Create(ctx, branchID, req.Name, req.Days)
}

// Update replaces name + days, enforcing branch membership.
func (s *Service) Update(ctx context.Context, id string, branchIDs []string, req UpdateOfficeTimingRequest) (*models.OfficeTiming, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !containsBranch(branchIDs, existing.BranchID) {
		return nil, errors.New("forbidden")
	}
	return s.repo.Update(ctx, id, req.Name, req.Days)
}

// Delete removes a timing, enforcing branch membership.
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

// Activate sets the branch's active office timing, enforcing branch membership.
func (s *Service) Activate(ctx context.Context, timingID string, branchIDs []string) error {
	return s.repo.Activate(ctx, timingID, branchIDs)
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
