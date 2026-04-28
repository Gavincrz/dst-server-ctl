import { describe, expect, it } from 'vitest';

import { taskLogButtonLabel } from './taskLogs';

describe('taskLogButtonLabel', () => {
  it('shows loading while a request is in flight', () => {
    expect(taskLogButtonLabel({ loading: true, expanded: false, hasLogs: false })).toBe('Loading Logs');
  });

  it('shows hide when logs are expanded', () => {
    expect(taskLogButtonLabel({ loading: false, expanded: true, hasLogs: true })).toBe('Hide Logs');
  });

  it('shows show when logs were loaded but are collapsed', () => {
    expect(taskLogButtonLabel({ loading: false, expanded: false, hasLogs: true })).toBe('Show Logs');
  });

  it('shows view before any logs are loaded', () => {
    expect(taskLogButtonLabel({ loading: false, expanded: false, hasLogs: false })).toBe('View Logs');
  });
});
