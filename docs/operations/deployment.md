# Deployment Guide

This guide covers deploying Stratavore to production environments.

## Table of Contents

1. [Deployment Options](#deployment-options)
2. [Prerequisites](#prerequisites)
3. [Production Configuration](#production-configuration)
4. [Database Setup](#database-setup)
5. [Message Queue Setup](#message-queue-setup)
6. [Security Configuration](#security-configuration)
7. [High Availability](#high-availability)
8. [Monitoring and Observability](#monitoring-and-observability)
9. [Backup and Recovery](#backup-and-recovery)
10. [Scaling Considerations](#scaling-considerations)

## Deployment Options

### Option 1: Single Node Deployment
Suitable for small teams and development environments.

### Option 2: High Availability Deployment
Recommended for production use with redundancy.

### Option 3: Cloud-Native Deployment
Deploy on Kubernetes or cloud platforms.

### Option 4: Docker Compose Deployment
Simple deployment using Docker Compose.

## Prerequisites

### System Requirements

**Minimum (Single Node):**
- CPU: 2 cores
- Memory: 4GB RAM
- Storage: 20GB SSD
- Network: 100 Mbps

**Recommended (Production):**
- CPU: 4+ cores
- Memory: 8GB+ RAM
- Storage: 100GB+ SSD
- Network: 1 Gbps

### Software Requirements
- PostgreSQL 14+ with pgvector extension
- RabbitMQ 3.12+
- Go 1.22+ (for building from source)
- Docker and Docker Compose (for containerized deployment)
- SSL certificates (for production)

## Production Configuration

### Configuration File
Create `/etc/stratavore/stratavore.yaml`:

```yaml
# Production configuration
database:
  postgresql:
    host: postgres.internal
    port: 5432
    database: stratavore_prod
    user: stratavore
    password: "${STRATAVORE_DB_PASSWORD}"
    ssl_mode: require
    max_connections: 20
    connection_timeout: 30s

messaging:
  rabbitmq:
    host: rabbitmq.internal
    port: 5672
    username: stratavore
    password: "${STRATAVORE_RABBITMQ_PASSWORD}"
    exchange: stratavore.events
    publisher_confirms: true
    ssl_enabled: true

daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30
  max_concurrent_runners: 100

metrics:
  prometheus:
    enabled: true
    port: 9091
    path: /metrics

security:
  mtls:
    enabled: true
    cert_file: /etc/stratavore/certs/server.crt
    key_file: /etc/stratavore/certs/server.key
    ca_file: /etc/stratavore/certs/ca.crt

logging:
  level: info
  format: json
  output: /var/log/stratavore/stratavore.log
```

### Environment Variables
```bash
# Database credentials
export STRATAVORE_DB_PASSWORD="your_secure_password"

# RabbitMQ credentials
export STRATAVORE_RABBITMQ_PASSWORD="your_rabbitmq_password"

# TLS certificates
export STRATAVORE_TLS_CERT_FILE="/etc/stratavore/certs/server.crt"
export STRATAVORE_TLS_KEY_FILE="/etc/stratavore/certs/server.key"
```

## Database Setup

### PostgreSQL Configuration

**postgresql.conf optimizations:**
```ini
# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connection settings
max_connections = 200
listen_addresses = '*'

# WAL settings for replication
wal_level = replica
max_wal_senders = 3
wal_keep_segments = 32

# Performance settings
random_page_cost = 1.1
effective_io_concurrency = 200
```

**pg_hba.conf for secure access:**
```
# Local connections
local   all             postgres                                peer
local   all             stratavore                              peer

# Remote connections (require SSL)
hostssl all             stratavore      10.0.0.0/8               md5
hostssl replication     replicator      10.0.0.0/8               md5
```

### Database Creation
```bash
# Create user and database
sudo -u postgres createuser stratavore
sudo -u postgres createdb -O stratavore stratavore_prod

# Enable pgvector extension
sudo -u postgres psql -d stratavore_prod -c "CREATE EXTENSION IF NOT EXISTS vector;"

# Run migrations
stratavore-migrate up --database-url "postgres://stratavore:password@postgres.internal:5432/stratavore_prod"
```

### Performance Optimization
```sql
-- Create performance indexes
CREATE INDEX CONCURRENTLY idx_runners_status_project ON runners(status, project_name) 
WHERE status IN ('running', 'starting', 'paused');

CREATE INDEX CONCURRENTLY idx_sessions_project_created ON sessions(project_name, created_at DESC);

CREATE INDEX CONCURRENTLY idx_events_timestamp_entity ON events(timestamp DESC, entity_type, entity_id);

-- Update table statistics
ANALYZE runners;
ANALYZE sessions;
ANALYZE events;
```

## Message Queue Setup

### RabbitMQ Configuration

**rabbitmq.conf:**
```ini
# Basic settings
listeners.ssl.default = 5671
ssl_options.cacertfile = /etc/rabbitmq/certs/ca.crt
ssl_options.certfile   = /etc/rabbitmq/certs/server.crt
ssl_options.keyfile    = /etc/rabbitmq/certs/server.key

# Memory limits
vm_memory_high_watermark.relative = 0.6

# Disk space limits
disk_free_limit.absolute = 1GB

# Cluster settings
cluster_formation.peer_discovery_backend = classic_config
cluster_formation.classic_config.nodes.1 = rabbit@node1
cluster_formation.classic_config.nodes.2 = rabbit@node2
```

**Setup exchanges and queues:**
```bash
# Enable management plugin
rabbitmq-plugins enable rabbitmq_management

# Create user
rabbitmqctl add_user stratavore your_password
rabbitmqctl set_user_tags stratavore administrator
rabbitmqctl set_permissions -p / stratavore ".*" ".*" ".*"

# Define exchange and queues
rabbitmqadmin declare exchange name=stratavore.events type=topic durable=true
rabbitmqadmin declare queue name=stratavore.daemon.events durable=true
rabbitmqadmin declare queue name=stratavore.metrics durable=true
rabbitmqadmin declare queue name=stratavore.alerts durable=true

# Bind queues
rabbitmqadmin declare binding source=stratavore.events destination=stratavore.daemon.events routing_key="#"
rabbitmqadmin declare binding source=stratavore.events destination=stratavore.metrics routing_key="metrics.*"
rabbitmqadmin declare binding source=stratavore.events destination=stratavore.alerts routing_key="system.alert.*"
```

## Security Configuration

### TLS/SSL Setup

**Generate certificates (using Let's Encrypt):**
```bash
# Install certbot
sudo apt-get install certbot

# Generate certificates
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates to Stratavore
sudo mkdir -p /etc/stratavore/certs
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem /etc/stratavore/certs/server.crt
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem /etc/stratavore/certs/server.key
```

**Or use self-signed certificates:**
```bash
# Generate CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# Generate server certificate
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
```

### Firewall Configuration
```bash
# Allow gRPC traffic (50051)
sudo ufw allow 50051/tcp

# Allow Prometheus metrics (9091)
sudo ufw allow 9091/tcp

# Allow PostgreSQL (5432) - restrict to internal network
sudo ufw allow from 10.0.0.0/8 to any port 5432

# Allow RabbitMQ (5672, 15672) - restrict to internal network
sudo ufw allow from 10.0.0.0/8 to any port 5672
sudo ufw allow from 10.0.0.0/8 to any port 15672
```

### Systemd Service

**Create `/etc/systemd/system/stratavored.service`:**
```ini
[Unit]
Description=Stratavore Daemon
After=network.target postgresql.service rabbitmq.service
Wants=postgresql.service rabbitmq.service

[Service]
Type=simple
User=stratavore
Group=stratavore
ExecStart=/usr/local/bin/stratavored --config /etc/stratavore/stratavore.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=stratavore

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/stratavore /var/lib/stratavore

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

**Enable and start service:**
```bash
# Create user
sudo useradd -r -s /bin/false stratavore
sudo mkdir -p /var/log/stratavore /var/lib/stratavore
sudo chown stratavore:stratavore /var/log/stratavore /var/lib/stratavore

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable stratavored
sudo systemctl start stratavored
```

## High Availability

### Database Replication

**Primary PostgreSQL configuration:**
```ini
# In postgresql.conf
wal_level = replica
max_wal_senders = 3
wal_keep_segments = 64
archive_mode = on
archive_command = 'cp %p /var/lib/postgresql/archive/%f'
```

**Replica setup:**
```bash
# On replica server
pg_basebackup -h primary-ip -U replicator -D /var/lib/postgresql/data -v -P -W

# Configure recovery.conf
standby_mode = 'on'
primary_conninfo = 'host=primary-ip port=5432 user=replicator'
```

### RabbitMQ Cluster

**Cluster nodes configuration:**
```ini
# On each node
RABBITMQ_NODENAME=rabbit@node1  # or node2, node3
RABBITMQ_ERLANG_COOKIE=same_cookie_value_on_all_nodes
```

**Form cluster:**
```bash
# On node2 and node3
rabbitmqctl stop_app
rabbitmqctl reset
rabbitmqctl join_cluster rabbit@node1
rabbitmqctl start_app

# Enable mirrored queues
rabbitmqctl set_policy ha-all "^stratavore\." '{"ha-mode":"all","ha-sync-mode":"automatic"}'
```

### Load Balancing

**HAProxy configuration for gRPC:**
```
frontend stratavore_grpc
    bind *:50051
    mode tcp
    default_backend stratavore_daemons

backend stratavore_daemons
    mode tcp
    balance roundrobin
    option tcp-check
    server daemon1 10.0.1.10:50051 check
    server daemon2 10.0.1.11:50051 check
    server daemon3 10.0.1.12:50051 check
```

## Monitoring and Observability

### Prometheus Configuration

**prometheus.yml:**
```yaml
scrape_configs:
  - job_name: 'stratavore'
    static_configs:
      - targets: ['localhost:9091', '10.0.1.11:9091', '10.0.1.12:9091']
    scrape_interval: 15s
    metrics_path: /metrics

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'rabbitmq'
    static_configs:
      - targets: ['rabbitmq-exporter:9419']
```

### Grafana Dashboards

**Key metrics to monitor:**
- Runner count and status
- Token usage rates
- Database connection pool usage
- RabbitMQ queue depths
- System resource usage
- Request latency and error rates

### Alerting Rules

**Prometheus alerts.yml:**
```yaml
groups:
  - name: stratavore.rules
    rules:
      - alert: HighRunnerFailureRate
        expr: rate(stratavore_runner_failures_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High runner failure rate detected"

      - alert: DatabaseConnectionHigh
        expr: stratavore_db_connections_active / stratavore_db_connections_max > 0.8
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted"
```

## Backup and Recovery

### Database Backup

**Automated backup script:**
```bash
#!/bin/bash
# /usr/local/bin/backup-stratavore.sh

BACKUP_DIR="/var/backups/stratavore"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="stratavore_prod"

# Create backup directory
mkdir -p $BACKUP_DIR

# Database backup
pg_dump -h localhost -U stratavore -d $DB_NAME -Fc > $BACKUP_DIR/stratavore_$DATE.dump

# Compress old backups
find $BACKUP_DIR -name "*.dump" -mtime +7 -exec gzip {} \;

# Remove backups older than 30 days
find $BACKUP_DIR -name "*.dump.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/stratavore_$DATE.dump s3://your-backup-bucket/stratavore/
```

**Configure cron job:**
```bash
# Add to crontab
0 2 * * * /usr/local/bin/backup-stratavore.sh
```

### Point-in-Time Recovery

**Enable WAL archiving:**
```ini
# In postgresql.conf
archive_mode = on
archive_command = 'rsync -a %p /var/lib/postgresql/wal_archive/%f'
```

**Recovery procedure:**
```bash
# Stop PostgreSQL
sudo systemctl stop postgresql

# Restore base backup
pg_restore -h localhost -U stratavore -d stratavore_prod /path/to/backup.dump

# Configure recovery
echo "restore_command = 'cp /var/lib/postgresql/wal_archive/%f %p'" >> /var/lib/postgresql/data/recovery.conf
echo "recovery_target_time = '2024-01-15 14:30:00'" >> /var/lib/postgresql/data/recovery.conf

# Start PostgreSQL
sudo systemctl start postgresql
```

## Scaling Considerations

### Horizontal Scaling

**Multiple daemon instances:**
- Deploy 3+ daemon instances behind load balancer
- Each daemon handles ~1000 concurrent runners
- Use database for coordination and state management

**Database scaling:**
- Use read replicas for reporting queries
- Partition large tables by project or time
- Consider sharding for very large deployments

**Message queue scaling:**
- Use RabbitMQ cluster with mirrored queues
- Separate high-volume queues to dedicated nodes
- Consider Kafka for very high throughput

### Performance Tuning

**Database optimization:**
```sql
-- Connection pool sizing
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';

-- Query optimization
EXPLAIN ANALYZE SELECT * FROM runners WHERE project_name = 'project' AND status = 'running';
```

**Application tuning:**
```yaml
# In stratavore.yaml
daemon:
  max_concurrent_runners: 1000
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30

database:
  postgresql:
    max_connections: 50
    connection_timeout: 30s
    query_timeout: 10s
```

### Capacity Planning

**Resource usage per runner:**
- Memory: ~10-50MB (varies with session size)
- CPU: Minimal (mostly idle)
- Database: ~1KB for runner record
- Network: ~1KB per heartbeat

**Scaling thresholds:**
- 100 concurrent runners: Single node sufficient
- 500 concurrent runners: Consider multiple daemons
- 1000+ concurrent runners: Full HA deployment needed

---

For more information, see the [Monitoring Guide](monitoring.md) or [Troubleshooting Guide](troubleshooting.md).