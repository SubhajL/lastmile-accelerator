import { describe, expect, test } from 'vitest';

import { Subjects, SubscriptionSubjects, isEventType } from './subjects';

describe('subjects registry', () => {
  test('exposes stable subject constants', () => {
    expect(Subjects.snapshot.ready).toEqual('snapshot.ready');
    expect(Subjects.fixes.created).toEqual('fixes.created');
    expect(Subjects.fixes.applied).toEqual('fixes.applied');
    expect(Subjects.publish.started).toEqual('publish.started');
    expect(Subjects.publish.healthy).toEqual('publish.healthy');
    expect(Subjects.publish.rolledback).toEqual('publish.rolledback');
    expect(Subjects.publish.failed).toEqual('publish.failed');
    expect(Subjects.checks.failed).toEqual('checks.failed');
    expect(Subjects.slo.budget_exhausted).toEqual('slo.budget_exhausted');
    expect(Subjects.errors.critical).toEqual('errors.critical');
  });

  test('provides wildcard subscription patterns', () => {
    expect(SubscriptionSubjects).toEqual([
      'snapshot.*',
      'fixes.*',
      'publish.*',
      'checks.*',
      'slo.*',
      'errors.*'
    ]);
  });

  test('isEventType narrows known values', () => {
    expect(isEventType('snapshot.ready')).toBe(true);
    expect(isEventType('fixes.created')).toBe(true);
    expect(isEventType('unknown.event')).toBe(false);
  });
});
