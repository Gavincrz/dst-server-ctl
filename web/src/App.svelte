<script lang="ts">
  import { onMount } from 'svelte';

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

  let controller: ControllerStatus | null = null;
  let installation: InstallationStatus | null = null;
  let installTasks: InstallationTask[] = [];
  let loading = true;
  let polling = false;
  let installSubmitting = false;
  let refreshError = '';
  let installError = '';
  let actionMessage = '';
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

  async function refresh(options: { background?: boolean } = {}) {
    const background = options.background ?? false;

    if (background) {
      polling = true;
    } else {
      loading = true;
    }
    refreshError = '';

    try {
      const [controllerStatus, installationStatus, tasks] = await Promise.all([
        fetchJSON<ControllerStatus>('/api/v1/status'),
        fetchJSON<InstallationStatus>('/api/v1/installation'),
        fetchJSON<InstallationTask[]>('/api/v1/install/tasks')
      ]);
      controller = controllerStatus;
      installation = installationStatus;
      installTasks = tasks;
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

  function canStartInstall(status: InstallationStatus | null, tasks: InstallationTask[]) {
    if (!status) {
      return false;
    }
    if (status.steamcmdInstalled && status.dstInstalled) {
      return false;
    }
    return !hasActiveInstallTasks(tasks);
  }

  function taskTypeLabel(type: string) {
    switch (type) {
      case 'install_steamcmd':
        return 'Install SteamCMD';
      case 'install_dst':
        return 'Install DST';
      default:
        return type;
    }
  }

  function taskStatusLabel(status: string) {
    return status.split('_').join(' ');
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
      actionMessage = 'Install tasks created. Progress will refresh automatically.';
      await refresh({ background: true });
    } catch (err) {
      installError = err instanceof Error ? err.message : 'Install request failed';
    } finally {
      installSubmitting = false;
    }
  }

  onMount(() => {
    let pollTimer: ReturnType<typeof setInterval> | null = null;

    void refreshNow();
    pollTimer = setInterval(() => {
      if (!hasActiveInstallTasks(installTasks) || installSubmitting) {
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
      <span>Install Tasks</span>
      <strong>{actionMessage}</strong>
    </section>
  {/if}

  {#if installError}
    <section class="notice" aria-live="polite">
      <span>Install Request Failed</span>
      <strong>{installError}</strong>
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
      <small>version {controller?.version ?? '-'}</small>
    </div>
    <div class="metric">
      <span>Installation</span>
      <strong>{overallInstallState(installation)}</strong>
      <small>managed instance</small>
    </div>
    <div class="metric">
      <span>Last Refresh</span>
      <strong>{refreshedAt ? refreshedAt.toLocaleTimeString() : '-'}</strong>
      <small>{loading ? 'loading' : polling ? 'polling install progress' : 'current snapshot'}</small>
    </div>
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

    <section class="panel panel-wide" aria-label="Installation tasks">
      <div class="panel-heading">
        <div>
          <h2>Installation Tasks</h2>
          <p class="subtle">Create managed install tasks and inspect recent execution state.</p>
        </div>
        <button
          type="button"
          class="install-action"
          disabled={!canStartInstall(installation, installTasks) || installSubmitting}
          on:click={startInstall}
        >
          {installSubmitting ? 'Starting Install' : 'Start Install'}
        </button>
      </div>

      <div class="task-summary">
        <span class="badge">{installTasks.length} task{installTasks.length === 1 ? '' : 's'}</span>
        {#if installation && installation.steamcmdInstalled && installation.dstInstalled}
          <span class="task-summary-text">Managed installation is already ready.</span>
        {:else if hasActiveInstallTasks(installTasks)}
          <span class="task-summary-text">
            An installation run is already queued or in progress. Status refreshes every 3 seconds.
          </span>
        {:else}
          <span class="task-summary-text">No active install run. Use the button to create a new run.</span>
        {/if}
      </div>

      {#if installTasks.length === 0}
        <div class="empty-state">
          <strong>No installation tasks yet</strong>
          <p>Start an install run to create SteamCMD and DST tasks in the managed root.</p>
        </div>
      {:else}
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
      {/if}
    </section>
  </section>
</main>
