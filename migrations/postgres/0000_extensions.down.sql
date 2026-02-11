-- Drop custom types
DROP TYPE IF EXISTS event_severity CASCADE;
DROP TYPE IF EXISTS runtime_type CASCADE;
DROP TYPE IF EXISTS conversation_mode CASCADE;
DROP TYPE IF EXISTS project_status CASCADE;
DROP TYPE IF EXISTS runner_status CASCADE;

-- Note: Extensions are not dropped as they may be used by other databases
