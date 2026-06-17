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
// ASSIGNABLE USERS
// ─────────────────────────────────────────────

func (s *Service) GetAssignableUsers(ctx context.Context, branchIDs []string) ([]AssignableUser, error) {
	return s.repo.FindAssignableUsers(ctx, branchIDs)
}

// ─────────────────────────────────────────────

// ─────────────────────────────────────────────
// VACANCIES
// ─────────────────────────────────────────────

func (s *Service) GetAllVacancies(ctx context.Context, branchIDs []string, userID, role string) ([]models.Vacancy, error) {
	assignedTo := ""
	// Non-super-admin users see vacancies in their branch scope OR vacancies assigned to them
	if len(branchIDs) > 0 {
		assignedTo = userID
	}
	return s.repo.FindAllVacancies(ctx, branchIDs, assignedTo)
}

func (s *Service) GetVacancyByID(ctx context.Context, id string, branchIDs []string, userID string) (*models.Vacancy, error) {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return nil, errors.New("vacancy not found")
	}
	if !s.canAccessVacancy(v, branchIDs, userID) {
		return nil, errors.New("forbidden")
	}
	return v, nil
}

func (s *Service) CreateVacancy(ctx context.Context, branchID, createdBy string, req CreateVacancyRequest) (*models.Vacancy, error) {
	assignedTo := req.AssignedTo
	if len(assignedTo) == 0 {
		assignedTo = []string{createdBy}
	}
	return s.repo.CreateVacancy(ctx, branchID, createdBy, assignedTo, req)
}

func (s *Service) UpdateVacancy(ctx context.Context, id string, branchIDs []string, userID string, req UpdateVacancyRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancy(ctx, id, req)
}

func (s *Service) UpdateVacancyStatus(ctx context.Context, id string, branchIDs []string, userID string, req UpdateVacancyStatusRequest) (*models.Vacancy, error) {
	if err := s.guardVacancy(ctx, id, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.UpdateVacancyStatus(ctx, id, req.Status)
}

func (s *Service) DeleteVacancy(ctx context.Context, id string, branchIDs []string, userID string) error {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return errors.New("vacancy not found")
	}
	if !s.canAccessVacancy(v, branchIDs, userID) {
		return errors.New("forbidden")
	}
	return s.repo.DeleteVacancy(ctx, id)
}

// canAccessVacancy checks if a user can access a vacancy.
// Access is granted if: super_admin (branchIDs == nil), OR branch is in scope, OR user is one of the assignees.
func (s *Service) canAccessVacancy(v *models.Vacancy, branchIDs []string, userID string) bool {
	if branchIDs == nil {
		return true // super_admin
	}
	if containsBranch(branchIDs, v.BranchID) {
		return true
	}
	for _, id := range v.AssignedTo {
		if id == userID {
			return true
		}
	}
	return false
}

// canAccessVacancyByID is a convenience wrapper that fetches the vacancy first.
func (s *Service) canAccessVacancyByID(ctx context.Context, id string, branchIDs []string, userID string) (bool, error) {
	v, err := s.repo.FindVacancyByID(ctx, id)
	if err != nil {
		return false, errors.New("vacancy not found")
	}
	return s.canAccessVacancy(v, branchIDs, userID), nil
}

func (s *Service) guardVacancy(ctx context.Context, id string, branchIDs []string, userID string) error {
	if branchIDs == nil {
		return nil
	}
	ok, err := s.canAccessVacancyByID(ctx, id, branchIDs, userID)
	if err != nil {
		return err
	}
	if !ok {
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

func (s *Service) BulkApply(ctx context.Context, vacancyID string, branchIDs []string, userID string, req BulkApplyRequest) (*BulkApplyResult, error) {
	if err := s.guardVacancy(ctx, vacancyID, branchIDs, userID); err != nil {
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

func (s *Service) GetApplicationsByVacancy(ctx context.Context, vacancyID string, branchIDs []string, userID string, statusFilter string) ([]models.Application, error) {
	if err := s.guardVacancy(ctx, vacancyID, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationsByVacancy(ctx, vacancyID, statusFilter)
}

func (s *Service) GetApplicationByID(ctx context.Context, id string, branchIDs []string, userID string) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.FindApplicationByID(ctx, id)
}

func (s *Service) UpdateApplicationStatus(ctx context.Context, id string, branchIDs []string, userID string, req UpdateApplicationStatusRequest) (*models.Application, error) {
	if err := s.guardApplication(ctx, id, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.UpdateApplicationStatus(ctx, id, req.Status, req.Notes)
}

func (s *Service) DeleteApplication(ctx context.Context, id string, branchIDs []string, userID string) error {
	if err := s.guardApplication(ctx, id, branchIDs, userID); err != nil {
		return err
	}
	return s.repo.DeleteApplication(ctx, id)
}

func (s *Service) guardApplication(ctx context.Context, id string, branchIDs []string, userID string) error {
	if branchIDs == nil {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, id)
	if err != nil {
		return errors.New("application not found")
	}
	if containsBranch(branchIDs, vBranchID) {
		return nil
	}
	// Check if user is the assigned manager for this vacancy
	ok, err := s.isAppAssignedToUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden")
	}
	return nil
}

func (s *Service) isAppAssignedToUser(ctx context.Context, applicationID, userID string) (bool, error) {
	vacancyID, err := s.repo.FindVacancyIDByApplicationID(ctx, applicationID)
	if err != nil {
		return false, errors.New("application not found")
	}
	return s.repo.IsUserAssignedToVacancy(ctx, vacancyID, userID)
}

// ─────────────────────────────────────────────
// INTERVIEWS
// ─────────────────────────────────────────────

func (s *Service) CreateInterview(ctx context.Context, applicationID string, branchIDs []string, userID string, req CreateInterviewRequest) (*models.Interview, error) {
	if err := s.guardApplication(ctx, applicationID, branchIDs, userID); err != nil {
		return nil, err
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, errors.New("invalid scheduled_at: use RFC3339 format (e.g. 2026-04-17T10:00:00Z)")
	}
	return s.repo.CreateInterview(ctx, applicationID, req.InterviewerID, scheduledAt, req.Type, req.Location)
}

func (s *Service) UpdateInterview(ctx context.Context, id string, branchIDs []string, userID string, req UpdateInterviewRequest) (*models.Interview, error) {
	if err := s.guardInterview(ctx, id, branchIDs, userID); err != nil {
		return nil, err
	}
	return s.repo.UpdateInterview(ctx, id, req)
}

func (s *Service) DeleteInterview(ctx context.Context, id string, branchIDs []string, userID string) error {
	if err := s.guardInterview(ctx, id, branchIDs, userID); err != nil {
		return err
	}
	return s.repo.DeleteInterview(ctx, id)
}

func (s *Service) guardInterview(ctx context.Context, id string, branchIDs []string, userID string) error {
	if branchIDs == nil {
		return nil
	}
	vBranchID, err := s.repo.FindVacancyBranchIDByInterviewID(ctx, id)
	if err != nil {
		return errors.New("interview not found")
	}
	if containsBranch(branchIDs, vBranchID) {
		return nil
	}
	// Check if user is the assigned manager for this vacancy
	ok, err := s.isInterviewAssignedToUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden")
	}
	return nil
}

func (s *Service) isInterviewAssignedToUser(ctx context.Context, interviewID, userID string) (bool, error) {
	applicationID, err := s.repo.FindApplicationIDByInterviewID(ctx, interviewID)
	if err != nil {
		return false, errors.New("interview not found")
	}
	return s.isAppAssignedToUser(ctx, applicationID, userID)
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

func (s *Service) Hire(ctx context.Context, applicationID string, branchIDs []string, userID string, req HireRequest) (*HireResult, error) {
	var targetBranchID string
	if branchIDs == nil {
		// super_admin: use the vacancy's branch
		vBranchID, err := s.repo.FindVacancyBranchIDByApplicationID(ctx, applicationID)
		if err != nil {
			return nil, errors.New("application not found")
		}
		targetBranchID = vBranchID
	} else {
		if err := s.guardApplication(ctx, applicationID, branchIDs, userID); err != nil {
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
