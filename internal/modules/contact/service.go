package contact

import (
	"context"
	"strings"

	"rmp-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSubmission(ctx context.Context, req CreateSubmissionRequest, ipAddress, userAgent string) (*models.ContactSubmission, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Message = strings.TrimSpace(req.Message)
	req.Company = trimOptional(req.Company)
	req.Phone = trimOptional(req.Phone)
	req.Category = trimOptional(req.Category)

	var normalizedIP *string
	if ip := strings.TrimSpace(ipAddress); ip != "" {
		normalizedIP = &ip
	}

	var normalizedUserAgent *string
	if ua := strings.TrimSpace(userAgent); ua != "" {
		normalizedUserAgent = &ua
	}

	return s.repo.CreateSubmission(ctx, req, normalizedIP, normalizedUserAgent)
}

func (s *Service) GetAll(ctx context.Context, statusFilter string) ([]*models.ContactSubmission, error) {
	return s.repo.GetAll(ctx, statusFilter)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.ContactSubmission, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id, status string) (*models.ContactSubmission, error) {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
