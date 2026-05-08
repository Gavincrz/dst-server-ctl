import { describe, expect, it } from 'vitest';

import {
  activeTaskPollIntervalMs,
  runtimeOnlyPollIntervalMs,
  statusPollIntervalMs,
  type StatusPollingState
} from './statusPolling';

function state(overrides: Partial<StatusPollingState> = {}): StatusPollingState {
  return {
    hasActiveInstallTasks: false,
    hasActiveUpdateTasks: false,
    runtimeRunning: false,
    installSubmitting: false,
    updateSubmitting: false,
    updateCheckSubmitting: false,
    runtimeSubmitting: false,
    polling: false,
    ...overrides
  };
}

describe('statusPollIntervalMs', () => {
  it('returns fast polling while install tasks are active', () => {
    expect(statusPollIntervalMs(state({ hasActiveInstallTasks: true }))).toBe(activeTaskPollIntervalMs);
  });

  it('returns fast polling while update tasks are active', () => {
    expect(statusPollIntervalMs(state({ hasActiveUpdateTasks: true }))).toBe(activeTaskPollIntervalMs);
  });

  it('returns slower polling when only runtime is running', () => {
    expect(statusPollIntervalMs(state({ runtimeRunning: true }))).toBe(runtimeOnlyPollIntervalMs);
  });

  it('disables polling while a foreground action is submitting', () => {
    expect(statusPollIntervalMs(state({ runtimeRunning: true, runtimeSubmitting: true }))).toBeNull();
  });

  it('disables polling while a background refresh is already in flight', () => {
    expect(statusPollIntervalMs(state({ runtimeRunning: true, polling: true }))).toBeNull();
  });

  it('disables polling when nothing is active', () => {
    expect(statusPollIntervalMs(state())).toBeNull();
  });
});
