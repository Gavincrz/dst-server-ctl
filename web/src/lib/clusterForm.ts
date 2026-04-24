export type ClusterShard = {
  name: string;
  enabled: boolean;
};

export type ClusterConfig = {
  clusterName: string;
  clusterDescription: string;
  gameMode: string;
  maxPlayers: number;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  shards: ClusterShard[];
  createdAt: string;
  updatedAt: string;
};

export type ClusterUpdateRequest = {
  clusterName: string;
  clusterDescription: string;
  gameMode: string;
  maxPlayers: number;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  shards: ClusterShard[];
};

export type ClusterFormState = {
  clusterName: string;
  clusterDescription: string;
  gameMode: string;
  maxPlayers: string;
  language: string;
  pvp: boolean;
  pauseWhenEmpty: boolean;
  masterEnabled: boolean;
  cavesEnabled: boolean;
};

export function clusterFormFromConfig(config: ClusterConfig): ClusterFormState {
  return {
    clusterName: config.clusterName,
    clusterDescription: config.clusterDescription,
    gameMode: config.gameMode,
    maxPlayers: String(config.maxPlayers),
    language: config.language,
    pvp: config.pvp,
    pauseWhenEmpty: config.pauseWhenEmpty,
    masterEnabled: shardEnabled(config.shards, 'Master'),
    cavesEnabled: shardEnabled(config.shards, 'Caves')
  };
}

export function clusterRequestFromConfig(config: ClusterConfig): ClusterUpdateRequest {
  return {
    clusterName: config.clusterName,
    clusterDescription: config.clusterDescription,
    gameMode: config.gameMode,
    maxPlayers: config.maxPlayers,
    language: config.language,
    pvp: config.pvp,
    pauseWhenEmpty: config.pauseWhenEmpty,
    shards: [
      { name: 'Master', enabled: shardEnabled(config.shards, 'Master') },
      { name: 'Caves', enabled: shardEnabled(config.shards, 'Caves') }
    ]
  };
}

export function clusterRequestFromForm(form: ClusterFormState): ClusterUpdateRequest {
  const parsedMaxPlayers = Number.parseInt(form.maxPlayers.trim(), 10);

  return {
    clusterName: form.clusterName,
    clusterDescription: form.clusterDescription,
    gameMode: form.gameMode,
    maxPlayers: Number.isNaN(parsedMaxPlayers) ? 0 : parsedMaxPlayers,
    language: form.language,
    pvp: form.pvp,
    pauseWhenEmpty: form.pauseWhenEmpty,
    shards: [
      { name: 'Master', enabled: form.masterEnabled },
      { name: 'Caves', enabled: form.cavesEnabled }
    ]
  };
}

export function clusterFormIsDirty(form: ClusterFormState, config: ClusterConfig): boolean {
  return JSON.stringify(clusterRequestFromForm(form)) !== JSON.stringify(clusterRequestFromConfig(config));
}

function shardEnabled(shards: ClusterShard[], name: string): boolean {
  return shards.find((shard) => shard.name === name)?.enabled ?? false;
}
