package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"dst-server-ctl/internal/domain"
)

type Services struct {
	Status       StatusReader
	Installation InstallationStatusReader
	Cluster      ClusterConfigManager
	InstallTasks InstallationTaskService
	Runtime      RuntimeService
	RuntimeLogs  RuntimeLogService
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

type ClusterConfigManager interface {
	Get(ctx context.Context) (domain.ClusterConfig, error)
	Update(ctx context.Context, config domain.ClusterConfig) (domain.ClusterConfig, error)
}

type RuntimeService interface {
	Status(ctx context.Context) (domain.RuntimeStatus, error)
	Start(ctx context.Context) error
	Restart(ctx context.Context) error
	Stop(ctx context.Context) error
}

type RuntimeLogService interface {
	Get(ctx context.Context, shard domain.ShardName, maxLines int) ([]string, error)
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

	mux.HandleFunc("GET /api/v1/cluster", func(w http.ResponseWriter, r *http.Request) {
		config, err := services.Cluster.Get(r.Context())
		if errors.Is(err, domain.ErrClusterConfigNotFound) {
			respondError(w, http.StatusNotFound, "cluster config not initialized")
			return
		}
		if err != nil {
			logger.Error("get cluster config failed", "error", err)
			respondError(w, http.StatusInternalServerError, "cluster config unavailable")
			return
		}

		respondJSON(w, clusterResponseFromDomain(config))
	})

	mux.HandleFunc("PUT /api/v1/cluster", func(w http.ResponseWriter, r *http.Request) {
		var request updateClusterRequest
		if err := decodeJSON(r, &request); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		config, err := services.Cluster.Update(r.Context(), request.toDomain())
		if errors.Is(err, domain.ErrClusterConfigNotFound) {
			respondError(w, http.StatusNotFound, "cluster config not initialized")
			return
		}
		if errors.Is(err, domain.ErrInvalidClusterConfig) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err != nil {
			logger.Error("update cluster config failed", "error", err)
			respondError(w, http.StatusInternalServerError, "cluster config update failed")
			return
		}

		respondJSON(w, clusterResponseFromDomain(config))
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

	mux.HandleFunc("GET /api/v1/runtime", func(w http.ResponseWriter, r *http.Request) {
		status, err := services.Runtime.Status(r.Context())
		if err != nil {
			logger.Error("runtime status failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime status unavailable")
			return
		}

		respondJSON(w, runtimeResponseFromDomain(status))
	})

	mux.HandleFunc("POST /api/v1/runtime/start", func(w http.ResponseWriter, r *http.Request) {
		err := services.Runtime.Start(r.Context())
		switch {
		case errors.Is(err, domain.ErrDSTNotInstalled):
			respondError(w, http.StatusConflict, "dst is not installed")
			return
		case errors.Is(err, domain.ErrServerAlreadyRunning):
			respondError(w, http.StatusConflict, "server already running")
			return
		case err != nil:
			logger.Error("runtime start failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime start failed")
			return
		}

		status, err := services.Runtime.Status(r.Context())
		if err != nil {
			logger.Error("runtime status after start failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime status unavailable")
			return
		}

		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, runtimeResponseFromDomain(status))
	})

	mux.HandleFunc("POST /api/v1/runtime/stop", func(w http.ResponseWriter, r *http.Request) {
		err := services.Runtime.Stop(r.Context())
		switch {
		case errors.Is(err, domain.ErrServerNotRunning):
			respondError(w, http.StatusConflict, "server is not running")
			return
		case err != nil:
			logger.Error("runtime stop failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime stop failed")
			return
		}

		status, err := services.Runtime.Status(r.Context())
		if err != nil {
			logger.Error("runtime status after stop failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime status unavailable")
			return
		}

		respondJSON(w, runtimeResponseFromDomain(status))
	})

	mux.HandleFunc("POST /api/v1/runtime/restart", func(w http.ResponseWriter, r *http.Request) {
		err := services.Runtime.Restart(r.Context())
		switch {
		case errors.Is(err, domain.ErrDSTNotInstalled):
			respondError(w, http.StatusConflict, "dst is not installed")
			return
		case err != nil:
			logger.Error("runtime restart failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime restart failed")
			return
		}

		status, err := services.Runtime.Status(r.Context())
		if err != nil {
			logger.Error("runtime status after restart failed", "error", err)
			respondError(w, http.StatusInternalServerError, "runtime status unavailable")
			return
		}

		w.WriteHeader(http.StatusAccepted)
		respondJSON(w, runtimeResponseFromDomain(status))
	})

	mux.HandleFunc("GET /api/v1/runtime/logs", func(w http.ResponseWriter, r *http.Request) {
		shard := domain.ShardName(r.URL.Query().Get("shard"))
		lines := 200
		if value := r.URL.Query().Get("lines"); value != "" {
			var parsed int
			if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
				respondError(w, http.StatusBadRequest, "lines must be an integer")
				return
			}
			lines = parsed
		}

		entries, err := services.RuntimeLogs.Get(r.Context(), shard, lines)
		if errors.Is(err, domain.ErrInvalidShard) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err != nil {
			logger.Error("runtime logs failed", "error", err, "shard", shard)
			respondError(w, http.StatusInternalServerError, "runtime logs unavailable")
			return
		}

		respondJSON(w, runtimeLogResponse{
			Shard: string(shard),
			Lines: entries,
		})
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

type clusterResponse struct {
	ClusterName        string          `json:"clusterName"`
	ClusterDescription string          `json:"clusterDescription"`
	GameMode           string          `json:"gameMode"`
	MaxPlayers         int             `json:"maxPlayers"`
	Language           string          `json:"language"`
	PVP                bool            `json:"pvp"`
	PauseWhenEmpty     bool            `json:"pauseWhenEmpty"`
	Shards             []shardResponse `json:"shards"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

type shardResponse struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type runtimeResponse struct {
	Status          string               `json:"status"`
	Shards          []runtimeShardStatus `json:"shards"`
	RestartRequired bool                 `json:"restartRequired"`
	LastError       string               `json:"lastError,omitempty"`
}

type runtimeShardStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	PID     int    `json:"pid,omitempty"`
}

type runtimeLogResponse struct {
	Shard string   `json:"shard"`
	Lines []string `json:"lines"`
}

type updateClusterRequest struct {
	ClusterName        string        `json:"clusterName"`
	ClusterDescription string        `json:"clusterDescription"`
	GameMode           string        `json:"gameMode"`
	MaxPlayers         int           `json:"maxPlayers"`
	Language           string        `json:"language"`
	PVP                bool          `json:"pvp"`
	PauseWhenEmpty     bool          `json:"pauseWhenEmpty"`
	Shards             []shardConfig `json:"shards"`
}

type shardConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
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

func clusterResponseFromDomain(config domain.ClusterConfig) clusterResponse {
	shards := make([]shardResponse, 0, len(config.Shards))
	for _, shard := range config.Shards {
		shards = append(shards, shardResponse{
			Name:    string(shard.Name),
			Enabled: shard.Enabled,
		})
	}

	return clusterResponse{
		ClusterName:        config.ClusterName,
		ClusterDescription: config.ClusterDescription,
		GameMode:           config.GameMode,
		MaxPlayers:         config.MaxPlayers,
		Language:           config.Language,
		PVP:                config.PVP,
		PauseWhenEmpty:     config.PauseWhenEmpty,
		Shards:             shards,
		CreatedAt:          config.CreatedAt,
		UpdatedAt:          config.UpdatedAt,
	}
}

func runtimeResponseFromDomain(status domain.RuntimeStatus) runtimeResponse {
	shards := make([]runtimeShardStatus, 0, len(status.Shards))
	for _, shard := range status.Shards {
		shards = append(shards, runtimeShardStatus{
			Name:    string(shard.Name),
			Running: shard.Running,
			PID:     shard.PID,
		})
	}

	return runtimeResponse{
		Status:          string(status.Status),
		Shards:          shards,
		RestartRequired: status.RestartRequired,
		LastError:       status.LastError,
	}
}

func (r updateClusterRequest) toDomain() domain.ClusterConfig {
	shards := make([]domain.ShardConfig, 0, len(r.Shards))
	for _, shard := range r.Shards {
		shards = append(shards, domain.ShardConfig{
			Name:    domain.ShardName(shard.Name),
			Enabled: shard.Enabled,
		})
	}

	return domain.ClusterConfig{
		ClusterName:        r.ClusterName,
		ClusterDescription: r.ClusterDescription,
		GameMode:           r.GameMode,
		MaxPlayers:         r.MaxPlayers,
		Language:           r.Language,
		PVP:                r.PVP,
		PauseWhenEmpty:     r.PauseWhenEmpty,
		Shards:             shards,
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
