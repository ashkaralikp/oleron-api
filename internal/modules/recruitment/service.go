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
// VACANCIES
// ─────────────────────────────────────────────

func (s *Service) GetAllVacancies(ctx context.Context, role, branchID string) ([]models.Vacancy, error) {
	if role == "super_admin" {
		return s.repo.FindAllVacancies(ctx, "")
	}
	return s.repo.FindAllVacancies(ctx, branchID)
}

func (s *Service) GetVacancyByID(ctx context.Context, id, role, branchID string) (*models.Vacancy, error) {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return nil, errors.New("vacancy not found")
	}
	if role != "super_admin" && v.BranchID != branchID {
		return nil, errors.New("forbidden")
	}
	return v, nil
}

func (s *Service) CreateVacancy(ctx context.Context, branchID, createdBy string, req CreateVacancyRequest) (*models.Vacancy, error) {
	return s.repo.CreateVacancy(ctx, branchID, createdBy, req)
}

func (s *Service) UpdateVacancy(ctx context.Context, id, role, branchID string, req UpdateVacancyRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancy(ctx, id, req)
}

func (s *Service) UpdateVacancyStatus(ctx context.Context, id, role, branchID string, req UpdateVacancyStatusRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancyStatus(ctx, id, req.Status)
}

func (s *Service) DeleteVacancy(ctx context.Context, id, role, branchID string) error {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return errors.New("vacancy not found")
	}
	if role != "super_admin" && v.BranchID != branchID {
		return errors.New("forbidden")
	}
	if v.Status != "draft" {
		return errors.New("only draft vacancies can be deleted")
	}
	return s.repo.DeleteVacancy(ctx, id)
}

// guardVacancy checks branch ownership for non-super_admin.
func (s *Service) guardVacancy(ctx context.Context, id, role, branchID string) error {
	if role == "super_admin" {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchID(ctx, id)
	if err != nil {
		return errors.New("vacancy not found")
	}
	if vBranchID != branchID {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// APPLICATIONS
// ─────────────────────────────────────────────

// Apply is the public endpoint — no role or branch ownership check.
func (s *Service) Apply(ctx context.Context, vacancyID string, req ApplyRequest) (*models.Application, error) {
	return s.repo.CreateApplication(ctx, vacancyID, req)
}

func (s *Service) GetApplicationsByVacancy(ctx context.Context, vacancyID, role, branchID, statusFilter string) ([]models.Application, error) {
	if err := s.guardVacancy(ctx, vacancyID, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationsByVacancy(ctx, vacancyID, statusFilter)
}

func (s *Service) GetApplicationByID(ctx context.Context, id, role, branchID string) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationByID(ctx, id)
}

func (s *Service) UpdateApplicationStatus(ctx context.Context, id, role, branchID string, req UpdateApplicationStatusRequest) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.UpdateApplicationStatus(ctx, id, req.Status, req.Notes)
}

func (s *Service) DeleteApplication(ctx context.Context, id, role, branchID string) error {
	if err := s.guardApplication(ctx, id, role, branchID); err != nil {
		return err
	}
	return s.repo.DeleteApplication(ctx, id)
}

// guardApplication resolves the vacancy's branch_id from the application and checks ownership.
func (s *Service) guardApplication(ctx context.Context, id, role, branchID string) error {
	if role == "super_admin" {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, id)
	if err != nil {
		return errors.New("application not found")
	}
	if vBranchID != branchID {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// INTERVIEWS
// ─────────────────────────────────────────────

func (s *Service) CreateInterview(ctx context.Context, applicationID, role, branchID string, req CreateInterviewRequest) (*models.Interview, error) {
	if err := s.guardApplication(ctx, applicationID, role, branchID); err != nil {
		return nil, err
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, errors.New("invalid scheduled_at: use RFC3339 format (e.g. 2026-04-17T10:00:00Z)")
	}
	return s.repo.CreateInterview(ctx, applicationID, req.InterviewerID, scheduledAt, req.Type, req.Location)
}

func (s *Service) UpdateInterview(ctx context.Context, id, role, branchID string, req UpdateInterviewRequest) (*models.Interview, error) {
	if err := s.guardInterview(ctx, id, role, branchID); err != nil {
		return nil, err
	}
	return s.repo.UpdateInterview(ctx, id, req)
}

func (s *Service) DeleteInterview(ctx context.Context, id, role, branchID string) error {
	if err := s.guardInterview(ctx, id, role, branchID); err != nil {
		return err
	}
	return s.repo.DeleteInterview(ctx, id)
}

// guardInterview resolves the vacancy's branch_id through interview → application → vacancy.
func (s *Service) guardInterview(ctx context.Context, id, role, branchID string) error {
	if role == "super_admin" {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByInterviewID(ctx, id)
	if err != nil {
		return errors.New("interview not found")
	}
	if vBranchID != branchID {
		return errors.New("forbidden")
	}
	return nil
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

func (s *Service) Hire(ctx context.Context, applicationID, role, branchID string, req HireRequest) (*HireResult, error) {
	// Resolve target branch: super_admin uses the vacancy's branch; others use their own.
	var targetBranchID string
	if role == "super_admin" {
		vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, applicationID)
		if err != nil {
			return nil, errors.New("application not found")
		}
		targetBranchID = vBranchID
	} else {
		if err := s.guardApplication(ctx, applicationID, role, branchID); err != nil {
			return nil, err
		}
		targetBranchID = branchID
	}

	passwordHash, err := hash.HashPassword(req.TempPassword)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	return s.repo.HireApplicant(ctx, applicationID, targetBranchID, passwordHash, req)
}
