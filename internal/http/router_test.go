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
	})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/installation", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, nethttp.StatusInternalServerError)
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

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
