ALTER TABLE workflows 
ADD COLUMN version INTEGER DEFAULT 1,
ADD COLUMN parent_id UUID REFERENCES workflows(id),
ADD COLUMN is_latest BOOLEAN DEFAULT true,
ADD COLUMN published_at TIMESTAMP;

-- Create workflow_versions table
CREATE TABLE workflow_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    definition JSONB NOT NULL,
    change_summary TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT false,
    
    UNIQUE(workflow_id, version)
);

CREATE INDEX idx_workflow_versions_workflow_id ON workflow_versions(workflow_id);
CREATE INDEX idx_workflow_versions_version ON workflow_versions(workflow_id, version DESC);

-- Trigger để auto-increment version
CREATE OR REPLACE FUNCTION increment_workflow_version()
RETURNS TRIGGER AS $$
BEGIN
    -- Get max version
    SELECT COALESCE(MAX(version), 0) + 1 
    INTO NEW.version
    FROM workflow_versions
    WHERE workflow_id = NEW.workflow_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_workflow_version
BEFORE INSERT ON workflow_versions
FOR EACH ROW
EXECUTE FUNCTION increment_workflow_version();