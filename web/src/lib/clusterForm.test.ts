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
    gameMode: 'survival',
    maxPlayers: 6,
    language: 'en',
    pvp: false,
    pauseWhenEmpty: true,
    shards: [
      { name: 'Master', enabled: true },
      { name: 'Caves', enabled: true }
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
    expect(form.masterEnabled).toBe(true);
    expect(form.cavesEnabled).toBe(true);
    expect(clusterFormIsDirty(form, config)).toBe(false);
  });

  it('builds a stable update payload from form state', () => {
    const form: ClusterFormState = {
      clusterName: 'Managed DST',
      clusterDescription: 'Seasonal world',
      gameMode: 'endless',
      maxPlayers: ' 8 ',
      language: 'zh',
      pvp: true,
      pauseWhenEmpty: false,
      masterEnabled: true,
      cavesEnabled: false
    };

    expect(clusterRequestFromForm(form)).toEqual({
      clusterName: 'Managed DST',
      clusterDescription: 'Seasonal world',
      gameMode: 'endless',
      maxPlayers: 8,
      language: 'zh',
      pvp: true,
      pauseWhenEmpty: false,
      shards: [
        { name: 'Master', enabled: true },
        { name: 'Caves', enabled: false }
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
