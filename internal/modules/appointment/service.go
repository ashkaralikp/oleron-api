package appointment

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

func (s *Service) GetAll(ctx context.Context) ([]models.Appointment, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Appointment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateAppointmentRequest) (*models.Appointment, error) {
	a := &models.Appointment{
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "scheduled",
	}
	if req.Notes != "" {
		a.Notes = &req.Notes
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateAppointmentRequest) error {
	a := &models.Appointment{
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    req.Status,
	}
	if req.Notes != "" {
		a.Notes = &req.Notes
	}

	return s.repo.Update(ctx, id, a)
}
