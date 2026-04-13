package report

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	report, err := h.service.Generate(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate report")
		return
	}

	response.Success(w, report)
}
