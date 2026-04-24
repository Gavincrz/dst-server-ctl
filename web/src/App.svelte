<script lang="ts">
  import { onMount } from 'svelte';

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

  let controller: ControllerStatus | null = null;
  let installation: InstallationStatus | null = null;
  let loading = true;
  let error = '';
  let refreshedAt: Date | null = null;

  async function fetchJSON<T>(path: string): Promise<T> {
    const response = await fetch(path);
    if (!response.ok) {
      throw new Error(`${path} returned HTTP ${response.status}`);
    }
    return response.json() as Promise<T>;
  }

  async function refresh() {
    loading = true;
    error = '';

    try {
      const [controllerStatus, installationStatus] = await Promise.all([
        fetchJSON<ControllerStatus>('/api/v1/status'),
        fetchJSON<InstallationStatus>('/api/v1/installation')
      ]);
      controller = controllerStatus;
      installation = installationStatus;
      refreshedAt = new Date();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Request failed';
    } finally {
      loading = false;
    }
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

  onMount(() => {
    void refresh();
  });
</script>

<main>
  <header class="topbar">
    <div>
      <p class="eyebrow">dst-server-ctl</p>
      <h1>Server Status</h1>
    </div>
    <button type="button" class="refresh" disabled={loading} on:click={refresh}>
      {loading ? 'Refreshing' : 'Refresh'}
    </button>
  </header>

  {#if error}
    <section class="notice" aria-live="polite">
      <span>Backend request failed</span>
      <strong>{error}</strong>
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
      <small>{loading ? 'loading' : 'current snapshot'}</small>
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
  </section>
</main>
