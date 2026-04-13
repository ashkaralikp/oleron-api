package report

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Generate(ctx context.Context, req ReportRequest) (*ReportResponse, error) {
	patients, err := s.repo.GetPatientCount(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	billing, err := s.repo.GetBillingTotal(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	appointments, err := s.repo.GetAppointmentCount(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	return &ReportResponse{
		Type:           req.Type,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		TotalPatients:  patients,
		TotalBilling:   billing,
		TotalAppoints:  appointments,
	}, nil
}
