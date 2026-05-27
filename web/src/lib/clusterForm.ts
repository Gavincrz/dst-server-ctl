export type ClusterShard = {
  name: string;
  enabled: boolean;
  serverPort: number;
  masterServerPort: number;
  authenticationPort: number;
  worldGenPreset: string;
  worldGenOverrides: WorldOverride[];
};

export type WorldOverride = {
  key: string;
  value: string;
};

export type WorldSettingOption = {
  value: string;
  label: string;
};

export type WorldSettingField = {
  formKey: keyof WorldSettingsFormState;
  overrideKey: string;
  label: string;
  description: string;
  options: WorldSettingOption[];
};

export type WorldSettingsFormState = {
  worldSize: string;
  branching: string;
  loop: string;
  startLocation: string;
  seasonStart: string;
  day: string;
  weather: string;
  autumn: string;
  winter: string;
  spring: string;
  summer: string;
  roads: string;
  touchstone: string;
  boons: string;
  cavePonds: string;
  wormAttacks: string;
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
  masterWorldGenPreset: string;
  masterWorldSettings: WorldSettingsFormState;
  masterExtraWorldGenOverrides: string;
  cavesServerPort: string;
  cavesMasterServerPort: string;
  cavesAuthenticationPort: string;
  cavesWorldGenPreset: string;
  cavesWorldSettings: WorldSettingsFormState;
  cavesExtraWorldGenOverrides: string;
};

const frequencyOptions: WorldSettingOption[] = [
  { value: 'never', label: 'Never' },
  { value: 'rare', label: 'Rare' },
  { value: 'default', label: 'Default' },
  { value: 'often', label: 'Often' },
  { value: 'always', label: 'Always' }
];

const seasonLengthOptions: WorldSettingOption[] = [
  { value: 'noseason', label: 'Disabled' },
  { value: 'veryshortseason', label: 'Very Short' },
  { value: 'shortseason', label: 'Short' },
  { value: 'default', label: 'Default' },
  { value: 'longseason', label: 'Long' },
  { value: 'verylongseason', label: 'Very Long' },
  { value: 'random', label: 'Random' }
];

const emptyWorldSettingOption: WorldSettingOption = { value: '', label: 'Preset Default' };

const worldSettingBindings: WorldSettingField[] = [
  {
    formKey: 'worldSize',
    overrideKey: 'world_size',
    label: 'World Size',
    description: 'Control the overall map footprint.',
    options: [
      { value: 'small', label: 'Small' },
      { value: 'medium', label: 'Medium' },
      { value: 'default', label: 'Default' },
      { value: 'huge', label: 'Huge' }
    ]
  },
  {
    formKey: 'branching',
    overrideKey: 'branching',
    label: 'Branches',
    description: 'Bias the map toward straighter or more branchy layouts.',
    options: [
      { value: 'never', label: 'Never' },
      { value: 'least', label: 'Least' },
      { value: 'default', label: 'Default' },
      { value: 'most', label: 'Most' }
    ]
  },
  {
    formKey: 'loop',
    overrideKey: 'loop',
    label: 'Loops',
    description: 'Control whether the map tends to wrap back around itself.',
    options: [
      { value: 'never', label: 'Never' },
      { value: 'default', label: 'Default' },
      { value: 'always', label: 'Always' }
    ]
  },
  {
    formKey: 'startLocation',
    overrideKey: 'start_location',
    label: 'Start Location',
    description: 'Choose the initial spawn style near the Florid Postern.',
    options: [
      { value: 'default', label: 'Default' },
      { value: 'plus', label: 'Extra Supplies' },
      { value: 'darkness', label: 'Darkness' },
      { value: 'caves', label: 'Caves' }
    ]
  },
  {
    formKey: 'seasonStart',
    overrideKey: 'season_start',
    label: 'Starting Season',
    description: 'Pick which season the surface shard starts in.',
    options: [
      { value: 'default', label: 'Default' },
      { value: 'autumn', label: 'Autumn' },
      { value: 'winter', label: 'Winter' },
      { value: 'spring', label: 'Spring' },
      { value: 'summer', label: 'Summer' },
      { value: 'random', label: 'Random' }
    ]
  },
  {
    formKey: 'day',
    overrideKey: 'day',
    label: 'Day Cycle',
    description: 'Adjust the day, dusk, and night balance.',
    options: [
      { value: 'default', label: 'Default' },
      { value: 'longday', label: 'Long Day' },
      { value: 'longdusk', label: 'Long Dusk' },
      { value: 'longnight', label: 'Long Night' },
      { value: 'noday', label: 'No Day' },
      { value: 'nodusk', label: 'No Dusk' },
      { value: 'nonight', label: 'No Night' },
      { value: 'onlyday', label: 'Only Day' },
      { value: 'onlydusk', label: 'Only Dusk' },
      { value: 'onlynight', label: 'Only Night' }
    ]
  },
  {
    formKey: 'weather',
    overrideKey: 'weather',
    label: 'Weather',
    description: 'Tune the frequency of rain and seasonal weather.',
    options: frequencyOptions
  },
  {
    formKey: 'autumn',
    overrideKey: 'autumn',
    label: 'Autumn Length',
    description: 'Override autumn duration on the surface shard.',
    options: seasonLengthOptions
  },
  {
    formKey: 'winter',
    overrideKey: 'winter',
    label: 'Winter Length',
    description: 'Override winter duration on the surface shard.',
    options: seasonLengthOptions
  },
  {
    formKey: 'spring',
    overrideKey: 'spring',
    label: 'Spring Length',
    description: 'Override spring duration on the surface shard.',
    options: seasonLengthOptions
  },
  {
    formKey: 'summer',
    overrideKey: 'summer',
    label: 'Summer Length',
    description: 'Override summer duration on the surface shard.',
    options: seasonLengthOptions
  },
  {
    formKey: 'roads',
    overrideKey: 'roads',
    label: 'Roads',
    description: 'Control how often prebuilt roads appear.',
    options: frequencyOptions
  },
  {
    formKey: 'touchstone',
    overrideKey: 'touchstone',
    label: 'Touch Stones',
    description: 'Adjust the frequency of resurrection touch stones.',
    options: frequencyOptions
  },
  {
    formKey: 'boons',
    overrideKey: 'boons',
    label: 'Boons',
    description: 'Control bonus starter set pieces near the spawn area.',
    options: frequencyOptions
  },
  {
    formKey: 'cavePonds',
    overrideKey: 'cave_ponds',
    label: 'Cave Ponds',
    description: 'Adjust the number of cave fishing ponds.',
    options: frequencyOptions
  },
  {
    formKey: 'wormAttacks',
    overrideKey: 'wormattacks',
    label: 'Worm Attacks',
    description: 'Tune periodic worm raid frequency in the caves.',
    options: frequencyOptions
  }
];

export const masterWorldSettingFields = worldSettingBindings.filter(
  (field) => field.formKey !== 'cavePonds' && field.formKey !== 'wormAttacks'
);

export const cavesWorldSettingFields = worldSettingBindings.filter(
  (field) => field.formKey !== 'seasonStart' &&
    field.formKey !== 'day' &&
    field.formKey !== 'weather' &&
    field.formKey !== 'autumn' &&
    field.formKey !== 'winter' &&
    field.formKey !== 'spring' &&
    field.formKey !== 'summer' &&
    field.formKey !== 'roads' &&
    field.formKey !== 'touchstone' &&
    field.formKey !== 'boons'
);

export function clusterFormFromConfig(config: ClusterConfig): ClusterFormState {
  const master = shard(config.shards, 'Master');
  const caves = shard(config.shards, 'Caves');
  const masterWorld = splitWorldSettings(master.worldGenOverrides);
  const cavesWorld = splitWorldSettings(caves.worldGenOverrides);

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
    masterWorldGenPreset: master.worldGenPreset,
    masterWorldSettings: masterWorld.settings,
    masterExtraWorldGenOverrides: formatWorldOverrides(masterWorld.extraOverrides),
    cavesServerPort: String(caves.serverPort),
    cavesMasterServerPort: String(caves.masterServerPort),
    cavesAuthenticationPort: String(caves.authenticationPort),
    cavesWorldGenPreset: caves.worldGenPreset,
    cavesWorldSettings: cavesWorld.settings,
    cavesExtraWorldGenOverrides: formatWorldOverrides(cavesWorld.extraOverrides)
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
        authenticationPort: parseNumber(form.masterAuthenticationPort),
        worldGenPreset: form.masterWorldGenPreset,
        worldGenOverrides: buildWorldOverrides(form.masterWorldSettings, form.masterExtraWorldGenOverrides)
      },
      {
        name: 'Caves',
        enabled: form.cavesEnabled,
        serverPort: parseNumber(form.cavesServerPort),
        masterServerPort: parseNumber(form.cavesMasterServerPort),
        authenticationPort: parseNumber(form.cavesAuthenticationPort),
        worldGenPreset: form.cavesWorldGenPreset,
        worldGenOverrides: buildWorldOverrides(form.cavesWorldSettings, form.cavesExtraWorldGenOverrides)
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
    authenticationPort: 0,
    worldGenPreset: name === 'Master' ? 'SURVIVAL_TOGETHER' : 'DST_CAVE',
    worldGenOverrides: []
  };
}

function parseNumber(value: string): number {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function formatWorldOverrides(overrides: WorldOverride[]): string {
  return [...overrides]
    .sort((a, b) => a.key.localeCompare(b.key))
    .map((override) => `${override.key}=${override.value}`)
    .join('\n');
}

function parseWorldOverrides(value: string): WorldOverride[] {
  return sortWorldOverrides(
    value
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0)
    .map((line) => {
      const separator = line.indexOf('=');
      if (separator < 0) {
        return { key: line, value: '' };
      }
      return {
        key: line.slice(0, separator).trim(),
        value: line.slice(separator + 1).trim()
      };
    })
  );
}

function splitWorldSettings(overrides: WorldOverride[]): { settings: WorldSettingsFormState; extraOverrides: WorldOverride[] } {
  const settings = emptyWorldSettingsForm();
  const remaining = new Map(overrides.map((override) => [override.key, override.value]));

  for (const field of worldSettingBindings) {
    const value = remaining.get(field.overrideKey);
    if (value !== undefined) {
      settings[field.formKey] = value;
      remaining.delete(field.overrideKey);
    }
  }

  return {
    settings,
    extraOverrides: sortWorldOverrides(
      Array.from(remaining, ([key, value]) => ({ key, value }))
    )
  };
}

function buildWorldOverrides(settings: WorldSettingsFormState, extraOverrides: string): WorldOverride[] {
  const merged = new Map(parseWorldOverrides(extraOverrides).map((override) => [override.key, override.value]));

  for (const field of worldSettingBindings) {
    merged.delete(field.overrideKey);

    const value = settings[field.formKey].trim();
    if (value !== '') {
      merged.set(field.overrideKey, value);
    }
  }

  return sortWorldOverrides(
    Array.from(merged, ([key, value]) => ({ key, value }))
  );
}

function emptyWorldSettingsForm(): WorldSettingsFormState {
  return {
    worldSize: '',
    branching: '',
    loop: '',
    startLocation: '',
    seasonStart: '',
    day: '',
    weather: '',
    autumn: '',
    winter: '',
    spring: '',
    summer: '',
    roads: '',
    touchstone: '',
    boons: '',
    cavePonds: '',
    wormAttacks: ''
  };
}

function sortWorldOverrides(overrides: WorldOverride[]): WorldOverride[] {
  return [...overrides].sort((a, b) => a.key.localeCompare(b.key));
}

export function worldSettingOptions(field: WorldSettingField): WorldSettingOption[] {
  return [emptyWorldSettingOption, ...field.options];
}
