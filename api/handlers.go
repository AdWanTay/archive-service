package api

import (
	"archive-service/config"
	"archive-service/task"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	tm  *task.Manager
	cfg config.Config
}

func NewHandler(tm *task.Manager, cfg config.Config) *Handler {
	return &Handler{tm: tm, cfg: cfg}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID, err := h.tm.CreateTask()
	if err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"task_id": taskID}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
func (h *Handler) RouteTaskSubpaths(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/task/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	taskID, action := parts[0], parts[1]
	switch action {
	case "add":
		h.AddFile(w, r, taskID)
	case "status":
		h.Status(w, taskID)
	default:
		http.NotFound(w, r)
	}
}
func (h *Handler) AddFile(w http.ResponseWriter, r *http.Request, taskID string) {
	type Req struct {
		URL string `json:"url"`
	}

	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if err := h.tm.AddFile(taskID, req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "File queued"}); err != nil {
		http.Error(w, "Failed write response", http.StatusInternalServerError)
	}
}

func (h *Handler) Status(w http.ResponseWriter, taskID string) {
	res, err := h.tm.GetStatus(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed  write response", http.StatusInternalServerError)
	}
}
