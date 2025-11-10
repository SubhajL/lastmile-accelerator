import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as nats from '../../../nats';
import { publishProjectEvent, publishMemberEvent, publishEnvironmentEvent } from '../../../events/eventPublisher';

describe('events/eventPublisher.ts', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('publishProjectEvent publishes to projects.created with timestamp and traceparent', async () => {
    const spy = vi.spyOn(nats, 'publish').mockResolvedValue();

    await publishProjectEvent('created', { projectId: 'p1', tenantId: 't1' }, '00-abcd');

    expect(spy).toHaveBeenCalledTimes(1);
    const [subject, payload, trace] = spy.mock.calls[0];
    expect(subject).toBe('projects.created');
    expect(payload.projectId).toBe('p1');
    expect(payload.tenantId).toBe('t1');
    expect(typeof payload.timestamp).toBe('string');
    expect(trace).toBe('00-abcd');
  });

  it('publishMemberEvent publishes to members.role_changed', async () => {
    const spy = vi.spyOn(nats, 'publish').mockResolvedValue();

    await publishMemberEvent('role_changed', { memberId: 'm1', tenantId: 't1', newRole: 'admin' });

    const [subject, payload] = spy.mock.calls[0];
    expect(subject).toBe('members.role_changed');
    expect(payload.memberId).toBe('m1');
    expect(payload.newRole).toBe('admin');
    expect(typeof payload.timestamp).toBe('string');
  });

  it('publishEnvironmentEvent publishes to environments.deleted', async () => {
    const spy = vi.spyOn(nats, 'publish').mockResolvedValue();

    await publishEnvironmentEvent('deleted', { environmentId: 'e1', projectId: 'p1', tenantId: 't1' });

    const [subject, payload] = spy.mock.calls[0];
    expect(subject).toBe('environments.deleted');
    expect(payload.environmentId).toBe('e1');
    expect(typeof payload.timestamp).toBe('string');
  });
});