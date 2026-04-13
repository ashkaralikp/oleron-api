package billing

import (
	"context"

	"clinic-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context) ([]models.Billing, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Billing, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateBillingRequest) (*models.Billing, error) {
	b := &models.Billing{
		PatientID:   req.PatientID,
		TotalAmount: req.TotalAmount,
		PaidAmount:  req.PaidAmount,
		Status:      "pending",
	}
	if req.Notes != "" {
		b.Notes = &req.Notes
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateBillingRequest) error {
	b := &models.Billing{
		TotalAmount: req.TotalAmount,
		PaidAmount:  req.PaidAmount,
		Status:      req.Status,
	}
	if req.Notes != "" {
		b.Notes = &req.Notes
	}

	return s.repo.Update(ctx, id, b)
}
