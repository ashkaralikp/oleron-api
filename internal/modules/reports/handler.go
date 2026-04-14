package reports

import (
	"net/http"

	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Handler{service: svc}
}

// GetAttendanceReport godoc
// GET /reports/attendance
// Query params: date_from, date_to, user_id, status
func (h *Handler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ := r.Context().Value(middleware.UserBranchIDKey).(string)

	q := r.URL.Query()
	filter := AttendanceFilter{
		DateFrom: q.Get("date_from"),
		DateTo:   q.Get("date_to"),
		UserID:   q.Get("user_id"),
		Status:   q.Get("status"),
	}

	records, err := h.service.GetAttendanceReport(r.Context(), role, branchID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch attendance report")
		return
	}

	response.Success(w, records)
}
