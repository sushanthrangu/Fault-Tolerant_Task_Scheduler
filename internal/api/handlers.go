package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"task-scheduler/internal/repo"
)

type Handlers struct {
	Repo repo.JobRepository
}

func NewHandlers(r repo.JobRepository) *Handlers {
	return &Handlers{Repo: r}
}

type createJobReq struct {
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	MaxAttempts int             `json:"max_attempts"`
}

func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handlers) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Type) == "" || len(req.Payload) == 0 {
		http.Error(w, `{"error":"type_and_payload_required"}`, http.StatusBadRequest)
		return
	}
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}

	idempotency := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	var idemPtr *string
	if idempotency != "" {
		idemPtr = &idempotency
	}

	jobID := newID()

	job, err := h.Repo.CreateJob(r.Context(), jobID, req.Type, req.Payload, req.MaxAttempts, idemPtr)
	if err != nil {
		http.Error(w, `{"error":"create_failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(job)
}

func (h *Handlers) GetJob(w http.ResponseWriter, r *http.Request) {
	// Expect /jobs/{id}
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	job, err := h.Repo.GetJobByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"fetch_failed"}`, http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(job)
}

// tiny ID generator (replace with uuid if you already use one)
func newID() string {
	return strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")
}
