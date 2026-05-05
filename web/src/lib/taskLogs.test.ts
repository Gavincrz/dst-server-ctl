import { describe, expect, it } from 'vitest';

import { activeExpandedLogIDsFor, activeExpandedTaskIDs, expandedLogIDs, expandedTaskIDs, taskLogButtonLabel } from './taskLogs';

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

  it('uses custom labels for non-task log panels', () => {
    expect(taskLogButtonLabel(
      { loading: false, expanded: false, hasLogs: false },
      {
        loading: 'Loading Check Logs',
        hide: 'Hide Check Logs',
        show: 'Show Check Logs',
        view: 'View Check Logs'
      }
    )).toBe('View Check Logs');
  });
});

describe('expandedTaskIDs', () => {
  it('returns expanded ids for generic log panels', () => {
    expect(expandedLogIDs({ Master: true, Caves: false, Extra: true }, [{ id: 'Master' }, { id: 'Caves' }])).toEqual(['Master']);
  });

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

describe('activeExpandedTaskIDs', () => {
  it('returns expanded generic log ids that still match the activity predicate', () => {
    expect(activeExpandedLogIDsFor(
      { Master: true, Caves: true, Extra: true },
      [
        { id: 'Master', running: true },
        { id: 'Caves', running: false }
      ],
      (item) => 'running' in item && item.running
    )).toEqual(['Master']);
  });

  it('returns only expanded tasks that are still active', () => {
    expect(activeExpandedTaskIDs(
      { 'task-1': true, 'task-2': true, 'task-3': true, 'task-4': false },
      [
        { id: 'task-1', status: 'running' },
        { id: 'task-2', status: 'pending' },
        { id: 'task-3', status: 'succeeded' },
        { id: 'task-4', status: 'running' }
      ]
    )).toEqual(['task-1', 'task-2']);
  });
});
