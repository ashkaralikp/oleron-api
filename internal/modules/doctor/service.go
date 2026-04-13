package doctor

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

func (s *Service) GetAll(ctx context.Context) ([]models.Doctor, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Doctor, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateDoctorRequest) (*models.Doctor, error) {
	d := &models.Doctor{
		UserID:         req.UserID,
		Specialization: req.Specialization,
		LicenseNo:      req.LicenseNo,
		IsAvailable:    true,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateDoctorRequest) error {
	d := &models.Doctor{
		Specialization: req.Specialization,
		LicenseNo:      req.LicenseNo,
	}
	if req.IsAvailable != nil {
		d.IsAvailable = *req.IsAvailable
	}

	return s.repo.Update(ctx, id, d)
}
