package server

import (
	"fmt"
	"net/http"

	"clinic-api/internal/config"
	"clinic-api/internal/router"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg    *config.Config
	db     *pgxpool.Pool
	router http.Handler
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	s := &Server{cfg: cfg, db: db}
	s.router = router.Setup(cfg, db)
	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return http.ListenAndServe(addr, s.router)
}
