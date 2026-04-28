type TaskLogButtonState = {
  loading: boolean;
  expanded: boolean;
  hasLogs: boolean;
};

type TaskLogTask = {
  id: string;
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

export function expandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return tasks.filter((task) => expanded[task.id]).map((task) => task.id);
}
