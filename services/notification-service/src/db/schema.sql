-- notification_templates: Template definitions for notifications
CREATE TABLE IF NOT EXISTS notification_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  channel VARCHAR(50) NOT NULL CHECK (channel IN ('email', 'sms', 'in-app', 'webhook', 'slack')),
  subject VARCHAR(500),
  body TEXT NOT NULL,
  variables JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_templates_channel ON notification_templates(channel);
CREATE INDEX idx_templates_name ON notification_templates(name);

-- notification_logs: Record of all sent notifications
CREATE TABLE IF NOT EXISTS notification_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  user_id UUID,
  channel VARCHAR(50) NOT NULL,
  template_id UUID REFERENCES notification_templates(id),
  status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'sent', 'failed', 'retrying')),
  sent_at TIMESTAMP WITH TIME ZONE,
  error TEXT,
  metadata JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_logs_tenant_id ON notification_logs(tenant_id);
CREATE INDEX idx_logs_user_id ON notification_logs(user_id);
CREATE INDEX idx_logs_created_at ON notification_logs(created_at DESC);
CREATE INDEX idx_logs_status ON notification_logs(status);
CREATE INDEX idx_logs_channel ON notification_logs(channel);

-- notification_preferences: User preferences for notifications
CREATE TABLE IF NOT EXISTS notification_preferences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  channel VARCHAR(50) NOT NULL,
  enabled BOOLEAN DEFAULT true,
  frequency VARCHAR(50) DEFAULT 'realtime' CHECK (frequency IN ('realtime', 'digest', 'daily', 'weekly')),
  quiet_hours_start TIME,
  quiet_hours_end TIME,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(user_id, channel)
);

CREATE INDEX idx_preferences_user_id ON notification_preferences(user_id);

-- delivery_attempts: Track delivery retry attempts
CREATE TABLE IF NOT EXISTS delivery_attempts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  notification_id UUID NOT NULL REFERENCES notification_logs(id) ON DELETE CASCADE,
  attempt INTEGER NOT NULL,
  provider VARCHAR(100),
  status VARCHAR(50) NOT NULL,
  sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  error TEXT
);

CREATE INDEX idx_attempts_notification_id ON delivery_attempts(notification_id);
CREATE INDEX idx_attempts_status ON delivery_attempts(status);

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update trigger to templates
CREATE TRIGGER update_notification_templates_updated_at
  BEFORE UPDATE ON notification_templates
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- Apply update trigger to preferences
CREATE TRIGGER update_notification_preferences_updated_at
  BEFORE UPDATE ON notification_preferences
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
