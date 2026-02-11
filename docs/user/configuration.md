# Configuration Guide

This guide covers Stratavore configuration in detail.

## Configuration Sources

Stratavore uses a hierarchical configuration system. Configuration values are loaded from multiple sources in order of precedence:

1. **Command line flags** (highest precedence)
2. **Environment variables** with `STRATAVORE_` prefix
3. **User config file**: `~/.config/stratavore/stratavore.yaml`
4. **System config file**: `/etc/stratavore/stratavore.yaml`
5. **Default values** (lowest precedence)

## Configuration File Structure

### Basic Configuration

```yaml
# ~/.config/stratavore/stratavore.yaml

# Database configuration
database:
  postgresql:
    host: localhost
    port: 5432
    database: stratavore_state
    user: stratavore
    password: "${STRATAVORE_DB_PASSWORD}"
    ssl_mode: prefer
    max_connections: 20
    connection_timeout: 30s

# Message queue configuration
messaging:
  rabbitmq:
    host: localhost
    port: 5672
    username: stratavore
    password: "${STRATAVORE_RABBITMQ_PASSWORD}"
    exchange: stratavore.events
    publisher_confirms: true

# Daemon settings
daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30
  max_concurrent_runners: 100

# Metrics and monitoring
metrics:
  prometheus:
    enabled: true
    port: 9091
    path: /metrics

# Logging configuration
logging:
  level: info
  format: json
  output: /var/log/stratavore/stratavore.log
```

### Production Configuration

```yaml
# Production-ready configuration
database:
  postgresql:
    host: postgres.internal.company.com
    port: 5432
    database: stratavore_prod
    user: stratavore_app
    password: "${STRATAVORE_DB_PASSWORD}"
    ssl_mode: require
    max_connections: 50
    connection_timeout: 10s
    query_timeout: 5s
    max_idle_time: 1h

messaging:
  rabbitmq:
    host: rabbitmq.internal.company.com
    port: 5672
    username: stratavore_app
    password: "${STRATAVORE_RABBITMQ_PASSWORD}"
    exchange: stratavore.events
    publisher_confirms: true
    connection_timeout: 10s
    heartbeat: 30s

daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30
  max_concurrent_runners: 500
  graceful_shutdown_timeout: 60s

metrics:
  prometheus:
    enabled: true
    port: 9091
    path: /metrics
    namespace: stratavore

security:
  mtls:
    enabled: true
    cert_file: /etc/stratavore/certs/server.crt
    key_file: /etc/stratavore/certs/server.key
    ca_file: /etc/stratavore/certs/ca.crt
    client_auth: require

logging:
  level: info
  format: json
  output: /var/log/stratavore/stratavore.log
  max_size: 100MB
  max_backups: 10
  max_age: 30d
  compress: true
```

## Configuration Sections

### Database Configuration

#### PostgreSQL Settings

```yaml
database:
  postgresql:
    # Connection settings
    host: localhost                # Database host
    port: 5432                   # Database port
    database: stratavore_state    # Database name
    user: stratavore             # Database user
    password: "password"          # Database password
    
    # SSL settings
    ssl_mode: require             # SSL mode (disable, allow, prefer, require)
    ssl_cert: ""                 # SSL certificate file
    ssl_key: ""                  # SSL key file
    ssl_ca: ""                   # SSL CA file
    
    # Connection pool settings
    max_connections: 20           # Maximum connections in pool
    min_connections: 5            # Minimum connections in pool
    max_idle_time: 1h            # Maximum idle time per connection
    max_lifetime: 24h            # Maximum lifetime per connection
    
    # Timeouts
    connection_timeout: 30s       # Connection timeout
    query_timeout: 10s           # Query timeout
    idle_timeout: 5m              # Idle timeout
    
    # Performance settings
    prepared_statements: true     # Use prepared statements
    binary_params: true          # Use binary parameters
```

#### SQLite Cache Settings

```yaml
database:
  sqlite:
    # Cache database for fast reads
    path: /var/lib/stratavore/cache.db
    max_size: 1GB                # Maximum cache size
    ttl: 1h                      # Cache TTL
    vacuum_interval: 24h         # Vacuum interval
```

### Messaging Configuration

#### RabbitMQ Settings

```yaml
messaging:
  rabbitmq:
    # Connection settings
    host: localhost
    port: 5672
    username: guest
    password: "guest"
    vhost: "/"
    
    # Exchange settings
    exchange: stratavore.events
    exchange_type: topic
    exchange_durable: true
    
    # Queue settings
    queue_durable: true
    queue_auto_delete: false
    
    # Publisher settings
    publisher_confirms: true
    publisher_timeout: 5s
    mandatory: false
    immediate: false
    
    # Consumer settings
    consumer_prefetch: 10
    consumer_timeout: 30s
    auto_ack: false
    
    # Connection settings
    connection_timeout: 30s
    heartbeat: 30s
    reconnect_delay: 5s
    max_reconnect_attempts: 10
    
    # SSL settings (optional)
    ssl_enabled: false
    ssl_cert: ""
    ssl_key: ""
    ssl_ca: ""
```

### Daemon Configuration

```yaml
daemon:
  # gRPC server settings
  grpc_port: 50051
  grpc_host: "0.0.0.0"
  grpc_max_concurrent_streams: 1000
  grpc_max_message_size: 4MB
  
  # Runner management
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30
  max_concurrent_runners: 100
  runner_timeout: 300s
  graceful_shutdown_timeout: 60s
  
  # Outbox settings
  outbox_batch_size: 50
  outbox_poll_interval: 5s
  outbox_max_attempts: 10
  outbox_backoff_base: 2s
```

### Metrics Configuration

```yaml
metrics:
  prometheus:
    enabled: true
    port: 9091
    path: /metrics
    namespace: stratavore
    subsystem: ""
    
    # Histogram settings
    histogram_buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
    
    # Registry settings
    enable_go_metrics: true
    enable_process_metrics: true
```

### Security Configuration

```yaml
security:
  # mTLS settings
  mtls:
    enabled: false
    cert_file: ""
    key_file: ""
    ca_file: ""
    client_auth: require           # none, request, require, verify_if_given, verify
  
  # Authentication settings
  auth:
    jwt_enabled: false
    jwt_secret: ""
    jwt_expiry: 24h
    api_keys_enabled: false
  
  # Agent join tokens
  agent_tokens:
    enabled: true
    expiry: 5m
    secret_length: 32
```

### Logging Configuration

```yaml
logging:
  level: info                    # debug, info, warn, error
  format: json                   # json, text
  output: /var/log/stratavore/stratavore.log
  
  # Log rotation
  max_size: 100MB
  max_backups: 10
  max_age: 30d
  compress: true
  
  # Structured logging
  enable_caller: true
  enable_stacktrace: true
  
  # Field configuration
  fields:
    timestamp: true
    level: true
    logger: true
    caller: true
    message: true
```

## Environment Variables

All configuration values can be overridden using environment variables with the `STRATAVORE_` prefix. The variable name follows the YAML path structure.

### Examples

```bash
# Database settings
export STRATAVORE_DATABASE_POSTGRESQL_HOST="db.internal"
export STRATAVORE_DATABASE_POSTGRESQL_PASSWORD="secure_password"
export STRATAVORE_DATABASE_POSTGRESQL_SSL_MODE="require"

# RabbitMQ settings
export STRATAVORE_MESSAGING_RABBITMQ_HOST="rabbitmq.internal"
export STRATAVORE_MESSAGING_RABBITMQ_PASSWORD="rabbitmq_password"

# Daemon settings
export STRATAVORE_DAEMON_GRPC_PORT="50051"
export STRATAVORE_DAEMON_HEARTBEAT_INTERVAL_SECONDS="15"

# Security settings
export STRATAVORE_SECURITY_MTLS_ENABLED="true"
export STRATAVORE_SECURITY_MTLS_CERT_FILE="/etc/certs/server.crt"

# Logging settings
export STRATAVORE_LOGGING_LEVEL="debug"
export STRATAVORE_LOGGING_FORMAT="text"
```

### Variable Naming Rules

- YAML path `database.postgresql.host` becomes `STRATAVORE_DATABASE_POSTGRESQL_HOST`
- YAML path `messaging.rabbitmq.publisher_confirms` becomes `STRATAVORE_MESSAGING_RABBITMQ_PUBLISHER_CONFIRMS`
- YAML path `daemon.grpc_port` becomes `STRATAVORE_DAEMON_GRPC_PORT`

## Command Line Flags

### Global Flags

```bash
--config string          Path to configuration file
--debug                  Enable debug logging
--god                    Enable god mode
--help                   Show help
--version                Show version
--timeout duration        Command timeout
--log-level string       Log level (debug, info, warn, error)
```

### Daemon-specific Flags

```bash
--foreground             Run in foreground (don't daemonize)
--grpc-port int         gRPC server port
--config-check          Validate configuration and exit
```

## Validation

### Configuration Validation

You can validate your configuration before starting the daemon:

```bash
# Validate default configuration
stratavore config validate

# Validate specific file
stratavore config validate --config /path/to/config.yaml

# Show current effective configuration
stratavore config show

# Show configuration with sources
stratavore config show --source
```

### Validation Rules

Stratavore validates configuration against these rules:

#### Database
- Host and port must be valid
- Database name must be provided
- Connection pool settings must be positive
- Timeouts must be valid durations

#### RabbitMQ
- Host and port must be valid
- Exchange name must be provided
- Connection settings must be valid
- SSL files must exist if SSL is enabled

#### Daemon
- gRPC port must be available
- Intervals must be positive
- Runner limits must be reasonable
- Timeouts must be valid durations

#### Security
- Certificate files must exist if mTLS is enabled
- Paths must be readable
- JWT secrets must be sufficiently long

## Configuration Examples

### Development Configuration

```yaml
# Development setup with local services
database:
  postgresql:
    host: localhost
    port: 5432
    database: stratavore_dev
    user: stratavore
    password: "dev_password"
    ssl_mode: disable

messaging:
  rabbitmq:
    host: localhost
    port: 5672
    username: guest
    password: "guest"
    exchange: stratavore.events

daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 5
  reconcile_interval_seconds: 15

logging:
  level: debug
  format: text
  output: stdout
```

### Cloud Deployment

```yaml
# Cloud deployment with managed services
database:
  postgresql:
    host: "your-aws-rds-instance.us-east-1.rds.amazonaws.com"
    port: 5432
    database: stratavore_prod
    user: stratavore_app
    password: "${STRATAVORE_DB_PASSWORD}"
    ssl_mode: require

messaging:
  rabbitmq:
    host: "your-mq-instance.us-east-1.mq.amazonaws.com"
    port: 5672
    username: stratavore_app
    password: "${STRATAVORE_MQ_PASSWORD}"
    exchange: stratavore.events

daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30

security:
  mtls:
    enabled: true
    cert_file: "/etc/stratavore/certs/server.crt"
    key_file: "/etc/stratavore/certs/server.key"
    ca_file: "/etc/stratavore/certs/ca.crt"

logging:
  level: info
  format: json
  output: stdout
```

### High Availability Setup

```yaml
# High availability configuration
database:
  postgresql:
    host: postgres-primary.internal
    port: 5432
    database: stratavore_prod
    user: stratavore_app
    password: "${STRATAVORE_DB_PASSWORD}"
    ssl_mode: require
    max_connections: 100
    
    # Read replicas
    read_replicas:
      - host: postgres-replica1.internal
        port: 5432
        weight: 1
      - host: postgres-replica2.internal
        port: 5432
        weight: 1

messaging:
  rabbitmq:
    hosts:
      - rabbitmq1.internal:5672
      - rabbitmq2.internal:5672
      - rabbitmq3.internal:5672
    username: stratavore_app
    password: "${STRATAVORE_MQ_PASSWORD}"
    exchange: stratavore.events
    publisher_confirms: true

daemon:
  grpc_port: 50051
  heartbeat_interval_seconds: 10
  reconcile_interval_seconds: 30
  max_concurrent_runners: 1000

metrics:
  prometheus:
    enabled: true
    port: 9091
    namespace: stratavore
    enable_go_metrics: true
```

## Troubleshooting Configuration

### Common Issues

#### Database Connection Failed
```bash
# Test connection manually
psql -h localhost -U stratavore -d stratavore_state

# Check configuration
stratavore config show --component database

# Enable debug logging
stratavore daemon start --log-level debug
```

#### RabbitMQ Connection Failed
```bash
# Test RabbitMQ manually
rabbitmq-diagnostics -q check_running

# Check configuration
stratavore config show --component messaging

# Test connectivity
telnet localhost 5672
```

#### TLS Certificate Issues
```bash
# Verify certificate chain
openssl verify -CAfile ca.crt server.crt

# Check certificate dates
openssl x509 -in server.crt -noout -dates

# Test certificate with server
openssl s_client -connect localhost:50051 -cert client.crt -key client.key -CAfile ca.crt
```

### Debug Configuration

```bash
# Show effective configuration
stratavore config show --source

# Validate specific component
stratavore config validate --component database

# Show configuration errors
stratavore config validate --verbose

# Test configuration loading
stratavore daemon --config-check --debug
```

---

For more information, see the [Deployment Guide](../operations/deployment.md) or [Development Guide](../developer/development.md).