DROP TRIGGER IF EXISTS trigger_increment_workflow_version ON workflow_versions;
DROP FUNCTION IF EXISTS increment_workflow_version();
DROP TABLE IF EXISTS workflow_versions;

ALTER TABLE workflows
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS parent_id,
DROP COLUMN IF EXISTS is_latest,
DROP COLUMN IF EXISTS published_at;