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

// GetAll returns timings scoped by role.
// super_admin sees all branches; admin/manager see their branch only.
func (s *Service) GetAll(ctx context.Context, role, branchID string) ([]models.OfficeTiming, error) {
	if role == "super_admin" {
		return s.repo.FindAll(ctx, "")
	}
	return s.repo.FindAll(ctx, branchID)
}

// GetByID returns a single timing with days.
// admin/manager may only access timings from their own branch.
func (s *Service) GetByID(ctx context.Context, id, role, branchID string) (*models.OfficeTiming, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "super_admin" && t.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return t, nil
}

// Create creates a new timing for the caller's branch (or a specified branch for super_admin).
func (s *Service) Create(ctx context.Context, role, branchID string, req CreateOfficeTimingRequest) (*models.OfficeTiming, error) {
	targetBranch := branchID
	return s.repo.Create(ctx, targetBranch, req.Name, req.Days)
}

// Update replaces name + days; enforces branch ownership for non-super_admin.
func (s *Service) Update(ctx context.Context, id, role, branchID string, req UpdateOfficeTimingRequest) (*models.OfficeTiming, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != "super_admin" && existing.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return s.repo.Update(ctx, id, req.Name, req.Days)
}

// Delete removes a timing; enforces branch ownership for non-super_admin.
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

// Activate sets the branch's active office timing.
func (s *Service) Activate(ctx context.Context, timingID, role, branchID string) error {
	scopedBranchID := branchID
	if role == "super_admin" {
		scopedBranchID = "" // skip ownership check — repo will still set correct branch
	}
	return s.repo.Activate(ctx, timingID, scopedBranchID)
}
