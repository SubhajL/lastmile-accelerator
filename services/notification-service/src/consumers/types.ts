export interface SnapshotCreatedEvent {
  snapshotId: string;
  projectId: string;
  tenantId: string;
  mode: 'A' | 'B' | 'C';
  createdAt: string;
}

export interface SnapshotReadyEvent {
  snapshotId: string;
  projectId: string;
  tenantId: string;
  checksCompleted: number;
  issuesFound: number;
}

export interface FixListCreatedEvent {
  snapshotId: string;
  projectId: string;
  tenantId: string;
  count: number;
  tierBreakdown: {
    tier1: number;
    tier2: number;
    tier3: number;
  };
}

export interface FixesAppliedEvent {
  snapshotId: string;
  projectId: string;
  tenantId: string;
  fixIds: string[];
  manifestUrl: string;
}

export interface PublishEvent {
  publishId: string;
  projectId: string;
  tenantId: string;
  status: 'started' | 'healthy' | 'rolledback' | 'failed';
  snapshotId: string;
  error?: string;
}

export interface CheckFailedEvent {
  checkId: string;
  projectId: string;
  tenantId: string;
  checkType: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
}

export interface SLOBudgetExhaustedEvent {
  sloId: string;
  projectId: string;
  tenantId: string;
  metric: string;
  budgetRemaining: number;
}

export interface CriticalErrorEvent {
  errorId: string;
  projectId: string;
  tenantId: string;
  service: string;
  message: string;
  stackTrace?: string;
}

export interface User {
  id: string;
  email: string;
  name?: string;
  roles?: string[];
}

export type NotificationPriority = 'critical' | 'high' | 'normal' | 'low';

export interface NotificationJob {
  id: string;
  tenantId: string;
  userId: string;
  channel: string;
  templateName: string;
  priority: NotificationPriority;
  payload: Record<string, unknown>;
  createdAt: string;
  attempt: number;
  maxAttempts: number;
}
