package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestInstallationStatusEndpoint(t *testing.T) {
	createdAt := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)
	steamCMDInstalledAt := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	router := NewRouter(testLogger(), Services{
		Status: fakeStatusReader{},
		Installation: fakeInstallationStatusReader{
			state: domain.InstallationState{
				ManagedRoot:         "/srv/dst-server-ctl",
				SteamCMDInstalledAt: &steamCMDInstalledAt,
				CreatedAt:           createdAt,
				UpdatedAt:           createdAt,
			},
		},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/installation", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}

	var payload installationResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}

	if payload.ManagedRoot != "/srv/dst-server-ctl" {
		t.Fatalf("ManagedRoot = %q, want /srv/dst-server-ctl", payload.ManagedRoot)
	}
	if !payload.SteamCMDInstalled {
		t.Fatal("SteamCMDInstalled = false, want true")
	}
	if payload.DSTInstalled {
		t.Fatal("DSTInstalled = true, want false")
	}
}

func TestInstallationStatusEndpointReturnsNotFound(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{err: domain.ErrInstallationStateNotFound},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/installation", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusNotFound)
	}
	if !strings.Contains(response.Body.String(), "installation state not initialized") {
		t.Fatalf("body = %q, want not initialized error", response.Body.String())
	}
}

func TestInstallationStatusEndpointReturnsInternalServerError(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{err: errors.New("database unavailable")},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/installation", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusInternalServerError)
	}
}

func TestListInstallTasksEndpoint(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC)
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{
			tasks: []domain.Task{
				{
					ID:        "task-1",
					Type:      domain.TaskTypeInstallDST,
					Status:    domain.TaskStatusRunning,
					Detail:    "Install DST",
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
			},
		},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/install/tasks", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}

	var payload []taskResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("task count = %d, want 1", len(payload))
	}
	if payload[0].ID != "task-1" {
		t.Fatalf("first ID = %q, want task-1", payload[0].ID)
	}
}

func TestStartInstallTasksEndpoint(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{
			startedTasks: []domain.Task{
				{
					ID:        "task-1",
					Type:      domain.TaskTypeInstallSteamCMD,
					Status:    domain.TaskStatusPending,
					Detail:    "Install SteamCMD",
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
			},
		},
	})

	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/install/tasks", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusAccepted)
	}
}

func TestStartInstallTasksEndpointReturnsConflict(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{startErr: domain.ErrInstallAlreadyInProgress},
	})

	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/install/tasks", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusConflict)
	}
}

func TestGetClusterConfigEndpoint(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster: fakeClusterConfigService{
			config: domain.ClusterConfig{
				ClusterName:        "Managed DST",
				ClusterDescription: "test",
				GameMode:           "survival",
				MaxPlayers:         8,
				Language:           "en",
				PauseWhenEmpty:     true,
				Shards: []domain.ShardConfig{
					{Name: domain.ShardMaster, Enabled: true},
					{Name: domain.ShardCaves, Enabled: false},
				},
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
		},
		InstallTasks: fakeInstallationTaskService{},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/cluster", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}

	var payload clusterResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if payload.ClusterName != "Managed DST" {
		t.Fatalf("ClusterName = %q, want Managed DST", payload.ClusterName)
	}
	if len(payload.Shards) != 2 {
		t.Fatalf("shard count = %d, want 2", len(payload.Shards))
	}
}

func TestUpdateClusterConfigEndpoint(t *testing.T) {
	updatedAt := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster: fakeClusterConfigService{
			updated: domain.ClusterConfig{
				ClusterName:        "Managed DST",
				ClusterDescription: "test",
				GameMode:           "endless",
				MaxPlayers:         10,
				Language:           "en",
				PVP:                true,
				PauseWhenEmpty:     false,
				Shards: []domain.ShardConfig{
					{Name: domain.ShardMaster, Enabled: true},
					{Name: domain.ShardCaves, Enabled: true},
				},
				CreatedAt: updatedAt.Add(-time.Hour),
				UpdatedAt: updatedAt,
			},
		},
		InstallTasks: fakeInstallationTaskService{},
	})

	body := strings.NewReader(`{"clusterName":"Managed DST","clusterDescription":"test","gameMode":"endless","maxPlayers":10,"language":"en","pvp":true,"pauseWhenEmpty":false,"shards":[{"name":"Master","enabled":true},{"name":"Caves","enabled":true}]}`)
	request := httptest.NewRequest(nethttp.MethodPut, "/api/v1/cluster", body)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}
}

func TestUpdateClusterConfigEndpointReturnsBadRequest(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{updateErr: errors.Join(domain.ErrInvalidClusterConfig, errors.New("cluster name is required"))},
		InstallTasks: fakeInstallationTaskService{},
	})

	body := strings.NewReader(`{"clusterName":"","gameMode":"survival","maxPlayers":6,"language":"en","shards":[{"name":"Master","enabled":true}]}`)
	request := httptest.NewRequest(nethttp.MethodPut, "/api/v1/cluster", body)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusBadRequest)
	}
}

func TestUpdateClusterConfigEndpointReturnsInternalServerError(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{updateErr: errors.New("database unavailable")},
		InstallTasks: fakeInstallationTaskService{},
	})

	body := strings.NewReader(`{"clusterName":"Managed DST","gameMode":"survival","maxPlayers":6,"language":"en","shards":[{"name":"Master","enabled":true}]}`)
	request := httptest.NewRequest(nethttp.MethodPut, "/api/v1/cluster", body)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusInternalServerError)
	}
}

func TestRuntimeStatusEndpoint(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
		Runtime: fakeRuntimeService{
			status: domain.RuntimeStatus{
				Status: domain.ServerStatusRunning,
				Shards: []domain.ShardState{
					{Name: domain.ShardMaster, Running: true, PID: 101},
				},
			},
		},
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/runtime", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}

	var payload runtimeResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if payload.Status != "running" {
		t.Fatalf("status = %q, want running", payload.Status)
	}
	if len(payload.Shards) != 1 || payload.Shards[0].PID != 101 {
		t.Fatalf("shards = %#v, want master pid 101", payload.Shards)
	}
}

func TestRuntimeStartEndpoint(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
		Runtime: fakeRuntimeService{
			status: domain.RuntimeStatus{
				Status: domain.ServerStatusRunning,
				Shards: []domain.ShardState{{Name: domain.ShardMaster, Running: true, PID: 101}},
			},
		},
	})

	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/runtime/start", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusAccepted)
	}
}

func TestRuntimeStartEndpointReturnsConflict(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
		Runtime:      fakeRuntimeService{startErr: domain.ErrDSTNotInstalled},
	})

	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/runtime/start", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusConflict)
	}
}

func TestRuntimeStopEndpoint(t *testing.T) {
	router := NewRouter(testLogger(), Services{
		Status:       fakeStatusReader{},
		Installation: fakeInstallationStatusReader{},
		Cluster:      fakeClusterConfigService{},
		InstallTasks: fakeInstallationTaskService{},
		Runtime:      fakeRuntimeService{status: domain.RuntimeStatus{Status: domain.ServerStatusStopped}},
	})

	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/runtime/stop", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusOK)
	}
}

type fakeStatusReader struct{}

func (fakeStatusReader) Status() domain.Status {
	return domain.Status{Version: "test", Status: domain.ServerStatusUnknown}
}

type fakeInstallationStatusReader struct {
	state domain.InstallationState
	err   error
}

func (r fakeInstallationStatusReader) Status(context.Context) (domain.InstallationState, error) {
	if r.err != nil {
		return domain.InstallationState{}, r.err
	}
	return r.state, nil
}

type fakeInstallationTaskService struct {
	tasks        []domain.Task
	listErr      error
	startedTasks []domain.Task
	startErr     error
}

type fakeClusterConfigService struct {
	config    domain.ClusterConfig
	getErr    error
	updated   domain.ClusterConfig
	updateErr error
}

type fakeRuntimeService struct {
	status    domain.RuntimeStatus
	statusErr error
	startErr  error
	stopErr   error
}

func (s fakeInstallationTaskService) ListTasks(context.Context) ([]domain.Task, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.tasks, nil
}

func (s fakeInstallationTaskService) Start(context.Context) ([]domain.Task, error) {
	if s.startErr != nil {
		return nil, s.startErr
	}
	return s.startedTasks, nil
}

func (s fakeClusterConfigService) Get(context.Context) (domain.ClusterConfig, error) {
	if s.getErr != nil {
		return domain.ClusterConfig{}, s.getErr
	}
	return s.config, nil
}

func (s fakeClusterConfigService) Update(_ context.Context, config domain.ClusterConfig) (domain.ClusterConfig, error) {
	if s.updateErr != nil {
		return domain.ClusterConfig{}, s.updateErr
	}
	return s.updated, nil
}

func (s fakeRuntimeService) Status(context.Context) (domain.RuntimeStatus, error) {
	if s.statusErr != nil {
		return domain.RuntimeStatus{}, s.statusErr
	}
	return s.status, nil
}

func (s fakeRuntimeService) Start(context.Context) error {
	return s.startErr
}

func (s fakeRuntimeService) Stop(context.Context) error {
	return s.stopErr
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
