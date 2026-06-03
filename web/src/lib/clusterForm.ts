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
  taskSet: string;
  worldSize: string;
  branching: string;
  loop: string;
  startLocation: string;
  seasonStart: string;
  day: string;
  weather: string;
  lightning: string;
  wildfires: string;
  petrification: string;
  hounds: string;
  autumn: string;
  winter: string;
  spring: string;
  summer: string;
  spawnMode: string;
  ghostEnabled: string;
  ghostSanityDrain: string;
  portalResurrection: string;
  resetTime: string;
  krampus: string;
  beefaloHeat: string;
  roads: string;
  touchstone: string;
  boons: string;
  cavePonds: string;
  cavelight: string;
  prefabSwapsStart: string;
  moonFissure: string;
  terrariumChest: string;
  stagePlays: string;
  junkyard: string;
  earthquakes: string;
  wormAttacks: string;
  winterHounds: string;
  summerHounds: string;
  wormAttacksBoss: string;
  atriumGate: string;
  spawnProtection: string;
  dropEverythingOnDespawn: string;
  healthPenalty: string;
  temperatureDamage: string;
  hunger: string;
  darkness: string;
  specialEvent: string;
  crowCarnival: string;
  hallowedNights: string;
  wintersFeast: string;
  yearOfTheGobbler: string;
  yearOfTheVarg: string;
  yearOfThePig: string;
  yearOfTheCarrat: string;
  yearOfTheBeefalo: string;
  yearOfTheCatcoon: string;
  yearOfTheBunnyman: string;
  yearOfTheDragonfly: string;
  yearOfTheSnake: string;
  yearOfTheKnight: string;
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
  masterWorldSettings: WorldOverride[];
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
  masterWorldSettings: WorldOverride[];
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

type ShardFormName = 'Master' | 'Caves';
type WorldSettingScope = 'shard' | 'master-control';

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

const worldgenFrequencyOptions: WorldSettingOption[] = [
  { value: 'never', label: 'Never' },
  { value: 'rare', label: 'Rare' },
  { value: 'uncommon', label: 'Uncommon' },
  { value: 'default', label: 'Default' },
  { value: 'often', label: 'Often' },
  { value: 'mostly', label: 'Mostly' },
  { value: 'always', label: 'Always' },
  { value: 'insane', label: 'Insane' }
];

const petrificationOptions: WorldSettingOption[] = [
  { value: 'none', label: 'None' },
  { value: 'few', label: 'Few' },
  { value: 'default', label: 'Default' },
  { value: 'many', label: 'Many' },
  { value: 'max', label: 'Max' }
];

const spawnModeOptions: WorldSettingOption[] = [
  { value: 'fixed', label: 'Portal' },
  { value: 'scatter', label: 'Random' }
];

const masterTaskSetOptions: WorldSettingOption[] = [
  { value: 'default', label: 'Default' },
  { value: 'classic', label: 'Classic' }
];

const cavesTaskSetOptions: WorldSettingOption[] = [
  { value: 'cave_default', label: 'Cave Default' }
];

const ghostEnabledOptions: WorldSettingOption[] = [
  { value: 'none', label: 'Respawn' },
  { value: 'always', label: 'Become Ghost' }
];

const resetTimeOptions: WorldSettingOption[] = [
  { value: 'none', label: 'Disabled' },
  { value: 'slow', label: 'Slow' },
  { value: 'default', label: 'Default' },
  { value: 'fast', label: 'Fast' },
  { value: 'always', label: 'Instant' }
];

const emptyWorldSettingOption: WorldSettingOption = { value: '', label: 'Preset Default' };

const yesNoOptions: WorldSettingOption[] = [
  { value: 'never', label: 'Disabled' },
  { value: 'default', label: 'Default' }
];

const nonLethalOptions: WorldSettingOption[] = [
  { value: 'nonlethal', label: 'Non-Lethal' },
  { value: 'default', label: 'Default' }
];

const enabledDisabledOptions: WorldSettingOption[] = [
  { value: 'none', label: 'Disabled' },
  { value: 'always', label: 'Enabled' }
];

const extraEventOptions: WorldSettingOption[] = [
  { value: 'default', label: 'Default' },
  { value: 'enabled', label: 'Enabled' }
];

const eventToggleBindings: WorldSettingField[] = [
  { formKey: 'crowCarnival', overrideKey: 'crow_carnival', label: 'Crow Carnival', description: 'Enable the Crow Carnival event content.', options: extraEventOptions },
  { formKey: 'hallowedNights', overrideKey: 'hallowed_nights', label: 'Hallowed Nights', description: 'Enable the Hallowed Nights event content.', options: extraEventOptions },
  { formKey: 'wintersFeast', overrideKey: 'winters_feast', label: "Winter's Feast", description: "Enable the Winter's Feast event content.", options: extraEventOptions },
  { formKey: 'yearOfTheGobbler', overrideKey: 'year_of_the_gobbler', label: 'Year Of The Gobbler', description: 'Enable the Year of the Gobbler event content.', options: extraEventOptions },
  { formKey: 'yearOfTheVarg', overrideKey: 'year_of_the_varg', label: 'Year Of The Varg', description: 'Enable the Year of the Varg event content.', options: extraEventOptions },
  { formKey: 'yearOfThePig', overrideKey: 'year_of_the_pig', label: 'Year Of The Pig', description: 'Enable the Year of the Pig event content.', options: extraEventOptions },
  { formKey: 'yearOfTheCarrat', overrideKey: 'year_of_the_carrat', label: 'Year Of The Carrat', description: 'Enable the Year of the Carrat event content.', options: extraEventOptions },
  { formKey: 'yearOfTheBeefalo', overrideKey: 'year_of_the_beefalo', label: 'Year Of The Beefalo', description: 'Enable the Year of the Beefalo event content.', options: extraEventOptions },
  { formKey: 'yearOfTheCatcoon', overrideKey: 'year_of_the_catcoon', label: 'Year Of The Catcoon', description: 'Enable the Year of the Catcoon event content.', options: extraEventOptions },
  { formKey: 'yearOfTheBunnyman', overrideKey: 'year_of_the_bunnyman', label: 'Year Of The Bunnyman', description: 'Enable the Year of the Bunnyman event content.', options: extraEventOptions },
  { formKey: 'yearOfTheDragonfly', overrideKey: 'year_of_the_dragonfly', label: 'Year Of The Dragonfly', description: 'Enable the Year of the Dragonfly event content.', options: extraEventOptions },
  { formKey: 'yearOfTheSnake', overrideKey: 'year_of_the_snake', label: 'Year Of The Snake', description: 'Enable the Year of the Snake event content.', options: extraEventOptions },
  { formKey: 'yearOfTheKnight', overrideKey: 'year_of_the_knight', label: 'Year Of The Knight', description: 'Enable the Year of the Knight event content.', options: extraEventOptions }
];

const masterOnlyWorldSettingKeys = new Set<keyof WorldSettingsFormState>([
  'seasonStart',
  'day',
  'weather',
  'lightning',
  'wildfires',
  'petrification',
  'hounds',
  'winterHounds',
  'summerHounds',
  'autumn',
  'winter',
  'spring',
  'summer',
  'moonFissure',
  'terrariumChest',
  'stagePlays',
  'junkyard'
]);

const cavesOnlyWorldSettingKeys = new Set<keyof WorldSettingsFormState>([
  'cavePonds',
  'cavelight',
  'earthquakes',
  'wormAttacks',
  'wormAttacksBoss',
  'atriumGate'
]);

const masterControlledWorldSettingKeys = new Set<keyof WorldSettingsFormState>([
  'spawnMode',
  'ghostEnabled',
  'ghostSanityDrain',
  'portalResurrection',
  'resetTime',
  'beefaloHeat',
  'krampus',
  'roads',
  'touchstone',
  'boons',
  'spawnProtection',
  'dropEverythingOnDespawn',
  'healthPenalty',
  'temperatureDamage',
  'hunger',
  'darkness',
  'specialEvent',
  'crowCarnival',
  'hallowedNights',
  'wintersFeast',
  'yearOfTheGobbler',
  'yearOfTheVarg',
  'yearOfThePig',
  'yearOfTheCarrat',
  'yearOfTheBeefalo',
  'yearOfTheCatcoon',
  'yearOfTheBunnyman',
  'yearOfTheDragonfly',
  'yearOfTheSnake',
  'yearOfTheKnight'
]);

const worldSettingBindings: WorldSettingField[] = [
  {
    formKey: 'taskSet',
    overrideKey: 'task_set',
    label: 'Task Set',
    description: 'Choose the base task layout used to assemble the shard map.',
    options: masterTaskSetOptions
  },
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
    formKey: 'lightning',
    overrideKey: 'lightning',
    label: 'Lightning',
    description: 'Adjust how often lightning strikes happen on the surface.',
    options: frequencyOptions
  },
  {
    formKey: 'wildfires',
    overrideKey: 'wildfires',
    label: 'Wildfires',
    description: 'Control wildfire frequency during summer on the surface.',
    options: frequencyOptions
  },
  {
    formKey: 'petrification',
    overrideKey: 'petrification',
    label: 'Petrification',
    description: 'Tune how aggressively forests petrify into stone trees.',
    options: petrificationOptions
  },
  {
    formKey: 'hounds',
    overrideKey: 'hounds',
    label: 'Hound Attacks',
    description: 'Adjust roaming hound attack frequency on the surface.',
    options: frequencyOptions
  },
  {
    formKey: 'winterHounds',
    overrideKey: 'winterhounds',
    label: 'Winter Hounds',
    description: 'Enable or disable winter hound variants during attacks.',
    options: yesNoOptions
  },
  {
    formKey: 'summerHounds',
    overrideKey: 'summerhounds',
    label: 'Summer Hounds',
    description: 'Enable or disable summer hound variants during attacks.',
    options: yesNoOptions
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
    formKey: 'spawnMode',
    overrideKey: 'spawnmode',
    label: 'Spawn Mode',
    description: 'Choose whether survivors respawn at the portal or scatter randomly.',
    options: spawnModeOptions
  },
  {
    formKey: 'ghostEnabled',
    overrideKey: 'ghostenabled',
    label: 'Ghost Mode',
    description: 'Choose whether dead players become ghosts or respawn directly.',
    options: ghostEnabledOptions
  },
  {
    formKey: 'ghostSanityDrain',
    overrideKey: 'ghostsanitydrain',
    label: 'Ghost Sanity Drain',
    description: 'Control whether ghosts lose sanity over time.',
    options: enabledDisabledOptions
  },
  {
    formKey: 'portalResurrection',
    overrideKey: 'portalresurection',
    label: 'Portal Resurrection',
    description: 'Allow the Florid Postern to revive survivors directly.',
    options: enabledDisabledOptions
  },
  {
    formKey: 'resetTime',
    overrideKey: 'resettime',
    label: 'World Reset Time',
    description: 'Control how quickly an empty world resets itself.',
    options: resetTimeOptions
  },
  {
    formKey: 'beefaloHeat',
    overrideKey: 'beefaloheat',
    label: 'Beefalo Heat',
    description: 'Adjust how often beefalo mating season occurs.',
    options: frequencyOptions
  },
  {
    formKey: 'krampus',
    overrideKey: 'krampus',
    label: 'Krampus',
    description: 'Adjust how often Krampus spawns when naughty actions stack up.',
    options: frequencyOptions
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
    formKey: 'cavelight',
    overrideKey: 'cavelight',
    label: 'Cave Light Flowers',
    description: 'Adjust how quickly cave light flowers regrow and spread.',
    options: [
      { value: 'never', label: 'Never' },
      { value: 'veryslow', label: 'Very Slow' },
      { value: 'slow', label: 'Slow' },
      { value: 'default', label: 'Default' },
      { value: 'fast', label: 'Fast' },
      { value: 'veryfast', label: 'Very Fast' }
    ]
  },
  {
    formKey: 'prefabSwapsStart',
    overrideKey: 'prefabswaps_start',
    label: 'Starting Variety',
    description: 'Change how much the starting biome composition is shuffled.',
    options: [
      { value: 'classic', label: 'Classic' },
      { value: 'default', label: 'Default' },
      { value: 'highly random', label: 'Highly Random' }
    ]
  },
  {
    formKey: 'moonFissure',
    overrideKey: 'moon_fissure',
    label: 'Moon Fissures',
    description: 'Adjust how many moon fissures generate in the forest shard.',
    options: worldgenFrequencyOptions
  },
  {
    formKey: 'terrariumChest',
    overrideKey: 'terrariumchest',
    label: 'Terrarium Chest',
    description: 'Enable or disable the Terrarium chest worldgen feature.',
    options: yesNoOptions
  },
  {
    formKey: 'stagePlays',
    overrideKey: 'stageplays',
    label: 'Stage Plays',
    description: 'Enable or disable set pieces related to stage plays.',
    options: yesNoOptions
  },
  {
    formKey: 'junkyard',
    overrideKey: 'junkyard',
    label: 'Junkyard',
    description: 'Enable or disable the Junkyard worldgen feature.',
    options: yesNoOptions
  },
  {
    formKey: 'earthquakes',
    overrideKey: 'earthquakes',
    label: 'Earthquakes',
    description: 'Adjust cave earthquake frequency.',
    options: frequencyOptions
  },
  {
    formKey: 'wormAttacks',
    overrideKey: 'wormattacks',
    label: 'Worm Attacks',
    description: 'Tune periodic worm raid frequency in the caves.',
    options: frequencyOptions
  },
  {
    formKey: 'wormAttacksBoss',
    overrideKey: 'wormattacks_boss',
    label: 'Depth Worm Waves',
    description: 'Adjust larger cave worm attack waves and boss pressure.',
    options: frequencyOptions
  },
  {
    formKey: 'atriumGate',
    overrideKey: 'atriumgate',
    label: 'Atrium Gate Cooldown',
    description: 'Control how quickly the Ancient Gateway reactivates.',
    options: [
      { value: 'veryslow', label: 'Very Slow' },
      { value: 'slow', label: 'Slow' },
      { value: 'default', label: 'Default' },
      { value: 'fast', label: 'Fast' },
      { value: 'veryfast', label: 'Very Fast' }
    ]
  },
  {
    formKey: 'spawnProtection',
    overrideKey: 'spawnprotection',
    label: 'Spawn Protection',
    description: 'Control whether survivors get spawn-area protection when entering the world.',
    options: [
      { value: 'never', label: 'Disabled' },
      { value: 'default', label: 'Default' },
      { value: 'always', label: 'Always' }
    ]
  },
  {
    formKey: 'dropEverythingOnDespawn',
    overrideKey: 'dropeverythingondespawn',
    label: 'Drop Inventory On Despawn',
    description: 'Force players to drop their inventory when despawning.',
    options: [
      { value: 'default', label: 'Default' },
      { value: 'always', label: 'Always' }
    ]
  },
  {
    formKey: 'healthPenalty',
    overrideKey: 'healthpenalty',
    label: 'Health Penalty',
    description: 'Control whether resurrection health penalties apply.',
    options: [
      { value: 'none', label: 'Disabled' },
      { value: 'always', label: 'Enabled' }
    ]
  },
  {
    formKey: 'temperatureDamage',
    overrideKey: 'temperaturedamage',
    label: 'Temperature Damage',
    description: 'Allow lethal temperature damage or clamp it to non-lethal.',
    options: nonLethalOptions
  },
  {
    formKey: 'hunger',
    overrideKey: 'hunger',
    label: 'Starvation Damage',
    description: 'Allow lethal starvation damage or clamp it to non-lethal.',
    options: nonLethalOptions
  },
  {
    formKey: 'darkness',
    overrideKey: 'darkness',
    label: 'Darkness Damage',
    description: 'Allow lethal darkness damage or clamp it to non-lethal.',
    options: nonLethalOptions
  },
  {
    formKey: 'specialEvent',
    overrideKey: 'specialevent',
    label: 'Special Event',
    description: 'Use the default seasonal live event state or force none.',
    options: [
      { value: 'none', label: 'None' },
      { value: 'default', label: 'Default' }
    ]
  },
  ...eventToggleBindings
];

function worldSettingScope(field: WorldSettingField): WorldSettingScope {
  if (masterControlledWorldSettingKeys.has(field.formKey)) {
    return 'master-control';
  }

  return 'shard';
}

function worldSettingAvailableOnShard(field: WorldSettingField, shardName: ShardFormName): boolean {
  if (masterOnlyWorldSettingKeys.has(field.formKey)) {
    return shardName === 'Master';
  }
  if (cavesOnlyWorldSettingKeys.has(field.formKey)) {
    return shardName === 'Caves';
  }
  if (masterControlledWorldSettingKeys.has(field.formKey)) {
    return shardName === 'Master';
  }

  return true;
}

export const masterWorldGenSettingFields = worldSettingBindings.filter(
  (field) => worldSettingAvailableOnShard(field, 'Master') && worldSettingScope(field) === 'shard'
);

export const masterWorldControlSettingFields = worldSettingBindings.filter(
  (field) => worldSettingAvailableOnShard(field, 'Master') && worldSettingScope(field) === 'master-control'
);

export const cavesWorldSettingFields = worldSettingBindings.filter(
  (field) => worldSettingAvailableOnShard(field, 'Caves')
);

export function clusterFormFromConfig(config: ClusterConfig): ClusterFormState {
  const master = shard(config.shards, 'Master');
  const caves = shard(config.shards, 'Caves');
  const masterWorld = splitWorldSettings(master.worldGenOverrides, config.masterWorldSettings);
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
    masterWorldSettings: [...config.masterWorldSettings].sort((a, b) => a.key.localeCompare(b.key)),
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
    masterWorldSettings: buildWorldOverrides(
      form.masterWorldSettings,
      '',
      (field) => worldSettingScope(field) === 'master-control'
    ),
    shards: [
      {
        name: 'Master',
        enabled: form.masterEnabled,
        serverPort: parseNumber(form.masterServerPort),
        masterServerPort: parseNumber(form.masterMasterServerPort),
        authenticationPort: parseNumber(form.masterAuthenticationPort),
        worldGenPreset: form.masterWorldGenPreset,
        worldGenOverrides: buildWorldOverrides(
          form.masterWorldSettings,
          form.masterExtraWorldGenOverrides,
          (field) => worldSettingScope(field) === 'shard'
        )
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

function splitWorldSettings(
  overrides: WorldOverride[],
  extraSettings: WorldOverride[] = []
): { settings: WorldSettingsFormState; extraOverrides: WorldOverride[] } {
  const settings = emptyWorldSettingsForm();
  const remaining = new Map(overrides.map((override) => [override.key, override.value]));
  const masterControlled = new Map(extraSettings.map((override) => [override.key, override.value]));

  for (const field of worldSettingBindings) {
    const value = remaining.get(field.overrideKey);
    if (value !== undefined) {
      settings[field.formKey] = value;
      remaining.delete(field.overrideKey);
      continue;
    }

    if (worldSettingScope(field) === 'master-control') {
      const masterValue = masterControlled.get(field.overrideKey);
      if (masterValue !== undefined) {
        settings[field.formKey] = masterValue;
      }
    }
  }

  return {
    settings,
    extraOverrides: sortWorldOverrides(
      Array.from(remaining, ([key, value]) => ({ key, value }))
    )
  };
}

function buildWorldOverrides(
  settings: WorldSettingsFormState,
  extraOverrides: string,
  includeField: (field: WorldSettingField) => boolean = () => true
): WorldOverride[] {
  const merged = new Map(parseWorldOverrides(extraOverrides).map((override) => [override.key, override.value]));

  for (const field of worldSettingBindings) {
    if (!includeField(field)) {
      continue;
    }

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
    taskSet: '',
    worldSize: '',
    branching: '',
    loop: '',
    startLocation: '',
    seasonStart: '',
    day: '',
    weather: '',
    lightning: '',
    wildfires: '',
    petrification: '',
    hounds: '',
    autumn: '',
    winter: '',
    spring: '',
    summer: '',
    spawnMode: '',
    ghostEnabled: '',
    ghostSanityDrain: '',
    portalResurrection: '',
    resetTime: '',
    beefaloHeat: '',
    krampus: '',
    roads: '',
    touchstone: '',
    boons: '',
    cavePonds: '',
    cavelight: '',
    prefabSwapsStart: '',
    moonFissure: '',
    terrariumChest: '',
    stagePlays: '',
    junkyard: '',
    earthquakes: '',
    wormAttacks: '',
    winterHounds: '',
    summerHounds: '',
    wormAttacksBoss: '',
    atriumGate: '',
    spawnProtection: '',
    dropEverythingOnDespawn: '',
    healthPenalty: '',
    temperatureDamage: '',
    hunger: '',
    darkness: '',
    specialEvent: '',
    crowCarnival: '',
    hallowedNights: '',
    wintersFeast: '',
    yearOfTheGobbler: '',
    yearOfTheVarg: '',
    yearOfThePig: '',
    yearOfTheCarrat: '',
    yearOfTheBeefalo: '',
    yearOfTheCatcoon: '',
    yearOfTheBunnyman: '',
    yearOfTheDragonfly: '',
    yearOfTheSnake: '',
    yearOfTheKnight: ''
  };
}

function sortWorldOverrides(overrides: WorldOverride[]): WorldOverride[] {
  return [...overrides].sort((a, b) => a.key.localeCompare(b.key));
}

export function worldSettingOptions(field: WorldSettingField, shardName?: ShardFormName): WorldSettingOption[] {
  if (field.formKey === 'taskSet') {
    const taskSetOptions = shardName === 'Caves' ? cavesTaskSetOptions : masterTaskSetOptions;
    return [emptyWorldSettingOption, ...taskSetOptions];
  }

  return [emptyWorldSettingOption, ...field.options];
}
