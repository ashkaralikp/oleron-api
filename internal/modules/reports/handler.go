package reports

import (
	"net/http"

	"rmp-api/internal/dbscope"
	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db      *pgxpool.Pool
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db, service: NewService(NewRepository(db))}
}

// GET /reports/attendance
func (h *Handler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ := r.Context().Value(middleware.UserBranchIDKey).(string)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	q := r.URL.Query()
	filter := AttendanceFilter{
		DateFrom: q.Get("date_from"),
		DateTo:   q.Get("date_to"),
		UserID:   q.Get("user_id"),
		Status:   q.Get("status"),
	}

	records, err := h.service.GetAttendanceReport(r.Context(), branchIDs, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch attendance report")
		return
	}

	response.Success(w, records)
}
