CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY,
    slo_id UUID NOT NULL REFERENCES slos(id) ON DELETE CASCADE,
    threshold DECIMAL(5,2) NOT NULL CHECK (threshold >= 0 AND threshold <= 100),
    channels JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_rules_slo_id ON alert_rules(slo_id);

CREATE TABLE IF NOT EXISTS alert_history (
    id UUID PRIMARY KEY,
    alert_rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    slo_id UUID NOT NULL REFERENCES slos(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    compliance DECIMAL(5,2) NOT NULL,
    threshold DECIMAL(5,2) NOT NULL,
    notified BOOLEAN NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_alert_history_rule_timestamp ON alert_history(alert_rule_id, timestamp DESC);
