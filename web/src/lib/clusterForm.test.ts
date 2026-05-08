import { describe, expect, it } from 'vitest';

import {
  clusterFormFromConfig,
  clusterFormIsDirty,
  clusterRequestFromForm,
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
    shards: [
      { name: 'Master', enabled: true, serverPort: 10999, masterServerPort: 27016, authenticationPort: 8766 },
      { name: 'Caves', enabled: true, serverPort: 11000, masterServerPort: 27017, authenticationPort: 8767 }
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
      cavesServerPort: ' 11001 ',
      cavesMasterServerPort: ' 27021 ',
      cavesAuthenticationPort: ' 8769 '
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
      shards: [
        { name: 'Master', enabled: true, serverPort: 11000, masterServerPort: 27020, authenticationPort: 8768 },
        { name: 'Caves', enabled: false, serverPort: 11001, masterServerPort: 27021, authenticationPort: 8769 }
      ]
    });
  });

  it('marks the form dirty when editable fields change', () => {
    const config = sampleConfig();
    const form = clusterFormFromConfig(config);
    form.clusterName = 'New Cluster';

    expect(clusterFormIsDirty(form, config)).toBe(true);
  });
});
