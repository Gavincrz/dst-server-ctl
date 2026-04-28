type TaskLogButtonState = {
  loading: boolean;
  expanded: boolean;
  hasLogs: boolean;
};

type TaskLogTask = {
  id: string;
  status?: string;
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

export function activeExpandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return tasks
    .filter((task) => expanded[task.id] && (task.status === 'pending' || task.status === 'running'))
    .map((task) => task.id);
}
