package reports

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

// GetAttendanceReport returns attendance records visible to the caller.
// nil branchIDs = super_admin = all branches.
func (s *Service) GetAttendanceReport(ctx context.Context, branchIDs []string, f AttendanceFilter) ([]models.Attendance, error) {
	f.BranchIDs = branchIDs
	return s.repo.FindAttendance(ctx, f)
}
