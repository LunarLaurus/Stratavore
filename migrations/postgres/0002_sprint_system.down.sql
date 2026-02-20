-- Rollback sprint system migration

DROP TABLE IF EXISTS sprint_executions CASCADE;
DROP TABLE IF EXISTS sprint_tasks CASCADE;
DROP TABLE IF EXISTS sprints CASCADE;
DROP TABLE IF EXISTS model_registry CASCADE;
