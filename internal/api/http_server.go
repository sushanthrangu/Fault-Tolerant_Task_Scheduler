package api

import (
	"net/http"

	"task-scheduler/internal/repo"
)

type Server struct {
	h http.Handler
}

func NewServer(r repo.JobRepository) *Server {
	handlers := NewHandlers(r)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.Healthz)

	// Routes:
	// POST /jobs
	// GET  /jobs/{id}
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			handlers.CreateJob(w, req)
			return
		}
		http.NotFound(w, req)
	})

	mux.HandleFunc("/jobs/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			handlers.GetJob(w, req)
			return
		}
		http.NotFound(w, req)
	})

	return &Server{h: withMiddleware(mux)}
}

func (s *Server) Handler() http.Handler {
	return s.h
}
