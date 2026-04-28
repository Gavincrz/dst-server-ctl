import { describe, expect, it } from 'vitest';

import { expandedTaskIDs, taskLogButtonLabel } from './taskLogs';

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

describe('expandedTaskIDs', () => {
  it('returns only ids whose panels are expanded', () => {
    expect(expandedTaskIDs({ 'task-1': true, 'task-2': false, 'task-3': true }, [{ id: 'task-1' }, { id: 'task-2' }, { id: 'task-3' }])).toEqual([
      'task-1',
      'task-3'
    ]);
  });

  it('ignores expanded ids that are not in the current task list', () => {
    expect(expandedTaskIDs({ 'task-1': true, missing: true }, [{ id: 'task-1' }])).toEqual(['task-1']);
  });
});
