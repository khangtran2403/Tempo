CREATE TYPE execution_status AS ENUM ('running', 'success', 'failed', 'timeout');

CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    temporal_execution_id VARCHAR(255), -- Link to Temporal execution
    status execution_status DEFAULT 'running',
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms INTEGER,
    error_message TEXT,
    input_data JSONB, -- Data từ trigger
    output_data JSONB, -- Output từ tất cả actions
    CONSTRAINT valid_duration CHECK (duration_ms IS NULL OR duration_ms >= 0)
);

CREATE INDEX idx_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX idx_executions_user_id ON workflow_executions(user_id);
CREATE INDEX idx_executions_status ON workflow_executions(status);
CREATE INDEX idx_executions_created_at ON workflow_executions(started_at DESC);

-- Trigger để tự động tính toán duration
CREATE OR REPLACE FUNCTION set_execution_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL THEN
        NEW.duration_ms := EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_execution_duration
BEFORE UPDATE ON workflow_executions
FOR EACH ROW
EXECUTE FUNCTION set_execution_duration();