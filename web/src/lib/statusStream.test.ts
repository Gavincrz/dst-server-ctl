import { describe, expect, it, vi } from 'vitest';

import { connectStatusStream, closeStatusStream, type StatusStreamOptions } from './statusStream';
import type { EventSourceLike } from './logStream';

class FakeEventSource implements EventSourceLike {
  readonly listeners = new Map<string, Array<(event: MessageEvent<string>) => void>>();
  readyState = 1;
  onerror: ((event: Event) => void) | null = null;
  closed = false;

  addEventListener(type: string, listener: (event: MessageEvent<string>) => void): void {
    const listeners = this.listeners.get(type) ?? [];
    listeners.push(listener);
    this.listeners.set(type, listeners);
  }

  close(): void {
    this.closed = true;
    this.readyState = 2;
  }

  emit(type: string, payload: unknown) {
    const event = { data: JSON.stringify(payload) } as MessageEvent<string>;
    for (const listener of this.listeners.get(type) ?? []) {
      listener(event);
    }
  }
}

describe('statusStream', () => {
  it('uses fallback when EventSource is unavailable', async () => {
    const loadFallback = vi.fn();

    connectStatusStream({
      url: '/api/v1/dashboard/stream',
      current: () => null,
      replace: () => undefined,
      createEventSource: null,
      loadFallback,
      onSnapshot: () => undefined,
      onStreamError: () => undefined,
      onDisconnect: () => undefined,
      unavailableMessage: 'dashboard unavailable',
      disconnectedMessage: 'dashboard disconnected'
    });

    expect(loadFallback).toHaveBeenCalledTimes(1);
  });

  it('forwards snapshot events to the handler', () => {
    let current: EventSourceLike | null = null;
    const source = new FakeEventSource();
    const onSnapshot = vi.fn();

    connectStatusStream<{ value: string }>({
      url: '/api/v1/dashboard/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: () => source,
      loadFallback: () => undefined,
      onSnapshot,
      onStreamError: () => undefined,
      onDisconnect: () => undefined,
      unavailableMessage: 'dashboard unavailable',
      disconnectedMessage: 'dashboard disconnected'
    });

    source.emit('snapshot', { value: 'updated' });
    expect(onSnapshot).toHaveBeenCalledWith({ value: 'updated' });
  });

  it('forwards stream-error payloads', () => {
    let current: EventSourceLike | null = null;
    const source = new FakeEventSource();
    const onStreamError = vi.fn();

    connectStatusStream({
      url: '/api/v1/dashboard/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: () => source,
      loadFallback: () => undefined,
      onSnapshot: () => undefined,
      onStreamError,
      onDisconnect: () => undefined,
      unavailableMessage: 'dashboard unavailable',
      disconnectedMessage: 'dashboard disconnected'
    });

    source.emit('stream-error', { error: 'broken' });
    expect(onStreamError).toHaveBeenCalledWith('broken');
  });

  it('reports disconnects while the stream stays open', () => {
    let current: EventSourceLike | null = null;
    const source = new FakeEventSource();
    const onDisconnect = vi.fn();

    connectStatusStream({
      url: '/api/v1/dashboard/stream',
      current: () => current,
      replace: (next) => {
        current = next;
      },
      createEventSource: () => source,
      loadFallback: () => undefined,
      onSnapshot: () => undefined,
      onStreamError: () => undefined,
      onDisconnect,
      unavailableMessage: 'dashboard unavailable',
      disconnectedMessage: 'dashboard disconnected'
    });

    source.onerror?.(new Event('error'));
    expect(onDisconnect).toHaveBeenCalledWith('dashboard disconnected');
  });

  it('closes the active stream', () => {
    let current: EventSourceLike | null = new FakeEventSource();
    const replace = vi.fn((next: EventSourceLike | null) => {
      current = next;
    });

    closeStatusStream(() => current, replace);

    expect((replace.mock.calls[0] as [EventSourceLike | null])[0]).toBeNull();
    expect(current).toBeNull();
  });
});
