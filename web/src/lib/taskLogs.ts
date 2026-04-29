export type LogButtonState = {
  loading: boolean;
  expanded: boolean;
  hasLogs: boolean;
};

export type LogButtonLabels = {
  loading: string;
  hide: string;
  show: string;
  view: string;
};

export type SingleLogPanelState = {
  lines: string[];
  expanded: boolean;
  loading: boolean;
  error: string;
};

export type KeyedTaskLogCollectionState = {
  logs: Record<string, string[]>;
  loading: Record<string, boolean>;
  errors: Record<string, string>;
  expanded: Record<string, boolean>;
};

const defaultLogButtonLabels: LogButtonLabels = {
  loading: 'Loading Logs',
  hide: 'Hide Logs',
  show: 'Show Logs',
  view: 'View Logs'
};

type TaskLogTask = {
  id: string;
  status?: string;
};

export function taskLogButtonLabel(state: LogButtonState, labels: LogButtonLabels = defaultLogButtonLabels) {
  if (state.loading) {
    return labels.loading;
  }
  if (state.expanded) {
    return labels.hide;
  }
  if (state.hasLogs) {
    return labels.show;
  }
  return labels.view;
}

export function expandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return tasks.filter((task) => expanded[task.id]).map((task) => task.id);
}

export function activeExpandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return tasks
    .filter((task) => expanded[task.id] && (task.status === 'pending' || task.status === 'running'))
    .map((task) => task.id);
}
