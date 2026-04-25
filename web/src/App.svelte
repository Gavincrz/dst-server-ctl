<script lang="ts">
  import { onMount } from 'svelte';

  import {
    clusterFormFromConfig,
    clusterFormIsDirty,
    clusterRequestFromForm,
    type ClusterConfig,
    type ClusterFormState
  } from './lib/clusterForm';

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

  let controller: ControllerStatus | null = null;
  let installation: InstallationStatus | null = null;
  let cluster: ClusterConfig | null = null;
  let clusterForm: ClusterFormState | null = null;
  let runtime: RuntimeStatus | null = null;
  let runtimeLogs: Record<string, string[]> = {
    Master: [],
    Caves: []
  };
  let installTasks: InstallationTask[] = [];
  let loading = true;
  let polling = false;
  let installSubmitting = false;
  let clusterSubmitting = false;
  let runtimeSubmitting = false;
  let refreshError = '';
  let installError = '';
  let clusterError = '';
  let runtimeError = '';
  let actionMessage = '';
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

  function latestTask(tasks: InstallationTask[]) {
    return tasks[0] ?? null;
  }

  function activeTask(tasks: InstallationTask[]) {
    return tasks.find((task) => task.status === 'running') ?? tasks.find((task) => task.status === 'pending') ?? null;
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
      const [controllerStatus, installationStatus, clusterConfig, runtimeStatus, tasks, masterLogs, cavesLogs] = await Promise.all([
        fetchJSON<ControllerStatus>('/api/v1/status'),
        fetchJSON<InstallationStatus>('/api/v1/installation'),
        fetchJSON<ClusterConfig>('/api/v1/cluster'),
        fetchJSON<RuntimeStatus>('/api/v1/runtime'),
        fetchJSON<InstallationTask[]>('/api/v1/install/tasks'),
        fetchJSON<RuntimeLogResponse>('/api/v1/runtime/logs?shard=Master&lines=120'),
        fetchJSON<RuntimeLogResponse>('/api/v1/runtime/logs?shard=Caves&lines=120')
      ]);

      const previousCluster = cluster;
      controller = controllerStatus;
      installation = installationStatus;
      cluster = clusterConfig;
      runtime = runtimeStatus;
      runtimeLogs = {
        Master: masterLogs.lines,
        Caves: cavesLogs.lines
      };
      installTasks = tasks;
      if (!clusterForm || !previousCluster || !clusterFormIsDirty(clusterForm, previousCluster)) {
        clusterForm = clusterFormFromConfig(clusterConfig);
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
      if ((!hasActiveInstallTasks(installTasks) && !runtimeRunning(runtime)) || installSubmitting || runtimeSubmitting) {
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
          <span class={`badge badge-${runtimeLogs.Master.length > 0 ? 'succeeded' : 'idle'}`}>
            {runtimeLogs.Master.length > 0 ? `${runtimeLogs.Master.length} lines` : 'No output'}
          </span>
        </div>
        <pre class="log-output">{runtimeLogs.Master.length > 0 ? runtimeLogs.Master.join('\n') : 'No Master log output yet.'}</pre>
      </article>

      <article class="log-card">
        <div class="log-card-head">
          <strong>Caves</strong>
          <span class={`badge badge-${runtimeLogs.Caves.length > 0 ? 'succeeded' : 'idle'}`}>
            {runtimeLogs.Caves.length > 0 ? `${runtimeLogs.Caves.length} lines` : 'No output'}
          </span>
        </div>
        <pre class="log-output">{runtimeLogs.Caves.length > 0 ? runtimeLogs.Caves.join('\n') : 'No Caves log output yet.'}</pre>
      </article>
    </div>
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
              {#if task.error}
                <p class="task-error">{task.error}</p>
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
