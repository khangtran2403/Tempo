DROP TRIGGER IF EXISTS trigger_execution_duration ON workflow_executions;
DROP FUNCTION IF EXISTS set_execution_duration();
DROP TABLE IF EXISTS workflow_executions;
DROP TYPE IF EXISTS execution_status;