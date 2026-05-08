export type EventSourceLike = {
  addEventListener(type: string, listener: (event: MessageEvent<string>) => void): void;
  close(): void;
  readyState: number;
  onerror: ((event: Event) => void) | null;
};

export type EventSourceFactory = (url: string) => EventSourceLike;

export type LogStreamOptions<T> = {
  url: string;
  current: () => EventSourceLike | null | undefined;
  replace: (source: EventSourceLike | null) => void;
  createEventSource?: EventSourceFactory | null;
  loadFallback: () => void | Promise<void>;
  onSnapshot: (payload: T) => void;
  onAppend: (payload: T) => void;
  onStreamError: (message: string) => void;
  onDisconnect: (message: string) => void;
  unavailableMessage: string;
  disconnectedMessage: string;
};

const eventSourceClosedState = 2;

export function closeLogStream(current: () => EventSourceLike | null | undefined, replace: (source: EventSourceLike | null) => void) {
  current()?.close();
  replace(null);
}

export function connectLogStream<T>(options: LogStreamOptions<T>) {
  const {
    url,
    current,
    replace,
    createEventSource,
    loadFallback,
    onSnapshot,
    onAppend,
    onStreamError,
    onDisconnect,
    unavailableMessage,
    disconnectedMessage
  } = options;

  closeLogStream(current, replace);

  if (!createEventSource) {
    void loadFallback();
    return;
  }

  const source = createEventSource(url);
  replace(source);

  source.addEventListener('snapshot', (event) => {
    if (current() !== source) {
      return;
    }
    onSnapshot(JSON.parse(event.data) as T);
  });

  source.addEventListener('append', (event) => {
    if (current() !== source) {
      return;
    }
    onAppend(JSON.parse(event.data) as T);
  });

  source.addEventListener('stream-error', (event) => {
    if (current() !== source) {
      return;
    }
    const payload = JSON.parse(event.data) as { error?: string };
    onStreamError(payload.error || unavailableMessage);
  });

  source.onerror = () => {
    if (current() !== source || source.readyState === eventSourceClosedState) {
      return;
    }
    onDisconnect(disconnectedMessage);
  };
}
