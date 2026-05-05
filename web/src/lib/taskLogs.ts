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

export type KeyedLogCollectionState = {
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

type ExpandedItem = {
  id: string;
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

export function expandedLogIDs<T extends ExpandedItem>(expanded: Record<string, boolean>, items: T[]) {
  return items.filter((item) => expanded[item.id]).map((item) => item.id);
}

export function expandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return expandedLogIDs(expanded, tasks);
}

export function activeExpandedLogIDs(
  expanded: Record<string, boolean>,
  items: ExpandedItem[],
  isActive: (item: ExpandedItem) => boolean
) {
  return items
    .filter((item) => expanded[item.id] && isActive(item))
    .map((item) => item.id);
}

export function activeExpandedLogIDsFor<T extends ExpandedItem>(
  expanded: Record<string, boolean>,
  items: T[],
  isActive: (item: T) => boolean
) {
  return items
    .filter((item) => expanded[item.id] && isActive(item))
    .map((item) => item.id);
}

export function activeExpandedTaskIDs(expanded: Record<string, boolean>, tasks: TaskLogTask[]) {
  return activeExpandedLogIDsFor(expanded, tasks, (task) => task.status === 'pending' || task.status === 'running');
}
