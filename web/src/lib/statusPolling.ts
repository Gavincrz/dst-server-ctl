export const activeTaskPollIntervalMs = 3000;
export const runtimeOnlyPollIntervalMs = 10000;

export type StatusPollingState = {
  hasActiveInstallTasks: boolean;
  hasActiveUpdateTasks: boolean;
  runtimeRunning: boolean;
  installSubmitting: boolean;
  updateSubmitting: boolean;
  updateCheckSubmitting: boolean;
  runtimeSubmitting: boolean;
  polling: boolean;
};

export function statusPollIntervalMs(state: StatusPollingState): number | null {
  if (
    state.installSubmitting ||
    state.updateSubmitting ||
    state.updateCheckSubmitting ||
    state.runtimeSubmitting ||
    state.polling
  ) {
    return null;
  }

  if (state.hasActiveInstallTasks || state.hasActiveUpdateTasks) {
    return activeTaskPollIntervalMs;
  }

  if (state.runtimeRunning) {
    return runtimeOnlyPollIntervalMs;
  }

  return null;
}
