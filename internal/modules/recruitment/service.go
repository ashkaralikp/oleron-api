package recruitment

import (
	"context"
	"errors"
	"time"

	"rmp-api/internal/models"
	"rmp-api/pkg/hash"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ─────────────────────────────────────────────
// PUBLIC VACANCIES
// ─────────────────────────────────────────────

func (s *Service) GetPublicVacancies(ctx context.Context) ([]PublicVacancy, error) {
	return s.repo.FindPublicVacancies(ctx)
}

// ─────────────────────────────────────────────
// VACANCIES
// ─────────────────────────────────────────────

func (s *Service) GetAllVacancies(ctx context.Context, branchIDs []string) ([]models.Vacancy, error) {
	return s.repo.FindAllVacancies(ctx, branchIDs)
}

func (s *Service) GetVacancyByID(ctx context.Context, id string, branchIDs []string) (*models.Vacancy, error) {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return nil, errors.New("vacancy not found")
	}
	if !containsBranch(branchIDs, v.BranchID) {
		return nil, errors.New("forbidden")
	}
	return v, nil
}

func (s *Service) CreateVacancy(ctx context.Context, branchID, createdBy string, req CreateVacancyRequest) (*models.Vacancy, error) {
	return s.repo.CreateVacancy(ctx, branchID, createdBy, req)
}

func (s *Service) UpdateVacancy(ctx context.Context, id string, branchIDs []string, req UpdateVacancyRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancy(ctx, id, req)
}

func (s *Service) UpdateVacancyStatus(ctx context.Context, id string, branchIDs []string, req UpdateVacancyStatusRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancyStatus(ctx, id, req.Status)
}

func (s *Service) DeleteVacancy(ctx context.Context, id string, branchIDs []string) error {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return errors.New("vacancy not found")
	}
	if !containsBranch(branchIDs, v.BranchID) {
		return errors.New("forbidden")
	}
	if v.Status != "draft" {
		return errors.New("only draft vacancies can be deleted")
	}
	return s.repo.DeleteVacancy(ctx, id)
}

func (s *Service) guardVacancy(ctx context.Context, id string, branchIDs []string) error {
	if branchIDs == nil {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchID(ctx, id)
	if err != nil {
		return errors.New("vacancy not found")
	}
	if !containsBranch(branchIDs, vBranchID) {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// APPLICATIONS
// ─────────────────────────────────────────────

func (s *Service) Apply(ctx context.Context, vacancyID string, req ApplyRequest) (*models.Application, error) {
	return s.repo.CreateApplication(ctx, vacancyID, req)
}

func (s *Service) BulkApply(ctx context.Context, vacancyID string, branchIDs []string, req BulkApplyRequest) (*BulkApplyResult, error) {
	if err := s.guardVacancy(ctx, vacancyID, branchIDs); err != nil {
		return nil, err
	}
	result := &BulkApplyResult{Total: len(req.Applications)}
	for _, app := range req.Applications {
		created, err := s.repo.CreateApplication(ctx, vacancyID, app)
		row := ApplicationResult{Email: app.Email}
		if err != nil {
			row.Status = "failed"
			msg := err.Error()
			row.Error = &msg
			result.Failed++
		} else {
			row.Status = "created"
			row.ID = &created.ID
			result.Succeeded++
		}
		result.Results = append(result.Results, row)
	}
	return result, nil
}

func (s *Service) GetApplicationsByVacancy(ctx context.Context, vacancyID string, branchIDs []string, statusFilter string) ([]models.Application, error) {
	if err := s.guardVacancy(ctx, vacancyID, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationsByVacancy(ctx, vacancyID, statusFilter)
}

func (s *Service) GetApplicationByID(ctx context.Context, id string, branchIDs []string) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationByID(ctx, id)
}

func (s *Service) UpdateApplicationStatus(ctx context.Context, id string, branchIDs []string, req UpdateApplicationStatusRequest) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.UpdateApplicationStatus(ctx, id, req.Status, req.Notes)
}

func (s *Service) DeleteApplication(ctx context.Context, id string, branchIDs []string) error {
	if err := s.guardApplication(ctx, id, branchIDs); err != nil {
		return err
	}
	return s.repo.DeleteApplication(ctx, id)
}

func (s *Service) guardApplication(ctx context.Context, id string, branchIDs []string) error {
	if branchIDs == nil {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, id)
	if err != nil {
		return errors.New("application not found")
	}
	if !containsBranch(branchIDs, vBranchID) {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// INTERVIEWS
// ─────────────────────────────────────────────

func (s *Service) CreateInterview(ctx context.Context, applicationID string, branchIDs []string, req CreateInterviewRequest) (*models.Interview, error) {
	if err := s.guardApplication(ctx, applicationID, branchIDs); err != nil {
		return nil, err
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, errors.New("invalid scheduled_at: use RFC3339 format (e.g. 2026-04-17T10:00:00Z)")
	}
	return s.repo.CreateInterview(ctx, applicationID, req.InterviewerID, scheduledAt, req.Type, req.Location)
}

func (s *Service) UpdateInterview(ctx context.Context, id string, branchIDs []string, req UpdateInterviewRequest) (*models.Interview, error) {
	if err := s.guardInterview(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.UpdateInterview(ctx, id, req)
}

func (s *Service) DeleteInterview(ctx context.Context, id string, branchIDs []string) error {
	if err := s.guardInterview(ctx, id, branchIDs); err != nil {
		return err
	}
	return s.repo.DeleteInterview(ctx, id)
}

func (s *Service) guardInterview(ctx context.Context, id string, branchIDs []string) error {
	if branchIDs == nil {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByInterviewID(ctx, id)
	if err != nil {
		return errors.New("interview not found")
	}
	if !containsBranch(branchIDs, vBranchID) {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

func (s *Service) Hire(ctx context.Context, applicationID string, branchIDs []string, req HireRequest) (*HireResult, error) {
	var targetBranchID string
	if branchIDs == nil {
		// super_admin: use the vacancy's branch
		vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, applicationID)
		if err != nil {
			return nil, errors.New("application not found")
		}
		targetBranchID = vBranchID
	} else {
		if err := s.guardApplication(ctx, applicationID, branchIDs); err != nil {
			return nil, err
		}
		// Use the vacancy's branch (so we always assign to the correct branch)
		vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, applicationID)
		if err != nil {
			return nil, errors.New("application not found")
		}
		targetBranchID = vBranchID
	}

	passwordHash, err := hash.HashPassword(req.TempPassword)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	return s.repo.HireApplicant(ctx, applicationID, targetBranchID, passwordHash, req)
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
