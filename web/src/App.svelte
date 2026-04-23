<script lang="ts">
  type Status = {
    version: string;
    status: string;
  };

  let status: Status | null = null;
  let error = '';

  async function loadStatus() {
    error = '';
    try {
      const response = await fetch('/api/v1/status');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      status = await response.json();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Unknown error';
    }
  }
</script>

<main>
  <section class="hero">
    <p class="eyebrow">DST Server Controller</p>
    <h1>Clean native server management, without Docker.</h1>
    <p class="lede">
      This UI placeholder will become the local SSH-tunnel friendly control panel for installation,
      worlds, updates, logs, and mods.
    </p>
    <button on:click={loadStatus}>Check backend status</button>
  </section>

  {#if status}
    <section class="card">
      <span>Backend</span>
      <strong>{status.status}</strong>
      <small>version {status.version}</small>
    </section>
  {/if}

  {#if error}
    <section class="card error">
      <span>Backend request failed</span>
      <strong>{error}</strong>
    </section>
  {/if}
</main>

