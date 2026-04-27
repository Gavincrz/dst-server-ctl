package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/adapter/dstconfig"
	"dst-server-ctl/internal/adapter/dstserver"
	"dst-server-ctl/internal/adapter/logtail"
	"dst-server-ctl/internal/adapter/paths"
	"dst-server-ctl/internal/adapter/sqlite"
	"dst-server-ctl/internal/adapter/steamcmd"
	"dst-server-ctl/internal/adapter/taskid"
	apphttp "dst-server-ctl/internal/http"
	"dst-server-ctl/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	layout := paths.ManagedLayout(paths.DefaultManagedRoot())
	if err := paths.EnsureManagedLayout(layout); err != nil {
		logger.Error("managed root initialization failed", "root", layout.Root, "error", err)
		os.Exit(1)
	}

	store, err := sqlite.Open(ctx, filepath.Join(layout.State, "controller.db"))
	if err != nil {
		logger.Error("state database initialization failed", "path", layout.State, "error", err)
		os.Exit(1)
	}
	defer store.Close()

	statusService := service.NewStatusService("dev")
	installationService := service.NewInstallationService(layout, store)
	if _, err := installationService.Initialize(ctx); err != nil {
		logger.Error("installation state initialization failed", "root", layout.Root, "error", err)
		os.Exit(1)
	}
	clusterService := service.NewClusterConfigService(store, dstconfig.NewWriter(layout))
	if _, err := clusterService.Initialize(ctx); err != nil {
		logger.Error("cluster config initialization failed", "root", layout.Root, "error", err)
		os.Exit(1)
	}
	taskService := service.NewInstallTaskService(store, taskid.Generator{})
	installRunnerService := service.NewInstallRunnerService(
		layout,
		store,
		store,
		service.NewInstallPlanner(),
		taskService,
		steamcmd.NewClient(command.ExecRunner{}),
	)
	runtimeService := service.NewRuntimeService(
		layout,
		store,
		store,
		store,
		dstserver.NewClient(command.ExecRunner{}),
	)
	installTaskLogService := service.NewInstallTaskLogService(layout, logtail.Reader{})
	runtimeLogService := service.NewRuntimeLogService(layout, logtail.Reader{})
	runtimeHistoryService := service.NewRuntimeHistoryService(store)

	server := &http.Server{
		Addr: "127.0.0.1:8737",
		Handler: apphttp.NewRouter(logger, apphttp.Services{
			Status:          statusService,
			Installation:    installationService,
			Cluster:         clusterService,
			InstallTasks:    installRunnerService,
			InstallTaskLogs: installTaskLogService,
			Runtime:         runtimeService,
			RuntimeLogs:     runtimeLogService,
			RuntimeHistory:  runtimeHistoryService,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting dst-server-ctl", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
		os.Exit(1)
	}
}
