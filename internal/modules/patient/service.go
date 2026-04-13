package patient

import (
	"context"

	"rmp-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context) ([]models.Patient, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Patient, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreatePatientRequest) (*models.Patient, error) {
	p := &models.Patient{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}
	if req.Email != "" {
		p.Email = &req.Email
	}
	if req.DateOfBirth != "" {
		p.DateOfBirth = &req.DateOfBirth
	}
	if req.Gender != "" {
		p.Gender = &req.Gender
	}
	if req.Address != "" {
		p.Address = &req.Address
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdatePatientRequest) error {
	p := &models.Patient{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}
	if req.Email != "" {
		p.Email = &req.Email
	}
	if req.DateOfBirth != "" {
		p.DateOfBirth = &req.DateOfBirth
	}
	if req.Gender != "" {
		p.Gender = &req.Gender
	}
	if req.Address != "" {
		p.Address = &req.Address
	}

	return s.repo.Update(ctx, id, p)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
