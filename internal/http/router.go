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
}

type StatusReader interface {
	Status() domain.Status
}

type InstallationStatusReader interface {
	Status(ctx context.Context) (domain.InstallationState, error)
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

func installationResponseFromDomain(state domain.InstallationState) installationResponse {
	return installationResponse{
		ManagedRoot:       state.ManagedRoot,
		SteamCMDInstalled: state.SteamCMDInstalledAt != nil,
		DSTInstalled:      state.DSTInstalledAt != nil,
		CreatedAt:         state.CreatedAt,
		UpdatedAt:         state.UpdatedAt,
	}
}
