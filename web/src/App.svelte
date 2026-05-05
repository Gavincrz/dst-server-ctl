<script lang="ts">
  import { onMount } from 'svelte';

  import {
    clusterFormFromConfig,
    clusterFormIsDirty,
    clusterRequestFromForm,
    type ClusterConfig,
    type ClusterFormState
  } from './lib/clusterForm';
  import {
    activeExpandedLogIDs,
    activeExpandedTaskIDs,
    taskLogButtonLabel,
    type KeyedLogCollectionState,
    type SingleLogPanelState
  } from './lib/taskLogs';

  const installPollIntervalMs = 3000;

  type ControllerStatus = {
    version: string;
    status: string;
    startedAt?: string;
  };

  type InstallationStatus = {
    managedRoot: string;
    steamcmdInstalled: boolean;
    dstInstalled: boolean;
    createdAt: string;
    updatedAt: string;
  };

  type InstallationTask = {
    id: string;
    type: string;
    status: string;
    detail: string;
    error?: string;
    startedAt?: string;
    finishedAt?: string;
    createdAt: string;
    updatedAt: string;
  };

  type UpdateStatus = {
    currentVersion: string;
    latestVersion: string;
    updateAvailable: boolean;
    lastCheckedAt?: string;
    lastUpdatedAt?: string;
    lastError?: string;
    createdAt: string;
    updatedAt: string;
  };

  type RuntimeShardStatus = {
    name: string;
    running: boolean;
    pid?: number;
  };

  type RuntimeStatus = {
    status: string;
    shards: RuntimeShardStatus[];
    restartRequired: boolean;
    lastError?: string;
  };

  type RuntimeLogResponse = {
    shard: string;
    lines: string[];
  };

  type TaskLogResponse = {
    taskId: string;
    lines: string[];
  };

  type RuntimeHistoryEvent = {
    id: number;
    shard: string;
    kind: string;
    detail: string;
    createdAt: string;
  };

  type TaskLogKind = 'install' | 'update';
  type RuntimeShardLogKind = 'Master' | 'Caves';

  let controller: ControllerStatus | null = null;
  let installation: InstallationStatus | null = null;
  let updates: UpdateStatus | null = null;
  let cluster: ClusterConfig | null = null;
  let clusterForm: ClusterFormState | null = null;
  let runtime: RuntimeStatus | null = null;
  let runtimeLogState: KeyedLogCollectionState = {
    logs: {},
    loading: {},
    errors: {},
    expanded: {}
  };
  let runtimeHistory: RuntimeHistoryEvent[] = [];
  let installTasks: InstallationTask[] = [];
  let updateTasks: InstallationTask[] = [];
  let updateCheckLogState: SingleLogPanelState = {
    lines: [],
    expanded: false,
    loading: false,
    error: ''
  };
  let installTaskLogState: KeyedLogCollectionState = {
    logs: {},
    loading: {},
    errors: {},
    expanded: {}
  };
  let updateTaskLogState: KeyedLogCollectionState = {
    logs: {},
    loading: {},
    errors: {},
    expanded: {}
  };
  let loading = true;
  let polling = false;
  let installSubmitting = false;
  let updateSubmitting = false;
  let updateCheckSubmitting = false;
  let clusterSubmitting = false;
  let runtimeSubmitting = false;
  let refreshError = '';
  let installError = '';
  let updateError = '';
  let clusterError = '';
  let runtimeError = '';
  let actionMessage = '';
  let updateMessage = '';
  let clusterMessage = '';
  let runtimeMessage = '';
  let refreshedAt: Date | null = null;

  async function fetchJSON<T>(path: string): Promise<T> {
    const response = await fetch(path);
    if (!response.ok) {
      let message = `${path} returned HTTP ${response.status}`;
      try {
        const payload = (await response.json()) as { error?: string };
        if (payload.error) {
          message = payload.error;
        }
      } catch {
        // Ignore JSON parse failures and keep the status-based message.
      }
      throw new Error(message);
    }
    return response.json() as Promise<T>;
  }

  function hasActiveInstallTasks(tasks: InstallationTask[]) {
    return tasks.some((task) => task.status === 'pending' || task.status === 'running');
  }

  function hasActiveUpdateTasks(tasks: InstallationTask[]) {
    return tasks.some((task) => task.status === 'pending' || task.status === 'running');
  }

  function latestTask(tasks: InstallationTask[]) {
    return tasks[0] ?? null;
  }

  function activeTask(tasks: InstallationTask[]) {
    return tasks.find((task) => task.status === 'running') ?? tasks.find((task) => task.status === 'pending') ?? null;
  }

  function taskLogState(kind: TaskLogKind) {
    return kind === 'install' ? installTaskLogState : updateTaskLogState;
  }

  function setTaskLogState(kind: TaskLogKind, state: KeyedLogCollectionState) {
    if (kind === 'install') {
      installTaskLogState = state;
      return;
    }
    updateTaskLogState = state;
  }

  function taskLogEndpoint(kind: TaskLogKind, taskID: string) {
    return kind === 'install'
      ? `/api/v1/install/tasks/${taskID}/logs?lines=160`
      : `/api/v1/update/tasks/${taskID}/logs?lines=160`;
  }

  function taskLogUnavailableMessage(kind: TaskLogKind) {
    return kind === 'install' ? 'Install task logs unavailable' : 'Update task logs unavailable';
  }

  function runtimeLogEndpoint(shard: RuntimeShardLogKind) {
    return `/api/v1/runtime/logs?shard=${shard}&lines=120`;
  }

  async function refresh(options: { background?: boolean } = {}) {
    const background = options.background ?? false;

    if (background) {
      polling = true;
    } else {
      loading = true;
    }
    refreshError = '';

    try {
      const [controllerStatus, installationStatus, updateStatus, clusterConfig, runtimeStatus, installTaskStatus, updateTaskStatus, history] = await Promise.all([
        fetchJSON<ControllerStatus>('/api/v1/status'),
        fetchJSON<InstallationStatus>('/api/v1/installation'),
        fetchJSON<UpdateStatus>('/api/v1/update'),
        fetchJSON<ClusterConfig>('/api/v1/cluster'),
        fetchJSON<RuntimeStatus>('/api/v1/runtime'),
        fetchJSON<InstallationTask[]>('/api/v1/install/tasks'),
        fetchJSON<InstallationTask[]>('/api/v1/update/tasks'),
        fetchJSON<RuntimeHistoryEvent[]>('/api/v1/runtime/history?limit=12')
      ]);

      const previousCluster = cluster;
      controller = controllerStatus;
      installation = installationStatus;
      updates = updateStatus;
      cluster = clusterConfig;
      runtime = runtimeStatus;
      runtimeHistory = history;
      installTasks = installTaskStatus;
      updateTasks = updateTaskStatus;
      if (!clusterForm || !previousCluster || !clusterFormIsDirty(clusterForm, previousCluster)) {
        clusterForm = clusterFormFromConfig(clusterConfig);
      }

      if (activeExpandedTaskIDs(installTaskLogState.expanded, installTaskStatus).length > 0) {
        await refreshTaskLogs('install', installTaskStatus);
      }
      if (activeExpandedTaskIDs(updateTaskLogState.expanded, updateTaskStatus).length > 0) {
        await refreshTaskLogs('update', updateTaskStatus);
      }
      if (updateCheckLogState.expanded) {
        await refreshUpdateCheckLogs();
      }
      if (activeExpandedLogIDs(runtimeLogState.expanded, [{ id: 'Master' }, { id: 'Caves' }], () => runtimeStatus.status === 'running').length > 0) {
        await refreshRuntimeLogs(runtimeStatus);
      }

      refreshedAt = new Date();
    } catch (err) {
      refreshError = err instanceof Error ? err.message : 'Request failed';
    } finally {
      if (background) {
        polling = false;
      } else {
        loading = false;
      }
    }
  }

  function refreshNow() {
    void refresh();
  }

  function formatDate(value?: string | null) {
    if (!value) {
      return 'Not recorded';
    }
    return new Intl.DateTimeFormat(undefined, {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    }).format(new Date(value));
  }

  function boolLabel(value: boolean) {
    return value ? 'Ready' : 'Not installed';
  }

  function overallInstallState(status: InstallationStatus | null) {
    if (!status) {
      return 'Loading';
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return 'Ready';
    }
    if (status.steamcmdInstalled || status.dstInstalled) {
      return 'Partial';
    }
    return 'Not installed';
  }

  function clusterShardEnabled(name: 'Master' | 'Caves') {
    if (!clusterForm) {
      return false;
    }
    return name === 'Master' ? clusterForm.masterEnabled : clusterForm.cavesEnabled;
  }

  function clusterDirty() {
    if (!cluster || !clusterForm) {
      return false;
    }
    return clusterFormIsDirty(clusterForm, cluster);
  }

  function canStartInstall(status: InstallationStatus | null, tasks: InstallationTask[]) {
    if (!status) {
      return false;
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return false;
    }
    return !hasActiveInstallTasks(tasks);
  }

  function runtimeRunning(status: RuntimeStatus | null) {
    return status?.status === 'running';
  }

  function updateVersionLabel(value?: string) {
    return value && value.length > 0 ? value : 'Unknown';
  }

  function runtimeActionLabel(status: RuntimeStatus | null) {
    if (runtimeSubmitting) {
      return runtimeRunning(status) ? 'Start Server' : 'Starting';
    }
    return 'Start Server';
  }

  function canRestartRuntime(status: RuntimeStatus | null, install: InstallationStatus | null) {
    return !!install?.dstInstalled && !runtimeSubmitting;
  }

  function canStartRuntime(status: RuntimeStatus | null, install: InstallationStatus | null) {
    if (!install?.dstInstalled) {
      return false;
    }
    return !runtimeRunning(status);
  }

  function canStopRuntime(status: RuntimeStatus | null) {
    return runtimeRunning(status);
  }

  function taskTypeLabel(type: string) {
    switch (type) {
      case 'install_steamcmd':
        return 'SteamCMD';
      case 'install_dst':
        return 'DST Dedicated Server';
      case 'check_dst_update':
        return 'Version Check';
      case 'update_dst':
        return 'DST Update';
      default:
        return type;
    }
  }

  function taskStatusLabel(status: string) {
    switch (status) {
      case 'pending':
        return 'Waiting';
      case 'running':
        return 'Installing';
      case 'succeeded':
        return 'Done';
      case 'failed':
        return 'Failed';
      case 'idle':
        return 'Idle';
      default:
        return status.split('_').join(' ');
    }
  }

  function runtimeEventLabel(kind: string) {
    switch (kind) {
      case 'started':
        return 'Started';
      case 'stopped':
        return 'Stopped';
      case 'exited':
        return 'Exited';
      case 'retried':
        return 'Retried';
      default:
        return kind;
    }
  }

  function initializationHeading(status: InstallationStatus | null, tasks: InstallationTask[]) {
    if (!status) {
      return 'Loading initialization state';
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return 'Managed server is ready';
    }

    const currentTask = activeTask(tasks);
    if (currentTask?.status === 'running') {
      return `Installing ${taskTypeLabel(currentTask.type)}`;
    }
    if (currentTask?.status === 'pending') {
      return `Waiting to start ${taskTypeLabel(currentTask.type)}`;
    }

    const recentTask = latestTask(tasks);
    if (recentTask?.status === 'failed') {
      return 'Installation needs attention';
    }

    return 'Managed server is not installed yet';
  }

  function initializationMessage(status: InstallationStatus | null, tasks: InstallationTask[]) {
    if (!status) {
      return 'Loading installation details from the controller.';
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return 'SteamCMD and the DST dedicated server are installed inside the managed root.';
    }

    const currentTask = activeTask(tasks);
    if (currentTask?.status === 'running') {
      return `${taskTypeLabel(currentTask.type)} is currently running. Progress refreshes automatically every 3 seconds.`;
    }
    if (currentTask?.status === 'pending') {
      return `${taskTypeLabel(currentTask.type)} is waiting for the current install step to finish.`;
    }

    const recentTask = latestTask(tasks);
    if (recentTask?.status === 'failed') {
      return recentTask.error || 'The last install attempt failed. Review the error and retry when ready.';
    }

    return 'Start installation to prepare SteamCMD and the DST dedicated server in the managed root.';
  }

  function installActionLabel(status: InstallationStatus | null, tasks: InstallationTask[]) {
    if (installSubmitting) {
      return 'Starting Install';
    }
    if (!status) {
      return 'Start Install';
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return 'Install Complete';
    }

    const recentTask = latestTask(tasks);
    if (recentTask?.status === 'failed') {
      return 'Retry Install';
    }

    return 'Start Install';
  }

  function updateHeading(status: UpdateStatus | null) {
    if (!status) {
      return 'Loading update status';
    }
    if (status.updateAvailable) {
      return 'Update available for the managed DST install';
    }
    if (status.currentVersion && status.latestVersion && status.currentVersion === status.latestVersion) {
      return 'Managed DST install is up to date';
    }
    return 'Version status has not been checked yet';
  }

  function updateMessageBody(status: UpdateStatus | null) {
    if (!status) {
      return 'Loading update details from the controller.';
    }
    if (status.lastError) {
      return status.lastError;
    }
    if (status.updateAvailable) {
      return `Local build ${updateVersionLabel(status.currentVersion)} is behind remote build ${updateVersionLabel(status.latestVersion)}.`;
    }
    if (status.currentVersion && status.latestVersion && status.currentVersion === status.latestVersion) {
      return `Local build ${status.currentVersion} matches the latest remote build.`;
    }
    return 'Run a version check to compare the managed install against the latest DST dedicated server build.';
  }

  function canCheckUpdates(install: InstallationStatus | null, tasks: InstallationTask[]) {
    return !!install?.dstInstalled && !hasActiveUpdateTasks(tasks) && !updateSubmitting && !updateCheckSubmitting;
  }

  function canStartUpdate(status: UpdateStatus | null, install: InstallationStatus | null, tasks: InstallationTask[]) {
    return !!install?.dstInstalled && !!status?.updateAvailable && !hasActiveUpdateTasks(tasks) && !updateSubmitting && !updateCheckSubmitting;
  }

  function updateActionLabel(status: UpdateStatus | null, runtimeStatus: RuntimeStatus | null) {
    if (updateSubmitting) {
      return runtimeRunning(runtimeStatus) ? 'Stopping Server and Updating' : 'Starting Update';
    }
    if (runtimeRunning(runtimeStatus) && status?.updateAvailable) {
      return 'Stop Server and Update';
    }
    return 'Run Update';
  }

  function installTaskLogButtonLabel(taskID: string) {
    const state = taskLogState('install');
    return taskLogButtonLabel({
      loading: !!state.loading[taskID],
      expanded: !!state.expanded[taskID],
      hasLogs: taskID in state.logs
    });
  }

  function updateTaskLogButtonLabel(taskID: string) {
    const state = taskLogState('update');
    return taskLogButtonLabel({
      loading: !!state.loading[taskID],
      expanded: !!state.expanded[taskID],
      hasLogs: taskID in state.logs
    });
  }

  function updateCheckLogButtonLabel() {
    return taskLogButtonLabel(
      {
        loading: updateCheckLogState.loading,
        expanded: updateCheckLogState.expanded,
        hasLogs: updateCheckLogState.lines.length > 0
      },
      {
        loading: 'Loading Check Logs',
        hide: 'Hide Check Logs',
        show: 'Show Check Logs',
        view: 'View Check Logs'
      }
    );
  }

  function runtimeLogButtonLabel(shard: RuntimeShardLogKind) {
    return taskLogButtonLabel(
      {
        loading: !!runtimeLogState.loading[shard],
        expanded: !!runtimeLogState.expanded[shard],
        hasLogs: shard in runtimeLogState.logs
      },
      {
        loading: `Loading ${shard} Logs`,
        hide: `Hide ${shard} Logs`,
        show: `Show ${shard} Logs`,
        view: `View ${shard} Logs`
      }
    );
  }

  function resetClusterForm() {
    if (!cluster) {
      return;
    }
    clusterForm = clusterFormFromConfig(cluster);
    clusterError = '';
    clusterMessage = 'Reverted unsaved cluster changes.';
  }

  async function saveClusterConfig() {
    if (!clusterForm) {
      return;
    }

    clusterSubmitting = true;
    clusterError = '';
    clusterMessage = '';

    try {
      const response = await fetch('/api/v1/cluster', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(clusterRequestFromForm(clusterForm))
      });
      if (!response.ok) {
        let message = `Cluster update returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      cluster = (await response.json()) as ClusterConfig;
      clusterForm = clusterFormFromConfig(cluster);
      clusterMessage = 'Cluster configuration saved and regenerated for the managed root.';
      refreshedAt = new Date();
    } catch (err) {
      clusterError = err instanceof Error ? err.message : 'Cluster update failed';
    } finally {
      clusterSubmitting = false;
    }
  }

  async function startInstall() {
    installSubmitting = true;
    installError = '';
    actionMessage = '';

    try {
      const response = await fetch('/api/v1/install/tasks', { method: 'POST' });
      if (!response.ok) {
        let message = `Install request returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      installTasks = (await response.json()) as InstallationTask[];
      actionMessage = 'Installation started. Progress will refresh automatically.';
      await refresh({ background: true });
    } catch (err) {
      installError = err instanceof Error ? err.message : 'Install request failed';
    } finally {
      installSubmitting = false;
    }
  }

  async function checkUpdates() {
    updateCheckSubmitting = true;
    updateError = '';
    updateMessage = '';

    try {
      const response = await fetch('/api/v1/update/check', { method: 'POST' });
      if (!response.ok) {
        let message = `Update check returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      updates = (await response.json()) as UpdateStatus;
      updateMessage = updates.updateAvailable
        ? 'Update check completed. A newer DST build is available.'
        : 'Update check completed. The managed DST install is up to date.';
      refreshedAt = new Date();
      await refresh({ background: true });
    } catch (err) {
      updateError = err instanceof Error ? err.message : 'Update check failed';
    } finally {
      updateCheckSubmitting = false;
    }
  }

  async function startUpdate() {
    updateSubmitting = true;
    updateError = '';
    updateMessage = '';
    const allowStop = runtimeRunning(runtime);

    if (allowStop) {
      const confirmed = window.confirm('The managed DST server is currently running. Start update now and stop all shards first?');
      if (!confirmed) {
        updateSubmitting = false;
        return;
      }
    }

    try {
      const response = await fetch('/api/v1/update/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ allowStop })
      });
      if (!response.ok) {
        let message = `Update request returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      const task = (await response.json()) as InstallationTask;
      updateTasks = [task, ...updateTasks];
      updateMessage = allowStop
        ? 'DST update started after stopping the running server. Progress will refresh automatically.'
        : 'DST update started. Progress will refresh automatically.';
      await refresh({ background: true });
    } catch (err) {
      updateError = err instanceof Error ? err.message : 'Update request failed';
    } finally {
      updateSubmitting = false;
    }
  }

  async function toggleTaskLogs(kind: TaskLogKind, taskID: string) {
    const state = taskLogState(kind);
    if (state.expanded[taskID]) {
      setTaskLogState(kind, {
        ...state,
        expanded: { ...state.expanded, [taskID]: false }
      });
      return;
    }

    setTaskLogState(kind, {
      ...state,
      expanded: { ...state.expanded, [taskID]: true }
    });

    if (state.logs[taskID] || state.loading[taskID]) {
      return;
    }

    await loadTaskLogs(kind, taskID);
  }

  async function loadTaskLogs(kind: TaskLogKind, taskID: string) {
    const state = taskLogState(kind);
    setTaskLogState(kind, {
      ...state,
      loading: { ...state.loading, [taskID]: true },
      errors: { ...state.errors, [taskID]: '' }
    });

    try {
      const payload = await fetchJSON<TaskLogResponse>(taskLogEndpoint(kind, taskID));
      const currentState = taskLogState(kind);
      setTaskLogState(kind, {
        ...currentState,
        logs: { ...currentState.logs, [taskID]: payload.lines }
      });
    } catch (err) {
      const currentState = taskLogState(kind);
      setTaskLogState(kind, {
        ...currentState,
        errors: {
          ...currentState.errors,
          [taskID]: err instanceof Error ? err.message : taskLogUnavailableMessage(kind)
        }
      });
    } finally {
      const currentState = taskLogState(kind);
      setTaskLogState(kind, {
        ...currentState,
        loading: { ...currentState.loading, [taskID]: false }
      });
    }
  }

  async function refreshTaskLogs(kind: TaskLogKind, tasks: InstallationTask[]) {
    const state = taskLogState(kind);
    const taskIDs = activeExpandedTaskIDs(state.expanded, tasks).filter((taskID) => !state.loading[taskID]);
    if (taskIDs.length === 0) {
      return;
    }

    await Promise.all(taskIDs.map(async (taskID) => {
      try {
        const payload = await fetchJSON<TaskLogResponse>(taskLogEndpoint(kind, taskID));
        const currentState = taskLogState(kind);
        setTaskLogState(kind, {
          ...currentState,
          logs: { ...currentState.logs, [taskID]: payload.lines },
          errors: { ...currentState.errors, [taskID]: '' }
        });
      } catch (err) {
        const currentState = taskLogState(kind);
        setTaskLogState(kind, {
          ...currentState,
          errors: {
            ...currentState.errors,
            [taskID]: err instanceof Error ? err.message : taskLogUnavailableMessage(kind)
          }
        });
      }
    }));
  }

  async function toggleUpdateCheckLogs() {
    if (updateCheckLogState.expanded) {
      updateCheckLogState = { ...updateCheckLogState, expanded: false };
      return;
    }

    updateCheckLogState = { ...updateCheckLogState, expanded: true };
    if (updateCheckLogState.lines.length > 0 || updateCheckLogState.loading) {
      return;
    }

    await loadUpdateCheckLogs();
  }

  async function loadUpdateCheckLogs() {
    updateCheckLogState = {
      ...updateCheckLogState,
      loading: true,
      error: ''
    };

    try {
      const payload = await fetchJSON<TaskLogResponse>('/api/v1/update/check/logs?lines=160');
      updateCheckLogState = {
        ...updateCheckLogState,
        lines: payload.lines
      };
    } catch (err) {
      updateCheckLogState = {
        ...updateCheckLogState,
        error: err instanceof Error ? err.message : 'Update check logs unavailable'
      };
    } finally {
      updateCheckLogState = {
        ...updateCheckLogState,
        loading: false
      };
    }
  }

  async function refreshUpdateCheckLogs() {
    try {
      const payload = await fetchJSON<TaskLogResponse>('/api/v1/update/check/logs?lines=160');
      updateCheckLogState = {
        ...updateCheckLogState,
        lines: payload.lines,
        error: ''
      };
    } catch (err) {
      updateCheckLogState = {
        ...updateCheckLogState,
        error: err instanceof Error ? err.message : 'Update check logs unavailable'
      };
    }
  }

  async function toggleRuntimeLogs(shard: RuntimeShardLogKind) {
    if (runtimeLogState.expanded[shard]) {
      runtimeLogState = {
        ...runtimeLogState,
        expanded: { ...runtimeLogState.expanded, [shard]: false }
      };
      return;
    }

    runtimeLogState = {
      ...runtimeLogState,
      expanded: { ...runtimeLogState.expanded, [shard]: true }
    };

    if (runtimeLogState.logs[shard] || runtimeLogState.loading[shard]) {
      return;
    }

    await loadRuntimeLogs(shard);
  }

  async function loadRuntimeLogs(shard: RuntimeShardLogKind) {
    runtimeLogState = {
      ...runtimeLogState,
      loading: { ...runtimeLogState.loading, [shard]: true },
      errors: { ...runtimeLogState.errors, [shard]: '' }
    };

    try {
      const payload = await fetchJSON<RuntimeLogResponse>(runtimeLogEndpoint(shard));
      runtimeLogState = {
        ...runtimeLogState,
        logs: { ...runtimeLogState.logs, [shard]: payload.lines }
      };
    } catch (err) {
      runtimeLogState = {
        ...runtimeLogState,
        errors: {
          ...runtimeLogState.errors,
          [shard]: err instanceof Error ? err.message : 'Runtime logs unavailable'
        }
      };
    } finally {
      runtimeLogState = {
        ...runtimeLogState,
        loading: { ...runtimeLogState.loading, [shard]: false }
      };
    }
  }

  async function refreshRuntimeLogs(runtimeStatus: RuntimeStatus | null) {
    const shardIDs = activeExpandedLogIDs(runtimeLogState.expanded, [{ id: 'Master' }, { id: 'Caves' }], () => runtimeStatus?.status === 'running')
      .filter((shardID) => !runtimeLogState.loading[shardID]);

    if (shardIDs.length === 0) {
      return;
    }

    await Promise.all(shardIDs.map(async (shardID) => {
      try {
        const payload = await fetchJSON<RuntimeLogResponse>(runtimeLogEndpoint(shardID as RuntimeShardLogKind));
        runtimeLogState = {
          ...runtimeLogState,
          logs: { ...runtimeLogState.logs, [shardID]: payload.lines },
          errors: { ...runtimeLogState.errors, [shardID]: '' }
        };
      } catch (err) {
        runtimeLogState = {
          ...runtimeLogState,
          errors: {
            ...runtimeLogState.errors,
            [shardID]: err instanceof Error ? err.message : 'Runtime logs unavailable'
          }
        };
      }
    }));
  }

  async function startRuntime() {
    runtimeSubmitting = true;
    runtimeError = '';
    runtimeMessage = '';

    try {
      const response = await fetch('/api/v1/runtime/start', { method: 'POST' });
      if (!response.ok) {
        let message = `Runtime start returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      runtime = (await response.json()) as RuntimeStatus;
      runtimeMessage = 'Managed shards started from the generated cluster layout.';
      refreshedAt = new Date();
    } catch (err) {
      runtimeError = err instanceof Error ? err.message : 'Runtime start failed';
    } finally {
      runtimeSubmitting = false;
    }
  }

  async function stopRuntime() {
    runtimeSubmitting = true;
    runtimeError = '';
    runtimeMessage = '';

    try {
      const response = await fetch('/api/v1/runtime/stop', { method: 'POST' });
      if (!response.ok) {
        let message = `Runtime stop returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      runtime = (await response.json()) as RuntimeStatus;
      runtimeMessage = 'Managed shard processes were stopped.';
      refreshedAt = new Date();
    } catch (err) {
      runtimeError = err instanceof Error ? err.message : 'Runtime stop failed';
    } finally {
      runtimeSubmitting = false;
    }
  }

  async function restartRuntime() {
    runtimeSubmitting = true;
    runtimeError = '';
    runtimeMessage = '';

    try {
      const response = await fetch('/api/v1/runtime/restart', { method: 'POST' });
      if (!response.ok) {
        let message = `Runtime restart returned HTTP ${response.status}`;
        try {
          const payload = (await response.json()) as { error?: string };
          if (payload.error) {
            message = payload.error;
          }
        } catch {
          // Ignore JSON parse failures and keep the status-based message.
        }
        throw new Error(message);
      }

      runtime = (await response.json()) as RuntimeStatus;
      runtimeMessage = 'Managed shard processes were restarted with the latest cluster configuration.';
      refreshedAt = new Date();
    } catch (err) {
      runtimeError = err instanceof Error ? err.message : 'Runtime restart failed';
    } finally {
      runtimeSubmitting = false;
    }
  }

  onMount(() => {
    let pollTimer: ReturnType<typeof setInterval> | null = null;

    void refreshNow();
    pollTimer = setInterval(() => {
      if (
        (!hasActiveInstallTasks(installTasks) && !hasActiveUpdateTasks(updateTasks) && !runtimeRunning(runtime)) ||
        installSubmitting ||
        updateSubmitting ||
        updateCheckSubmitting ||
        runtimeSubmitting
      ) {
        return;
      }
      void refresh({ background: true });
    }, installPollIntervalMs);

    return () => {
      if (pollTimer) {
        clearInterval(pollTimer);
      }
    };
  });
</script>

<main>
  <header class="topbar">
    <div>
      <p class="eyebrow">dst-server-ctl</p>
      <h1>Server Status</h1>
    </div>
    <button type="button" class="refresh" disabled={loading} on:click={refreshNow}>
      {loading ? 'Refreshing' : 'Refresh'}
    </button>
  </header>

  {#if actionMessage}
    <section class="notice notice-success" aria-live="polite">
      <span>Installation</span>
      <strong>{actionMessage}</strong>
    </section>
  {/if}

  {#if clusterMessage}
    <section class="notice notice-success" aria-live="polite">
      <span>Cluster Configuration</span>
      <strong>{clusterMessage}</strong>
    </section>
  {/if}

  {#if updateMessage}
    <section class="notice notice-success" aria-live="polite">
      <span>Updates</span>
      <strong>{updateMessage}</strong>
    </section>
  {/if}

  {#if runtimeMessage}
    <section class="notice notice-success" aria-live="polite">
      <span>Runtime</span>
      <strong>{runtimeMessage}</strong>
    </section>
  {/if}

  {#if installError}
    <section class="notice" aria-live="polite">
      <span>Install Request Failed</span>
      <strong>{installError}</strong>
    </section>
  {/if}

  {#if clusterError}
    <section class="notice" aria-live="polite">
      <span>Cluster Save Failed</span>
      <strong>{clusterError}</strong>
    </section>
  {/if}

  {#if updateError}
    <section class="notice" aria-live="polite">
      <span>Update Request Failed</span>
      <strong>{updateError}</strong>
    </section>
  {/if}

  {#if runtimeError}
    <section class="notice" aria-live="polite">
      <span>Runtime Request Failed</span>
      <strong>{runtimeError}</strong>
    </section>
  {/if}

  {#if refreshError}
    <section class="notice" aria-live="polite">
      <span>Backend request failed</span>
      <strong>{refreshError}</strong>
    </section>
  {/if}

  <section class="summary" aria-label="Status summary">
    <div class="metric">
      <span>Controller</span>
      <strong>{controller?.status ?? 'Loading'}</strong>
      <small>
        version {controller?.version ?? '-'}{#if controller?.startedAt}
          {' '}· started {formatDate(controller.startedAt)}
        {/if}
      </small>
    </div>
    <div class="metric">
      <span>Runtime</span>
      <strong>{runtime?.status ?? 'Loading'}</strong>
      <small>{runtimeRunning(runtime) ? 'managed shards active' : 'managed shards stopped'}</small>
    </div>
    <div class="metric">
      <span>Last Refresh</span>
      <strong>{refreshedAt ? refreshedAt.toLocaleTimeString() : '-'}</strong>
      <small>{loading ? 'loading' : polling ? 'polling install progress' : 'current snapshot'}</small>
    </div>
  </section>

  <section class="panel panel-wide" aria-label="Updates">
    <div class="panel-heading">
      <div>
        <h2>Updates</h2>
        <p class="subtle">Compare the managed DST install against the latest upstream build and trigger a manual update when needed.</p>
      </div>
        <div class="panel-actions">
          <button type="button" class="secondary-action" disabled={!canCheckUpdates(installation, updateTasks)} on:click={checkUpdates}>
            {updateCheckSubmitting ? 'Checking' : 'Check Now'}
          </button>
          <button type="button" disabled={!canStartUpdate(updates, installation, updateTasks)} on:click={startUpdate}>
            {updateActionLabel(updates, runtime)}
          </button>
        </div>
      </div>

      <div class="runtime-layout">
      <div class="runtime-hero">
        <div class="runtime-copy">
          <span class={`badge badge-${updates?.updateAvailable ? 'failed' : 'succeeded'}`}>
            {updates?.updateAvailable ? 'Update Available' : 'Checked'}
          </span>
          <h3>{updateHeading(updates)}</h3>
          <p>{updateMessageBody(updates)}</p>
          {#if runtimeRunning(runtime)}
            <p>The managed server is currently running. Starting an update will require confirmation and stop both shards first.</p>
          {/if}
          {#if updates?.lastError}
            <p class="task-error">{updates.lastError}</p>
          {/if}
        </div>
        <div class="runtime-meta">
          <article class="checkpoint" class:complete={!!installation?.dstInstalled}>
            <strong>DST Installed</strong>
            <span>{installation ? boolLabel(installation.dstInstalled) : 'Loading'}</span>
          </article>
          <article class="checkpoint" class:complete={!updates?.updateAvailable && !!updates?.latestVersion}>
            <strong>Latest Check</strong>
            <span>{formatDate(updates?.lastCheckedAt)}</span>
          </article>
        </div>
      </div>

      <div class="task-meta">
        <div>
          <dt>Current Build</dt>
          <dd>{updateVersionLabel(updates?.currentVersion)}</dd>
        </div>
        <div>
          <dt>Latest Build</dt>
          <dd>{updateVersionLabel(updates?.latestVersion)}</dd>
        </div>
        <div>
          <dt>Last Checked</dt>
          <dd>{formatDate(updates?.lastCheckedAt)}</dd>
        </div>
        <div>
          <dt>Last Updated</dt>
          <dd>{formatDate(updates?.lastUpdatedAt)}</dd>
        </div>
      </div>

      <div class="task-actions">
        <button type="button" class="secondary-action" disabled={updateCheckLogState.loading} on:click={toggleUpdateCheckLogs}>
          {updateCheckLogButtonLabel()}
        </button>
        {#if updateCheckLogState.expanded}
          <button type="button" class="secondary-action" disabled={updateCheckLogState.loading} on:click={loadUpdateCheckLogs}>
            {updateCheckLogState.loading ? 'Refreshing Check Logs' : 'Refresh Check Logs'}
          </button>
        {/if}
      </div>

      {#if updateCheckLogState.expanded}
        {#if updateCheckLogState.error}
          <p class="task-error">{updateCheckLogState.error}</p>
        {:else}
          <pre class="log-output task-log-output">{updateCheckLogState.lines.length > 0 ? updateCheckLogState.lines.join('\n') : 'No log output recorded for version checks yet.'}</pre>
        {/if}
      {/if}
    </div>

    {#if updateTasks.length > 0}
      <div class="task-list">
        {#each updateTasks as task}
          <article class={`task task-${task.status}`}>
            <div class="task-head">
              <div>
                <strong>{taskTypeLabel(task.type)}</strong>
                <small>{task.detail}</small>
              </div>
              <span class={`badge badge-${task.status}`}>{taskStatusLabel(task.status)}</span>
            </div>
            <dl class="task-meta">
              <div>
                <dt>ID</dt>
                <dd>{task.id}</dd>
              </div>
              <div>
                <dt>Created</dt>
                <dd>{formatDate(task.createdAt)}</dd>
              </div>
              <div>
                <dt>Started</dt>
                <dd>{formatDate(task.startedAt)}</dd>
              </div>
              <div>
                <dt>Finished</dt>
                <dd>{formatDate(task.finishedAt)}</dd>
              </div>
            </dl>
            <div class="task-actions">
              <button type="button" class="secondary-action" disabled={updateTaskLogState.loading[task.id]} on:click={() => toggleTaskLogs('update', task.id)}>
                {updateTaskLogButtonLabel(task.id)}
              </button>
              {#if updateTaskLogState.expanded[task.id]}
                <button type="button" class="secondary-action" disabled={updateTaskLogState.loading[task.id]} on:click={() => loadTaskLogs('update', task.id)}>
                  {updateTaskLogState.loading[task.id] ? 'Refreshing Logs' : 'Refresh Logs'}
                </button>
              {/if}
            </div>
            {#if task.error}
              <p class="task-error">{task.error}</p>
            {/if}
            {#if updateTaskLogState.expanded[task.id]}
              {#if updateTaskLogState.errors[task.id]}
                <p class="task-error">{updateTaskLogState.errors[task.id]}</p>
              {:else}
                <pre class="log-output task-log-output">{updateTaskLogState.logs[task.id]?.length ? updateTaskLogState.logs[task.id].join('\n') : 'No log output recorded for this update task yet.'}</pre>
              {/if}
            {/if}
          </article>
        {/each}
      </div>
    {:else}
      <div class="empty-state">
        <strong>No update tasks yet</strong>
        <p>Use Check Now to compare versions, then start a manual update when a newer build is available.</p>
      </div>
    {/if}
  </section>

  <section class="panel panel-wide" aria-label="Runtime control">
    <div class="panel-heading">
      <div>
        <h2>Runtime Control</h2>
        <p class="subtle">Start or stop the managed Master and Caves shard processes from the controller.</p>
      </div>
      <div class="panel-actions">
        <button type="button" disabled={!canStopRuntime(runtime) || runtimeSubmitting} on:click={stopRuntime}>
          {runtimeSubmitting && runtimeRunning(runtime) ? 'Stopping' : 'Stop Server'}
        </button>
        <button type="button" disabled={!canRestartRuntime(runtime, installation)} on:click={restartRuntime}>
          {runtimeSubmitting ? 'Restarting' : 'Restart Server'}
        </button>
        <button type="button" class="secondary-action" disabled={!canStartRuntime(runtime, installation) || runtimeSubmitting} on:click={startRuntime}>
          {runtimeActionLabel(runtime)}
        </button>
      </div>
    </div>

    <div class="runtime-layout">
      <div class="runtime-hero">
        <div class="runtime-copy">
          <span class={`badge badge-${runtimeRunning(runtime) ? 'succeeded' : 'idle'}`}>
            {runtime?.status ?? 'loading'}
          </span>
          <h3>{runtimeRunning(runtime) ? 'Managed shards are running' : 'Managed shards are stopped'}</h3>
          <p>
            {#if !installation?.dstInstalled}
              Install the DST dedicated server before starting runtime processes.
            {:else if runtimeRunning(runtime)}
              The controller has active shard processes bound to the generated cluster layout under <code>clusters/primary</code>.
            {:else}
              Start the managed server to launch enabled shards from the current cluster configuration.
            {/if}
          </p>
          {#if runtime?.restartRequired}
            <p class="task-error">Cluster configuration changed since the current shard processes were started. Restart is required to apply it.</p>
          {/if}
          {#if runtime?.lastError}
            <p class="task-error">{runtime.lastError}</p>
          {/if}
        </div>
        <div class="runtime-meta">
          <article class="checkpoint" class:complete={runtimeRunning(runtime)}>
            <strong>Runtime State</strong>
            <span>{runtime?.status ?? 'Loading'}</span>
          </article>
          <article class="checkpoint" class:complete={installation?.dstInstalled}>
            <strong>DST Installed</strong>
            <span>{installation ? boolLabel(installation.dstInstalled) : 'Loading'}</span>
          </article>
        </div>
      </div>

      <div class="runtime-shards">
        {#if runtime && runtime.shards.length > 0}
          {#each runtime.shards as shard}
            <article class="shard-runtime-card">
              <div>
                <strong>{shard.name}</strong>
                <small>{shard.running ? `PID ${shard.pid}` : 'Stopped'}</small>
              </div>
              <span class={`badge badge-${shard.running ? 'succeeded' : 'idle'}`}>
                {shard.running ? 'Running' : 'Stopped'}
              </span>
            </article>
          {/each}
        {:else}
          <div class="empty-state">
            <strong>No running shards</strong>
            <p>The controller is not currently supervising Master or Caves processes.</p>
          </div>
        {/if}
      </div>
    </div>
  </section>

  <section class="panel panel-wide" aria-label="Runtime logs">
    <div class="panel-heading">
      <div>
        <h2>Shard Logs</h2>
        <p class="subtle">Recent output from the managed Master and Caves processes under the controller log directory.</p>
      </div>
    </div>

    <div class="log-grid">
      <article class="log-card">
        <div class="log-card-head">
          <strong>Master</strong>
          <span class={`badge badge-${runtimeLogState.logs.Master?.length ? 'succeeded' : 'idle'}`}>
            {runtimeLogState.logs.Master?.length ? `${runtimeLogState.logs.Master.length} lines` : 'No output'}
          </span>
        </div>
        <div class="task-actions">
          <button type="button" class="secondary-action" disabled={runtimeLogState.loading.Master} on:click={() => toggleRuntimeLogs('Master')}>
            {runtimeLogButtonLabel('Master')}
          </button>
          {#if runtimeLogState.expanded.Master}
            <button type="button" class="secondary-action" disabled={runtimeLogState.loading.Master} on:click={() => loadRuntimeLogs('Master')}>
              {runtimeLogState.loading.Master ? 'Refreshing Master Logs' : 'Refresh Master Logs'}
            </button>
          {/if}
        </div>
        {#if runtimeLogState.expanded.Master}
          {#if runtimeLogState.errors.Master}
            <p class="task-error">{runtimeLogState.errors.Master}</p>
          {:else}
            <pre class="log-output">{runtimeLogState.logs.Master?.length ? runtimeLogState.logs.Master.join('\n') : 'No Master log output yet.'}</pre>
          {/if}
        {/if}
      </article>

      <article class="log-card">
        <div class="log-card-head">
          <strong>Caves</strong>
          <span class={`badge badge-${runtimeLogState.logs.Caves?.length ? 'succeeded' : 'idle'}`}>
            {runtimeLogState.logs.Caves?.length ? `${runtimeLogState.logs.Caves.length} lines` : 'No output'}
          </span>
        </div>
        <div class="task-actions">
          <button type="button" class="secondary-action" disabled={runtimeLogState.loading.Caves} on:click={() => toggleRuntimeLogs('Caves')}>
            {runtimeLogButtonLabel('Caves')}
          </button>
          {#if runtimeLogState.expanded.Caves}
            <button type="button" class="secondary-action" disabled={runtimeLogState.loading.Caves} on:click={() => loadRuntimeLogs('Caves')}>
              {runtimeLogState.loading.Caves ? 'Refreshing Caves Logs' : 'Refresh Caves Logs'}
            </button>
          {/if}
        </div>
        {#if runtimeLogState.expanded.Caves}
          {#if runtimeLogState.errors.Caves}
            <p class="task-error">{runtimeLogState.errors.Caves}</p>
          {:else}
            <pre class="log-output">{runtimeLogState.logs.Caves?.length ? runtimeLogState.logs.Caves.join('\n') : 'No Caves log output yet.'}</pre>
          {/if}
        {/if}
      </article>
    </div>
  </section>

  <section class="panel panel-wide" aria-label="Runtime history">
    <div class="panel-heading">
      <div>
        <h2>Runtime History</h2>
        <p class="subtle">Recent persisted runtime events for shard start, stop, exit and auto-retry activity.</p>
      </div>
    </div>

    {#if runtimeHistory.length > 0}
      <div class="history-list">
        {#each runtimeHistory as event}
          <article class="task">
            <div class="task-head">
              <div>
                <strong>{event.shard}</strong>
                <small>{event.detail}</small>
              </div>
              <span class={`badge badge-${event.kind === 'exited' ? 'failed' : event.kind === 'retried' ? 'running' : 'succeeded'}`}>
                {runtimeEventLabel(event.kind)}
              </span>
            </div>
            <small class="history-meta">{formatDate(event.createdAt)}</small>
          </article>
        {/each}
      </div>
    {:else}
      <div class="empty-state">
        <strong>No runtime history yet</strong>
        <p>Shard lifecycle events will appear here after the controller starts supervising processes.</p>
      </div>
    {/if}
  </section>

  <section class="panel panel-wide" aria-label="Cluster configuration">
    <div class="panel-heading">
      <div>
        <h2>Cluster Configuration</h2>
        <p class="subtle">Edit the managed cluster state. Saving writes the structured config back through the controller API.</p>
      </div>
      <div class="panel-actions">
        <button type="button" class="secondary-action" disabled={!clusterDirty() || clusterSubmitting} on:click={resetClusterForm}>
          Reset
        </button>
        <button type="button" disabled={!clusterDirty() || clusterSubmitting || !clusterForm} on:click={saveClusterConfig}>
          {clusterSubmitting ? 'Saving' : 'Save Configuration'}
        </button>
      </div>
    </div>

    {#if cluster && clusterForm}
      <div class="cluster-layout">
        <form class="cluster-form" on:submit|preventDefault={saveClusterConfig}>
          <label class="field">
            <span>Cluster Name</span>
            <input bind:value={clusterForm.clusterName} disabled={clusterSubmitting} maxlength="64" />
          </label>

          <label class="field field-wide">
            <span>Description</span>
            <textarea bind:value={clusterForm.clusterDescription} disabled={clusterSubmitting} rows="3" maxlength="256"></textarea>
          </label>

          <label class="field">
            <span>Game Mode</span>
            <select bind:value={clusterForm.gameMode} disabled={clusterSubmitting}>
              <option value="survival">Survival</option>
              <option value="endless">Endless</option>
              <option value="wilderness">Wilderness</option>
            </select>
          </label>

          <label class="field">
            <span>Language</span>
            <select bind:value={clusterForm.language} disabled={clusterSubmitting}>
              <option value="en">English</option>
              <option value="zh">Chinese</option>
              <option value="zhr">Chinese (Traditional)</option>
              <option value="fr">French</option>
              <option value="de">German</option>
              <option value="it">Italian</option>
              <option value="ja">Japanese</option>
              <option value="ko">Korean</option>
              <option value="pl">Polish</option>
              <option value="pt">Portuguese</option>
              <option value="ru">Russian</option>
              <option value="es">Spanish</option>
            </select>
          </label>

          <label class="field">
            <span>Max Players</span>
            <input bind:value={clusterForm.maxPlayers} disabled={clusterSubmitting} inputmode="numeric" />
          </label>

          <div class="field field-wide">
            <span>World Rules</span>
            <div class="toggle-grid">
              <label class="toggle-card">
                <input bind:checked={clusterForm.pauseWhenEmpty} disabled={clusterSubmitting} type="checkbox" />
                <div>
                  <strong>Pause When Empty</strong>
                  <small>Pause the world simulation when no players are online.</small>
                </div>
              </label>
              <label class="toggle-card">
                <input bind:checked={clusterForm.pvp} disabled={clusterSubmitting} type="checkbox" />
                <div>
                  <strong>Enable PVP</strong>
                  <small>Allow players to damage each other on the managed cluster.</small>
                </div>
              </label>
            </div>
          </div>
        </form>

        <aside class="cluster-sidebar">
          <div class="config-card">
            <span class="badge">{clusterDirty() ? 'Unsaved Changes' : 'Saved'}</span>
            <dl class="details">
              <div>
                <dt>Created</dt>
                <dd>{formatDate(cluster.createdAt)}</dd>
              </div>
              <div>
                <dt>Updated</dt>
                <dd>{formatDate(cluster.updatedAt)}</dd>
              </div>
            </dl>
          </div>

          <div class="config-card">
            <h3>Shard Layout</h3>
            <div class="shard-list">
              <label class="shard-card shard-required">
                <input bind:checked={clusterForm.masterEnabled} disabled={true} type="checkbox" />
                <div>
                  <strong>Master</strong>
                  <small>Required overworld shard for every managed cluster.</small>
                </div>
                <span class={`badge badge-${clusterShardEnabled('Master') ? 'succeeded' : 'failed'}`}>
                  {clusterShardEnabled('Master') ? 'Enabled' : 'Disabled'}
                </span>
              </label>

              <label class="shard-card">
                <input bind:checked={clusterForm.cavesEnabled} disabled={clusterSubmitting} type="checkbox" />
                <div>
                  <strong>Caves</strong>
                  <small>Toggle the secondary shard while keeping the config structure stable.</small>
                </div>
                <span class={`badge badge-${clusterShardEnabled('Caves') ? 'succeeded' : 'idle'}`}>
                  {clusterShardEnabled('Caves') ? 'Enabled' : 'Disabled'}
                </span>
              </label>
            </div>
          </div>
        </aside>
      </div>
    {:else}
      <div class="empty-state">
        <strong>Loading cluster configuration</strong>
        <p>The controller is reading the managed cluster state.</p>
      </div>
    {/if}
  </section>

  <section class="content">
    <section class="panel" aria-label="Managed root">
      <div class="panel-heading">
        <h2>Managed Root</h2>
        <span class="badge">{overallInstallState(installation)}</span>
      </div>
      <dl class="details">
        <div>
          <dt>Path</dt>
          <dd>{installation?.managedRoot ?? '-'}</dd>
        </div>
        <p class="hint">
          Default path is <code>~/.local/share/dst-server-ctl</code> when
          <code>XDG_DATA_HOME</code> is unset.
        </p>
        <div>
          <dt>Initialized</dt>
          <dd>{formatDate(installation?.createdAt)}</dd>
        </div>
        <div>
          <dt>Updated</dt>
          <dd>{formatDate(installation?.updatedAt)}</dd>
        </div>
      </dl>
    </section>

    <section class="panel" aria-label="Installation components">
      <div class="panel-heading">
        <h2>Components</h2>
      </div>
      <div class="component-list">
        <div class:complete={installation?.steamcmdInstalled} class="component">
          <span class="dot" aria-hidden="true"></span>
          <div>
            <strong>SteamCMD</strong>
            <small>{installation ? boolLabel(installation.steamcmdInstalled) : 'Loading'}</small>
          </div>
        </div>
        <div class:complete={installation?.dstInstalled} class="component">
          <span class="dot" aria-hidden="true"></span>
          <div>
            <strong>DST Dedicated Server</strong>
            <small>{installation ? boolLabel(installation.dstInstalled) : 'Loading'}</small>
          </div>
        </div>
      </div>
    </section>

    <section class="panel panel-wide" aria-label="Initialization">
      <div class="panel-heading">
        <div>
          <h2>Initialization</h2>
          <p class="subtle">Use the controller to install and verify the managed DST server instance.</p>
        </div>
        <button
          type="button"
          class="install-action"
          disabled={!canStartInstall(installation, installTasks) || installSubmitting}
          on:click={startInstall}
        >
          {installActionLabel(installation, installTasks)}
        </button>
      </div>

      <div class="init-hero">
        <div class="init-hero-copy">
          <span class="init-kicker">Managed setup</span>
          <h3>{initializationHeading(installation, installTasks)}</h3>
          <p>{initializationMessage(installation, installTasks)}</p>
        </div>
        <div class="init-status">
          <span class={`badge badge-${latestTask(installTasks)?.status ?? 'idle'}`}>
            {installation && installation.steamcmdInstalled && installation.dstInstalled
              ? 'Ready'
              : taskStatusLabel(activeTask(installTasks)?.status ?? latestTask(installTasks)?.status ?? 'idle')}
          </span>
          <small>
            {#if hasActiveInstallTasks(installTasks)}
              Automatic refresh is enabled while installation is active.
            {:else if installation && installation.steamcmdInstalled && installation.dstInstalled}
              The controller can move on to configuration and runtime features.
            {:else}
              Installation starts only when you trigger it from this page.
            {/if}
          </small>
        </div>
      </div>

      <div class="checkpoint-list">
        <article class:complete={installation?.steamcmdInstalled} class="checkpoint">
          <strong>SteamCMD</strong>
          <span>{installation ? boolLabel(installation.steamcmdInstalled) : 'Loading'}</span>
        </article>
        <article class:complete={installation?.dstInstalled} class="checkpoint">
          <strong>DST Dedicated Server</strong>
          <span>{installation ? boolLabel(installation.dstInstalled) : 'Loading'}</span>
        </article>
      </div>

      {#if installTasks.length > 0}
        <div class="task-list">
          {#each installTasks as task}
            <article class={`task task-${task.status}`}>
              <div class="task-head">
                <div>
                  <strong>{taskTypeLabel(task.type)}</strong>
                  <small>{task.detail}</small>
                </div>
                <span class={`badge badge-${task.status}`}>{taskStatusLabel(task.status)}</span>
              </div>
              <dl class="task-meta">
                <div>
                  <dt>ID</dt>
                  <dd>{task.id}</dd>
                </div>
                <div>
                  <dt>Created</dt>
                  <dd>{formatDate(task.createdAt)}</dd>
                </div>
                <div>
                  <dt>Started</dt>
                  <dd>{formatDate(task.startedAt)}</dd>
                </div>
                <div>
                  <dt>Finished</dt>
                  <dd>{formatDate(task.finishedAt)}</dd>
                </div>
              </dl>
              <div class="task-actions">
                <button type="button" class="secondary-action" disabled={installTaskLogState.loading[task.id]} on:click={() => toggleTaskLogs('install', task.id)}>
                  {installTaskLogButtonLabel(task.id)}
                </button>
                {#if installTaskLogState.expanded[task.id]}
                  <button type="button" class="secondary-action" disabled={installTaskLogState.loading[task.id]} on:click={() => loadTaskLogs('install', task.id)}>
                    {installTaskLogState.loading[task.id] ? 'Refreshing Logs' : 'Refresh Logs'}
                  </button>
                {/if}
              </div>
              {#if task.error}
                <p class="task-error">{task.error}</p>
              {/if}
              {#if installTaskLogState.expanded[task.id]}
                {#if installTaskLogState.errors[task.id]}
                  <p class="task-error">{installTaskLogState.errors[task.id]}</p>
                {:else}
                  <pre class="log-output task-log-output">{installTaskLogState.logs[task.id]?.length ? installTaskLogState.logs[task.id].join('\n') : 'No log output recorded for this install task yet.'}</pre>
                {/if}
              {/if}
            </article>
          {/each}
        </div>
      {:else}
        <div class="empty-state">
          <strong>No installation attempts yet</strong>
          <p>The controller is ready to create the managed SteamCMD and DST installation when you start it.</p>
        </div>
      {/if}
    </section>
  </section>
</main>
