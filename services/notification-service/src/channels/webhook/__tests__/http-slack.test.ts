import { describe, it, expect, vi } from 'vitest';
import { createWebhookChannel } from '../../webhook/http.js';
import { createSlackChannel } from '../../slack/slack.js';

function job(): any { return { templateName: 'publish-failed', payload: { snapshotId: 's1', error: 'boom' } }; }

describe('webhook/http', () => {
  it('signs body and posts', async () => {
    const http = vi.fn().mockResolvedValue({ ok: true, status: 200, json: vi.fn() });
    const breaker = { execute: (fn: any) => fn() };
    const ch = createWebhookChannel({ http: http as any, url: 'https://example.com/hook', signingSecret: 'secret', metrics: { increment: vi.fn() }, breaker: breaker as any });

    const res = await ch.send(job());
    expect(res).toEqual({ ok: true });
    expect(http).toHaveBeenCalledWith('https://example.com/hook', expect.objectContaining({ method: 'POST', headers: expect.objectContaining({ 'X-Signature': expect.any(String) }) }));
  });
});

describe('slack/slack', () => {
  it('formats message and posts', async () => {
    const http = vi.fn().mockResolvedValue({ ok: true, status: 200, json: vi.fn() });
    const breaker = { execute: (fn: any) => fn() };
    const ch = createSlackChannel({ http: http as any, webhookUrl: 'https://hooks.slack.com/xxx', metrics: { increment: vi.fn() }, breaker: breaker as any });

    const res = await ch.send(job());
    expect(res).toEqual({ ok: true });
    expect(http).toHaveBeenCalledWith('https://hooks.slack.com/xxx', expect.objectContaining({ method: 'POST' }));
  });
});
