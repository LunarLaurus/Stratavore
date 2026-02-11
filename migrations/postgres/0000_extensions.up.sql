-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

-- Create custom types
CREATE TYPE runner_status AS ENUM ('starting', 'running', 'paused', 'terminated', 'failed');
CREATE TYPE project_status AS ENUM ('active', 'idle', 'archived');
CREATE TYPE conversation_mode AS ENUM ('new', 'continue', 'resume');
CREATE TYPE runtime_type AS ENUM ('process', 'container', 'remote');
CREATE TYPE event_severity AS ENUM ('debug', 'info', 'warning', 'error', 'critical');
