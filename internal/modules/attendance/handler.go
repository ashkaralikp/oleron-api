package attendance

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

// POST /attendance/punch
// Smart endpoint: punch-in if no record today, punch-out if already punched in.
func (h *Handler) Punch(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	result, err := h.service.Punch(r.Context(), userID)
	if err != nil {
		switch err.Error() {
		case "already punched out for today":
			response.Error(w, http.StatusConflict, err.Error())
		case "today is a public holiday", "today is not a working day":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to process punch")
		}
		return
	}

	if result.Action == "punch_in" {
		response.Created(w, result)
	} else {
		response.Success(w, result)
	}
}

// GET /attendance/today
// Returns the calling user's attendance record for today.
func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	result, err := h.service.GetToday(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch today's attendance")
		return
	}
	response.Success(w, result)
}
