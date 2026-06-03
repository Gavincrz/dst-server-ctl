import { describe, expect, it } from 'vitest';

import {
  cavesWorldSettingFields,
  clusterFormFromConfig,
  clusterFormIsDirty,
  clusterRequestFromForm,
  masterWorldControlSettingFields,
  masterWorldGenSettingFields,
  type ClusterConfig,
  type ClusterFormState
} from './clusterForm';

function sampleConfig(): ClusterConfig {
  return {
    clusterName: 'Managed DST',
    clusterDescription: 'Seasonal world',
    clusterPassword: '',
    clusterIntention: 'cooperative',
    gameMode: 'survival',
    maxPlayers: 6,
    language: 'en',
    pvp: false,
    pauseWhenEmpty: true,
    offlineCluster: false,
    lanOnlyCluster: false,
    tickRate: 15,
    consoleEnabled: true,
    bindIP: '127.0.0.1',
    masterPort: 10888,
    clusterKey: 'dst-server-ctl',
    masterWorldSettings: [],
    shards: [
      {
        name: 'Master',
        enabled: true,
        serverPort: 10999,
        masterServerPort: 27016,
        authenticationPort: 8766,
        worldGenPreset: 'SURVIVAL_TOGETHER',
        worldGenOverrides: [{ key: 'season_start', value: 'autumn' }]
      },
      {
        name: 'Caves',
        enabled: true,
        serverPort: 11000,
        masterServerPort: 27017,
        authenticationPort: 8767,
        worldGenPreset: 'DST_CAVE',
        worldGenOverrides: []
      }
    ],
    createdAt: '2026-04-24T01:00:00Z',
    updatedAt: '2026-04-24T02:00:00Z'
  };
}

describe('clusterForm helpers', () => {
  it('round trips config into a clean form state', () => {
    const config = sampleConfig();
    const form = clusterFormFromConfig(config);

    expect(form.maxPlayers).toBe('6');
    expect(form.tickRate).toBe('15');
    expect(form.masterEnabled).toBe(true);
    expect(form.cavesEnabled).toBe(true);
    expect(form.masterWorldSettings.taskSet).toBe('');
    expect(form.masterWorldSettings.seasonStart).toBe('autumn');
    expect(form.masterWorldSettings.prefabSwapsStart).toBe('');
    expect(form.masterWorldSettings.moonFissure).toBe('');
    expect(form.masterWorldSettings.balatro).toBe('');
    expect(form.masterWorldSettings.hounds).toBe('');
    expect(form.masterWorldSettings.frogRain).toBe('');
    expect(form.masterWorldSettings.riftsFrequency).toBe('');
    expect(form.masterWorldSettings.mutatedHounds).toBe('');
    expect(form.masterWorldSettings.penguinsMoon).toBe('');
    expect(form.masterWorldSettings.moonSpider).toBe('');
    expect(form.masterWorldSettings.mutatedBirds).toBe('');
    expect(form.masterWorldSettings.portalSpawnRate).toBe('');
    expect(form.masterWorldSettings.extraStartingItems).toBe('');
    expect(form.masterWorldSettings.specialEvent).toBe('');
    expect(form.masterWorldSettings.portalResurrection).toBe('');
    expect(form.masterWorldSettings.ghostSanityDrain).toBe('');
    expect(form.masterWorldSettings.beefaloHeat).toBe('');
    expect(form.masterWorldSettings.wintersFeast).toBe('');
    expect(form.masterWorldSettings.terrariumChest).toBe('');
    expect(form.cavesWorldSettings.earthquakes).toBe('');
    expect(form.cavesWorldSettings.wormAttacksBoss).toBe('');
    expect(form.cavesWorldSettings.cavelight).toBe('');
    expect(form.masterExtraWorldGenOverrides).toBe('');
    expect(clusterFormIsDirty(form, config)).toBe(false);
  });

  it('builds a stable update payload from form state', () => {
    const form: ClusterFormState = {
      clusterName: 'Managed DST',
      clusterDescription: 'Seasonal world',
      clusterPassword: 'secret',
      clusterIntention: 'social',
      gameMode: 'endless',
      maxPlayers: ' 8 ',
      language: 'zh',
      pvp: true,
      pauseWhenEmpty: false,
      offlineCluster: true,
      lanOnlyCluster: false,
      tickRate: ' 30 ',
      consoleEnabled: true,
      bindIP: '0.0.0.0',
      masterPort: ' 12000 ',
      clusterKey: 'cluster-abc',
      masterEnabled: true,
      cavesEnabled: false,
      masterServerPort: ' 11000 ',
      masterMasterServerPort: ' 27020 ',
      masterAuthenticationPort: ' 8768 ',
      masterWorldGenPreset: 'SURVIVAL_TOGETHER_CLASSIC',
      masterWorldSettings: {
        taskSet: 'classic',
        worldSize: 'huge',
        branching: 'most',
        loop: 'always',
        startLocation: 'plus',
        seasonStart: 'autumn',
        day: 'longday',
        weather: 'often',
        lightning: 'rare',
        wildfires: 'never',
        petrification: 'many',
        hounds: 'default',
        winterHounds: 'default',
        summerHounds: 'never',
        autumn: 'longseason',
        winter: 'default',
        spring: '',
        summer: '',
        spawnMode: 'scatter',
        ghostEnabled: 'always',
        ghostSanityDrain: 'none',
        portalResurrection: 'always',
        resetTime: 'fast',
        beefaloHeat: 'often',
        krampus: 'rare',
        roads: 'often',
        touchstone: 'rare',
        boons: 'always',
        cavelight: '',
        prefabSwapsStart: 'highly random',
        moonFissure: 'mostly',
        balatro: 'default',
        terrariumChest: 'default',
        stagePlays: 'never',
        junkyard: 'default',
        frogRain: 'often',
        meteorShowers: 'rare',
        hunt: 'default',
        alternateHunt: 'never',
        wanderingTraderEnabled: 'always',
        riftsFrequency: 'often',
        riftsEnabled: 'always',
        lunarHailFrequency: 'rare',
        mutatedHounds: 'default',
        penguinsMoon: 'never',
        moonSpider: 'often',
        mutatedBirds: 'default',
        mutatedMerm: 'never',
        mutatedSpiderQueen: 'default',
        acidRainEnabled: '',
        riftsFrequencyCave: '',
        riftsEnabledCave: '',
        portalSpawnRate: 'often',
        bananaBushPortalRate: 'rare',
        lightCrabPortalRate: 'default',
        monkeytailPortalRate: 'never',
        palmconeSeedPortalRate: 'always',
        powderMonkeyPortalRate: 'default',
        diseaseDelay: 'short',
        basicResourceRegrowth: 'always',
        extraStartingItems: '15',
        seasonalStartingItems: 'default',
        spawnProtection: 'always',
        dropEverythingOnDespawn: 'always',
        healthPenalty: 'none',
        lessDamageTaken: 'always',
        temperatureDamage: 'nonlethal',
        hunger: 'default',
        darkness: 'nonlethal',
        shadowCreatures: 'rare',
        brightmareCreatures: 'often',
        specialEvent: 'none',
        crowCarnival: 'enabled',
        hallowedNights: '',
        wintersFeast: 'enabled',
        yearOfTheGobbler: '',
        yearOfTheVarg: '',
        yearOfThePig: '',
        yearOfTheCarrat: '',
        yearOfTheBeefalo: '',
        yearOfTheCatcoon: '',
        yearOfTheBunnyman: '',
        yearOfTheDragonfly: '',
        yearOfTheSnake: '',
        yearOfTheKnight: '',
        cavePonds: '',
        earthquakes: '',
        wormAttacks: '',
        wormAttacksBoss: '',
        atriumGate: ''
      },
      masterExtraWorldGenOverrides: 'bearger=rare',
      cavesServerPort: ' 11001 ',
      cavesMasterServerPort: ' 27021 ',
      cavesAuthenticationPort: ' 8769 ',
      cavesWorldGenPreset: 'DST_CAVE_PLUS',
      cavesWorldSettings: {
        taskSet: 'cave_default',
        worldSize: 'medium',
        branching: '',
        loop: 'default',
        startLocation: 'caves',
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
        cavelight: 'fast',
        prefabSwapsStart: 'classic',
        moonFissure: '',
        balatro: '',
        terrariumChest: '',
        stagePlays: '',
        junkyard: '',
        frogRain: '',
        meteorShowers: '',
        hunt: '',
        alternateHunt: '',
        wanderingTraderEnabled: '',
        riftsFrequency: '',
        riftsEnabled: '',
        lunarHailFrequency: '',
        mutatedHounds: '',
        penguinsMoon: '',
        moonSpider: 'rare',
        mutatedBirds: 'default',
        mutatedMerm: '',
        mutatedSpiderQueen: 'never',
        acidRainEnabled: 'always',
        riftsFrequencyCave: 'rare',
        riftsEnabledCave: 'always',
        portalSpawnRate: '',
        bananaBushPortalRate: '',
        lightCrabPortalRate: '',
        monkeytailPortalRate: '',
        palmconeSeedPortalRate: '',
        powderMonkeyPortalRate: '',
        diseaseDelay: 'long',
        basicResourceRegrowth: '',
        extraStartingItems: '',
        seasonalStartingItems: '',
        spawnProtection: '',
        dropEverythingOnDespawn: '',
        healthPenalty: '',
        lessDamageTaken: '',
        temperatureDamage: '',
        hunger: '',
        darkness: '',
        shadowCreatures: '',
        brightmareCreatures: '',
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
        yearOfTheKnight: '',
        cavePonds: 'often',
        earthquakes: 'rare',
        wormAttacks: 'never',
        winterHounds: '',
        summerHounds: '',
        wormAttacksBoss: 'often',
        atriumGate: 'fast'
      },
      cavesExtraWorldGenOverrides: 'mushtree=often'
    };

    expect(clusterRequestFromForm(form)).toEqual({
      clusterName: 'Managed DST',
      clusterDescription: 'Seasonal world',
      clusterPassword: 'secret',
      clusterIntention: 'social',
      gameMode: 'endless',
      maxPlayers: 8,
      language: 'zh',
      pvp: true,
      pauseWhenEmpty: false,
      offlineCluster: true,
      lanOnlyCluster: false,
      tickRate: 30,
      consoleEnabled: true,
      bindIP: '0.0.0.0',
      masterPort: 12000,
      clusterKey: 'cluster-abc',
      masterWorldSettings: [
        { key: 'basicresource_regrowth', value: 'always' },
        { key: 'beefaloheat', value: 'often' },
        { key: 'boons', value: 'always' },
        { key: 'brightmarecreatures', value: 'often' },
        { key: 'crow_carnival', value: 'enabled' },
        { key: 'darkness', value: 'nonlethal' },
        { key: 'dropeverythingondespawn', value: 'always' },
        { key: 'extrastartingitems', value: '15' },
        { key: 'ghostenabled', value: 'always' },
        { key: 'ghostsanitydrain', value: 'none' },
        { key: 'healthpenalty', value: 'none' },
        { key: 'hunger', value: 'default' },
        { key: 'krampus', value: 'rare' },
        { key: 'lessdamagetaken', value: 'always' },
        { key: 'portalresurection', value: 'always' },
        { key: 'resettime', value: 'fast' },
        { key: 'roads', value: 'often' },
        { key: 'seasonalstartingitems', value: 'default' },
        { key: 'shadowcreatures', value: 'rare' },
        { key: 'spawnmode', value: 'scatter' },
        { key: 'spawnprotection', value: 'always' },
        { key: 'specialevent', value: 'none' },
        { key: 'temperaturedamage', value: 'nonlethal' },
        { key: 'touchstone', value: 'rare' },
        { key: 'winters_feast', value: 'enabled' }
      ],
      shards: [
        {
          name: 'Master',
          enabled: true,
          serverPort: 11000,
          masterServerPort: 27020,
          authenticationPort: 8768,
          worldGenPreset: 'SURVIVAL_TOGETHER_CLASSIC',
          worldGenOverrides: [
            { key: 'alternatehunt', value: 'never' },
            { key: 'autumn', value: 'longseason' },
            { key: 'balatro', value: 'default' },
            { key: 'bananabush_portalrate', value: 'rare' },
            { key: 'bearger', value: 'rare' },
            { key: 'branching', value: 'most' },
            { key: 'day', value: 'longday' },
            { key: 'disease_delay', value: 'short' },
            { key: 'frograin', value: 'often' },
            { key: 'hounds', value: 'default' },
            { key: 'hunt', value: 'default' },
            { key: 'junkyard', value: 'default' },
            { key: 'lightcrab_portalrate', value: 'default' },
            { key: 'lightning', value: 'rare' },
            { key: 'loop', value: 'always' },
            { key: 'lunarhail_frequency', value: 'rare' },
            { key: 'meteorshowers', value: 'rare' },
            { key: 'monkeytail_portalrate', value: 'never' },
            { key: 'moon_fissure', value: 'mostly' },
            { key: 'moon_spider', value: 'often' },
            { key: 'mutated_birds', value: 'default' },
            { key: 'mutated_hounds', value: 'default' },
            { key: 'mutated_merm', value: 'never' },
            { key: 'mutated_spiderqueen', value: 'default' },
            { key: 'palmcone_seed_portalrate', value: 'always' },
            { key: 'penguins_moon', value: 'never' },
            { key: 'petrification', value: 'many' },
            { key: 'portal_spawnrate', value: 'often' },
            { key: 'powder_monkey_portalrate', value: 'default' },
            { key: 'prefabswaps_start', value: 'highly random' },
            { key: 'rifts_enabled', value: 'always' },
            { key: 'rifts_frequency', value: 'often' },
            { key: 'season_start', value: 'autumn' },
            { key: 'stageplays', value: 'never' },
            { key: 'start_location', value: 'plus' },
            { key: 'summerhounds', value: 'never' },
            { key: 'task_set', value: 'classic' },
            { key: 'terrariumchest', value: 'default' },
            { key: 'wanderingtrader_enabled', value: 'always' },
            { key: 'weather', value: 'often' },
            { key: 'wildfires', value: 'never' },
            { key: 'winter', value: 'default' },
            { key: 'winterhounds', value: 'default' },
            { key: 'world_size', value: 'huge' }
          ]
        },
        {
          name: 'Caves',
          enabled: false,
          serverPort: 11001,
          masterServerPort: 27021,
          authenticationPort: 8769,
          worldGenPreset: 'DST_CAVE_PLUS',
          worldGenOverrides: [
            { key: 'acidrain_enabled', value: 'always' },
            { key: 'atriumgate', value: 'fast' },
            { key: 'cave_ponds', value: 'often' },
            { key: 'cavelight', value: 'fast' },
            { key: 'disease_delay', value: 'long' },
            { key: 'earthquakes', value: 'rare' },
            { key: 'loop', value: 'default' },
            { key: 'moon_spider', value: 'rare' },
            { key: 'mushtree', value: 'often' },
            { key: 'mutated_birds', value: 'default' },
            { key: 'mutated_spiderqueen', value: 'never' },
            { key: 'prefabswaps_start', value: 'classic' },
            { key: 'rifts_enabled_cave', value: 'always' },
            { key: 'rifts_frequency_cave', value: 'rare' },
            { key: 'start_location', value: 'caves' },
            { key: 'task_set', value: 'cave_default' },
            { key: 'world_size', value: 'medium' },
            { key: 'wormattacks', value: 'never' },
            { key: 'wormattacks_boss', value: 'often' }
          ]
        }
      ]
    });
  });

  it('preserves unknown world overrides in the extra overrides textarea', () => {
    const config = sampleConfig();
    config.masterWorldSettings = [
      { key: 'ghostsanitydrain', value: 'none' },
      { key: 'beefaloheat', value: 'often' },
      { key: 'portalresurection', value: 'always' },
      { key: 'winters_feast', value: 'enabled' },
      { key: 'basicresource_regrowth', value: 'always' },
      { key: 'extrastartingitems', value: '15' },
      { key: 'shadowcreatures', value: 'rare' }
    ];
    config.shards[0].worldGenOverrides = [
      { key: 'season_start', value: 'autumn' },
      { key: 'world_size', value: 'huge' },
      { key: 'task_set', value: 'classic' },
      { key: 'prefabswaps_start', value: 'highly random' },
      { key: 'moon_fissure', value: 'mostly' },
      { key: 'balatro', value: 'default' },
      { key: 'hounds', value: 'rare' },
      { key: 'frograin', value: 'often' },
      { key: 'rifts_frequency', value: 'often' },
      { key: 'moon_spider', value: 'often' },
      { key: 'mutated_birds', value: 'default' },
      { key: 'mutated_merm', value: 'never' },
      { key: 'portal_spawnrate', value: 'rare' },
      { key: 'penguins_moon', value: 'default' },
      { key: 'spawnprotection', value: 'always' },
      { key: 'specialevent', value: 'none' },
      { key: 'beefalo', value: 'often' }
    ];

    const form = clusterFormFromConfig(config);

    expect(form.masterWorldSettings.seasonStart).toBe('autumn');
    expect(form.masterWorldSettings.worldSize).toBe('huge');
    expect(form.masterWorldSettings.taskSet).toBe('classic');
    expect(form.masterWorldSettings.prefabSwapsStart).toBe('highly random');
    expect(form.masterWorldSettings.moonFissure).toBe('mostly');
    expect(form.masterWorldSettings.balatro).toBe('default');
    expect(form.masterWorldSettings.hounds).toBe('rare');
    expect(form.masterWorldSettings.frogRain).toBe('often');
    expect(form.masterWorldSettings.riftsFrequency).toBe('often');
    expect(form.masterWorldSettings.moonSpider).toBe('often');
    expect(form.masterWorldSettings.mutatedBirds).toBe('default');
    expect(form.masterWorldSettings.mutatedMerm).toBe('never');
    expect(form.masterWorldSettings.penguinsMoon).toBe('default');
    expect(form.masterWorldSettings.portalSpawnRate).toBe('rare');
    expect(form.masterWorldSettings.spawnProtection).toBe('always');
    expect(form.masterWorldSettings.specialEvent).toBe('none');
    expect(form.masterWorldSettings.ghostSanityDrain).toBe('none');
    expect(form.masterWorldSettings.beefaloHeat).toBe('often');
    expect(form.masterWorldSettings.portalResurrection).toBe('always');
    expect(form.masterWorldSettings.wintersFeast).toBe('enabled');
    expect(form.masterWorldSettings.basicResourceRegrowth).toBe('always');
    expect(form.masterWorldSettings.extraStartingItems).toBe('15');
    expect(form.masterWorldSettings.shadowCreatures).toBe('rare');
    expect(form.masterExtraWorldGenOverrides).toBe('beefalo=often');
  });

  it('marks the form dirty when editable fields change', () => {
    const config = sampleConfig();
    const form = clusterFormFromConfig(config);
    form.clusterName = 'New Cluster';

    expect(clusterFormIsDirty(form, config)).toBe(true);
  });

  it('splits master shard generation fields from master-controlled rules', () => {
    expect(masterWorldGenSettingFields.some((field) => field.formKey === 'seasonStart')).toBe(true);
    expect(masterWorldGenSettingFields.some((field) => field.formKey === 'specialEvent')).toBe(false);
    expect(masterWorldGenSettingFields.some((field) => field.formKey === 'wormAttacks')).toBe(false);
    expect(masterWorldGenSettingFields.some((field) => field.formKey === 'mutatedHounds')).toBe(true);
    expect(masterWorldGenSettingFields.some((field) => field.formKey === 'moonSpider')).toBe(true);

    expect(masterWorldControlSettingFields.some((field) => field.formKey === 'specialEvent')).toBe(true);
    expect(masterWorldControlSettingFields.some((field) => field.formKey === 'portalResurrection')).toBe(true);
    expect(masterWorldControlSettingFields.some((field) => field.formKey === 'seasonStart')).toBe(false);

    expect(cavesWorldSettingFields.some((field) => field.formKey === 'wormAttacks')).toBe(true);
    expect(cavesWorldSettingFields.some((field) => field.formKey === 'specialEvent')).toBe(false);
    expect(cavesWorldSettingFields.some((field) => field.formKey === 'moonSpider')).toBe(true);
    expect(cavesWorldSettingFields.some((field) => field.formKey === 'mutatedHounds')).toBe(false);
  });
});
