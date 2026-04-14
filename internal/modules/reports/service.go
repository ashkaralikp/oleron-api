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

// GetAttendanceReport returns attendance records filtered by role.
// super_admin sees all branches; admin and manager see their branch only.
func (s *Service) GetAttendanceReport(ctx context.Context, role, branchID string, f AttendanceFilter) ([]models.Attendance, error) {
	if role != "super_admin" {
		f.BranchID = branchID
	}
	return s.repo.FindAttendance(ctx, f)
}
