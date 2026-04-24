package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"dst-server-ctl/internal/domain"
)

type Services struct {
	Status       StatusReader
	Installation InstallationStatusReader
	InstallTasks InstallationTaskService
}

type StatusReader interface {
	Status() domain.Status
}

type InstallationStatusReader interface {
	Status(ctx context.Context) (domain.InstallationState, error)
}

type InstallationTaskService interface {
	ListTasks(ctx context.Context) ([]domain.Task, error)
	Start(ctx context.Context) ([]domain.Task, error)
}

func NewRouter(logger *slog.Logger, services Services) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, services.Status.Status())
	})

	mux.HandleFunc("GET /api/v1/installation", func(w http.ResponseWriter, r *http.Request) {
		state, err := services.Installation.Status(r.Context())
		if errors.Is(err, domain.ErrInstallationStateNotFound) {
			respondError(w, http.StatusNotFound, "installation state not initialized")
			return
		}
		if err != nil {
			logger.Error("installation status failed", "error", err)
			respondError(w, http.StatusInternalServerError, "installation status unavailable")
			return
		}

		respondJSON(w, installationResponseFromDomain(state))
	})

	mux.HandleFunc("GET /api/v1/install/tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks, err := services.InstallTasks.ListTasks(r.Context())
		if err != nil {
			logger.Error("list install tasks failed", "error", err)
			respondError(w, http.StatusInternalServerError, "install tasks unavailable")
			return
		}

		respondJSON(w, taskListResponseFromDomain(tasks))
	})

	mux.HandleFunc("POST /api/v1/install/tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks, err := services.InstallTasks.Start(r.Context())
		switch {
		case errors.Is(err, domain.ErrInstallAlreadyInProgress):
			respondError(w, http.StatusConflict, "install already in progress")
			return
		case errors.Is(err, domain.ErrInstallNotRequired):
			respondError(w, http.StatusConflict, "install not required")
			return
		case err != nil:
			logger.Error("start install failed", "error", err)
			respondError(w, http.StatusInternalServerError, "install start failed")
			return
		}

		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, taskListResponseFromDomain(tasks))
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("frontend placeholder served", "path", r.URL.Path)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("dst-server-ctl frontend is not embedded yet\n"))
	})

	return mux
}

func respondJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

type installationResponse struct {
	ManagedRoot       string    `json:"managedRoot"`
	SteamCMDInstalled bool      `json:"steamcmdInstalled"`
	DSTInstalled      bool      `json:"dstInstalled"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type taskResponse struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	Detail     string     `json:"detail"`
	Error      string     `json:"error,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func installationResponseFromDomain(state domain.InstallationState) installationResponse {
	return installationResponse{
		ManagedRoot:       state.ManagedRoot,
		SteamCMDInstalled: state.SteamCMDInstalledAt != nil,
		DSTInstalled:      state.DSTInstalledAt != nil,
		CreatedAt:         state.CreatedAt,
		UpdatedAt:         state.UpdatedAt,
	}
}

func taskListResponseFromDomain(tasks []domain.Task) []taskResponse {
	response := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, taskResponse{
			ID:         string(task.ID),
			Type:       string(task.Type),
			Status:     string(task.Status),
			Detail:     task.Detail,
			Error:      task.Error,
			StartedAt:  task.StartedAt,
			FinishedAt: task.FinishedAt,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  task.UpdatedAt,
		})
	}
	return response
}
