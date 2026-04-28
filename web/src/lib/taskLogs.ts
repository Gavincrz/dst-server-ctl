type TaskLogButtonState = {
  loading: boolean;
  expanded: boolean;
  hasLogs: boolean;
};

export function taskLogButtonLabel(state: TaskLogButtonState) {
  if (state.loading) {
    return 'Loading Logs';
  }
  if (state.expanded) {
    return 'Hide Logs';
  }
  if (state.hasLogs) {
    return 'Show Logs';
  }
  return 'View Logs';
}
