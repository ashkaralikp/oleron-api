package contact

import (
	"context"
	"fmt"
	"log"
	"strings"

	"rmp-api/internal/models"
	"rmp-api/pkg/email"
)

type Service struct {
	repo   *Repository
	mailer *email.Sender
	mailTo string
}

func NewService(repo *Repository, mailer *email.Sender, mailTo string) *Service {
	return &Service{repo: repo, mailer: mailer, mailTo: mailTo}
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

	submission, err := s.repo.CreateSubmission(ctx, req, normalizedIP, normalizedUserAgent)
	if err != nil {
		return nil, err
	}

	if s.mailer != nil && s.mailTo != "" {
		go func() {
			if emailErr := s.sendNotification(submission); emailErr != nil {
				log.Printf("contact email notification failed: %v", emailErr)
			}
		}()
	}

	return submission, nil
}

func (s *Service) sendNotification(sub *models.ContactSubmission) error {
	company := "—"
	if sub.Company != nil {
		company = *sub.Company
	}
	phone := "—"
	if sub.Phone != nil {
		phone = *sub.Phone
	}
	category := "—"
	if sub.Category != nil {
		category = *sub.Category
	}

	body := fmt.Sprintf(
		"New inquiry received on the Oleron website.\r\n\r\n"+
			"Name:     %s\r\n"+
			"Company:  %s\r\n"+
			"Email:    %s\r\n"+
			"Phone:    %s\r\n"+
			"Category: %s\r\n\r\n"+
			"Message:\r\n%s\r\n\r\n"+
			"---\r\nOleron Website",
		sub.Name, company, sub.Email, phone, category, sub.Message,
	)

	subject := fmt.Sprintf("New Inquiry from %s", sub.Name)
	return s.mailer.Send(s.mailTo, subject, body)
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
