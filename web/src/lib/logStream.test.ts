import { describe, expect, it, vi } from 'vitest';

import { closeLogStream, connectLogStream, type EventSourceLike } from './logStream';

class FakeEventSource implements EventSourceLike {
  readyState = 1;
  onerror: ((event: Event) => void) | null = null;
  closed = false;
  listeners = new Map<string, Array<(event: MessageEvent<string>) => void>>();

  addEventListener(type: string, listener: (event: MessageEvent<string>) => void) {
    const handlers = this.listeners.get(type) ?? [];
    handlers.push(listener);
    this.listeners.set(type, handlers);
  }

  close() {
    this.closed = true;
    this.readyState = 2;
  }

  emit(type: string, payload: unknown) {
    const event = { data: JSON.stringify(payload) } as MessageEvent<string>;
    for (const listener of this.listeners.get(type) ?? []) {
      listener(event);
    }
  }

  emitError() {
    this.onerror?.({} as Event);
  }
}

describe('closeLogStream', () => {
  it('closes the current source and clears the reference', () => {
    const source = new FakeEventSource();
    let current: EventSourceLike | null = source;

    closeLogStream(() => current, (next) => {
      current = next;
    });

    expect(source.closed).toBe(true);
    expect(current).toBeNull();
  });
});

describe('connectLogStream', () => {
  it('wires snapshot, append and stream errors from EventSource', () => {
    const source = new FakeEventSource();
    let current: EventSourceLike | null = null;
    const snapshot = vi.fn();
    const append = vi.fn();
    const streamError = vi.fn();
    const disconnect = vi.fn();

    connectLogStream<{ lines: string[] }>({
      url: '/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: () => source,
      loadFallback: vi.fn(),
      onSnapshot: snapshot,
      onAppend: append,
      onStreamError: streamError,
      onDisconnect: disconnect,
      unavailableMessage: 'logs unavailable',
      disconnectedMessage: 'stream disconnected'
    });

    source.emit('snapshot', { lines: ['boot'] });
    source.emit('append', { lines: ['ready'] });
    source.emit('stream-error', { error: 'backend failed' });

    expect(snapshot).toHaveBeenCalledWith({ lines: ['boot'] });
    expect(append).toHaveBeenCalledWith({ lines: ['ready'] });
    expect(streamError).toHaveBeenCalledWith('backend failed');
    expect(disconnect).not.toHaveBeenCalled();
  });

  it('reports disconnects for the active source', () => {
    const source = new FakeEventSource();
    let current: EventSourceLike | null = null;
    const disconnect = vi.fn();

    connectLogStream({
      url: '/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: () => source,
      loadFallback: vi.fn(),
      onSnapshot: vi.fn(),
      onAppend: vi.fn(),
      onStreamError: vi.fn(),
      onDisconnect: disconnect,
      unavailableMessage: 'logs unavailable',
      disconnectedMessage: 'stream disconnected'
    });

    source.emitError();

    expect(disconnect).toHaveBeenCalledWith('stream disconnected');
  });

  it('ignores events from a stale source after reconnect', () => {
    const first = new FakeEventSource();
    const second = new FakeEventSource();
    const created = [first, second];
    let current: EventSourceLike | null = null;
    const snapshot = vi.fn();

    const connect = () =>
      connectLogStream<{ lines: string[] }>({
        url: '/stream',
        current: () => current,
        replace: (next) => {
          current = next;
        },
        createEventSource: () => {
          const source = created.shift();
          if (!source) {
            throw new Error('missing source');
          }
          return source;
        },
        loadFallback: vi.fn(),
        onSnapshot: snapshot,
        onAppend: vi.fn(),
        onStreamError: vi.fn(),
        onDisconnect: vi.fn(),
        unavailableMessage: 'logs unavailable',
        disconnectedMessage: 'stream disconnected'
      });

    connect();
    connect();

    first.emit('snapshot', { lines: ['stale'] });
    second.emit('snapshot', { lines: ['fresh'] });

    expect(first.closed).toBe(true);
    expect(snapshot).toHaveBeenCalledTimes(1);
    expect(snapshot).toHaveBeenCalledWith({ lines: ['fresh'] });
  });

  it('falls back when EventSource is unavailable', async () => {
    let current: EventSourceLike | null = null;
    const fallback = vi.fn();

    connectLogStream({
      url: '/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: null,
      loadFallback: fallback,
      onSnapshot: vi.fn(),
      onAppend: vi.fn(),
      onStreamError: vi.fn(),
      onDisconnect: vi.fn(),
      unavailableMessage: 'logs unavailable',
      disconnectedMessage: 'stream disconnected'
    });

    await Promise.resolve();

    expect(fallback).toHaveBeenCalledTimes(1);
    expect(current).toBeNull();
  });
});
