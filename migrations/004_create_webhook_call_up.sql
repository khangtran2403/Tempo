CREATE TABLE webhook_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    method VARCHAR(10) NOT NULL,
    headers JSONB,
    body JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    execution_id UUID,
    status VARCHAR(20) NOT NULL,
    error TEXT,
    received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhook_history_workflow_id ON webhook_history(workflow_id);
CREATE INDEX idx_webhook_history_received_at ON webhook_history(received_at DESC);