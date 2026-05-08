export type ClusterShard = {
  name: string;
  enabled: boolean;
  serverPort: number;
  masterServerPort: number;
  authenticationPort: number;
};

export type ClusterConfig = {
  clusterName: string;
  clusterDescription: string;
  clusterPassword: string;
  clusterIntention: string;
  gameMode: string;
  maxPlayers: number;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  offlineCluster: boolean;
  lanOnlyCluster: boolean;
  tickRate: number;
  consoleEnabled: boolean;
  bindIP: string;
  masterPort: number;
  clusterKey: string;
  shards: ClusterShard[];
  createdAt: string;
  updatedAt: string;
};

export type ClusterUpdateRequest = {
  clusterName: string;
  clusterDescription: string;
  clusterPassword: string;
  clusterIntention: string;
  gameMode: string;
  maxPlayers: number;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  offlineCluster: boolean;
  lanOnlyCluster: boolean;
  tickRate: number;
  consoleEnabled: boolean;
  bindIP: string;
  masterPort: number;
  clusterKey: string;
  shards: ClusterShard[];
};

export type ClusterFormState = {
  clusterName: string;
  clusterDescription: string;
  clusterPassword: string;
  clusterIntention: string;
  gameMode: string;
  maxPlayers: string;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  offlineCluster: boolean;
  lanOnlyCluster: boolean;
  tickRate: string;
  consoleEnabled: boolean;
  bindIP: string;
  masterPort: string;
  clusterKey: string;
  masterEnabled: boolean;
  cavesEnabled: boolean;
  masterServerPort: string;
  masterMasterServerPort: string;
  masterAuthenticationPort: string;
  cavesServerPort: string;
  cavesMasterServerPort: string;
  cavesAuthenticationPort: string;
};

export function clusterFormFromConfig(config: ClusterConfig): ClusterFormState {
  const master = shard(config.shards, 'Master');
  const caves = shard(config.shards, 'Caves');

  return {
    clusterName: config.clusterName,
    clusterDescription: config.clusterDescription,
    clusterPassword: config.clusterPassword,
    clusterIntention: config.clusterIntention,
    gameMode: config.gameMode,
    maxPlayers: String(config.maxPlayers),
    language: config.language,
    pvp: config.pvp,
    pauseWhenEmpty: config.pauseWhenEmpty,
    offlineCluster: config.offlineCluster,
    lanOnlyCluster: config.lanOnlyCluster,
    tickRate: String(config.tickRate),
    consoleEnabled: config.consoleEnabled,
    bindIP: config.bindIP,
    masterPort: String(config.masterPort),
    clusterKey: config.clusterKey,
    masterEnabled: master.enabled,
    cavesEnabled: caves.enabled,
    masterServerPort: String(master.serverPort),
    masterMasterServerPort: String(master.masterServerPort),
    masterAuthenticationPort: String(master.authenticationPort),
    cavesServerPort: String(caves.serverPort),
    cavesMasterServerPort: String(caves.masterServerPort),
    cavesAuthenticationPort: String(caves.authenticationPort)
  };
}

export function clusterRequestFromConfig(config: ClusterConfig): ClusterUpdateRequest {
  const master = shard(config.shards, 'Master');
  const caves = shard(config.shards, 'Caves');

  return {
    clusterName: config.clusterName,
    clusterDescription: config.clusterDescription,
    clusterPassword: config.clusterPassword,
    clusterIntention: config.clusterIntention,
    gameMode: config.gameMode,
    maxPlayers: config.maxPlayers,
    language: config.language,
    pvp: config.pvp,
    pauseWhenEmpty: config.pauseWhenEmpty,
    offlineCluster: config.offlineCluster,
    lanOnlyCluster: config.lanOnlyCluster,
    tickRate: config.tickRate,
    consoleEnabled: config.consoleEnabled,
    bindIP: config.bindIP,
    masterPort: config.masterPort,
    clusterKey: config.clusterKey,
    shards: [
      { ...master, name: 'Master', enabled: master.enabled },
      { ...caves, name: 'Caves', enabled: caves.enabled }
    ]
  };
}

export function clusterRequestFromForm(form: ClusterFormState): ClusterUpdateRequest {
  const parsedMaxPlayers = Number.parseInt(form.maxPlayers.trim(), 10);
  const parsedTickRate = Number.parseInt(form.tickRate.trim(), 10);
  const parsedMasterPort = Number.parseInt(form.masterPort.trim(), 10);

  return {
    clusterName: form.clusterName,
    clusterDescription: form.clusterDescription,
    clusterPassword: form.clusterPassword,
    clusterIntention: form.clusterIntention,
    gameMode: form.gameMode,
    maxPlayers: Number.isNaN(parsedMaxPlayers) ? 0 : parsedMaxPlayers,
    language: form.language,
    pvp: form.pvp,
    pauseWhenEmpty: form.pauseWhenEmpty,
    offlineCluster: form.offlineCluster,
    lanOnlyCluster: form.lanOnlyCluster,
    tickRate: Number.isNaN(parsedTickRate) ? 0 : parsedTickRate,
    consoleEnabled: form.consoleEnabled,
    bindIP: form.bindIP,
    masterPort: Number.isNaN(parsedMasterPort) ? 0 : parsedMasterPort,
    clusterKey: form.clusterKey,
    shards: [
      {
        name: 'Master',
        enabled: form.masterEnabled,
        serverPort: parseNumber(form.masterServerPort),
        masterServerPort: parseNumber(form.masterMasterServerPort),
        authenticationPort: parseNumber(form.masterAuthenticationPort)
      },
      {
        name: 'Caves',
        enabled: form.cavesEnabled,
        serverPort: parseNumber(form.cavesServerPort),
        masterServerPort: parseNumber(form.cavesMasterServerPort),
        authenticationPort: parseNumber(form.cavesAuthenticationPort)
      }
    ]
  };
}

export function clusterFormIsDirty(form: ClusterFormState, config: ClusterConfig): boolean {
  return JSON.stringify(clusterRequestFromForm(form)) !== JSON.stringify(clusterRequestFromConfig(config));
}

function shard(shards: ClusterShard[], name: 'Master' | 'Caves'): ClusterShard {
  const found = shards.find((value) => value.name === name);
  if (found) {
    return found;
  }

  return {
    name,
    enabled: name === 'Master',
    serverPort: 0,
    masterServerPort: 0,
    authenticationPort: 0
  };
}

function parseNumber(value: string): number {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isNaN(parsed) ? 0 : parsed;
}
