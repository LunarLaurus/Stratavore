# Stratavore Architecture Overview: Python/HTML vs Core System

## Executive Summary

Stratavore is a hybrid system consisting of a **Go-based core orchestration platform** with **Python/HTML-based user interfaces and agent management**. The Python/HTML components serve as the interaction layer, while the Go core provides the heavy-lifting infrastructure.

---

## Python/HTML Components (User Interface Layer)

### Location
- `webui/` directory
- `agents/` directory (agent management)
- `.opencode/tools/` (development tools)

### Core Components

#### 1. Web UI (`webui/`)
**Purpose**: Browser-based monitoring and control interface

**Files**:
- `index.html` (774 lines) - Complete single-page application with:
  - Real-time job tracking dashboard
  - Agent status monitoring
  - Time tracking visualization
  - Manual agent control panel
  - Debug and health monitoring tools

- `server.py` (410 lines) - HTTP server providing:
  - REST API endpoints (`/api/status`, `/api/spawn-agent`, etc.)
  - JSON data serving from local files
  - CORS-enabled frontend communication
  - Agent management operations

#### 2. Agent Management (`agents/`)
**Purpose**: Multi-agent workflow orchestration

**Files**:
- `agent_manager.py` - Core agent lifecycle management
  - Agent spawning with personalities (cadet, senior, specialist, etc.)
  - Task assignment and status tracking
  - Thread-safe file-based persistence
  - Agent state management (idle, working, paused, error)

#### 3. Development Tools (`.opencode/tools/`)
**Purpose**: Development and integration utilities

**Files**:
- `stratavore_wrapper.py` - Claude Code integration
- `meridian_lex_stratavore.py` - Personality system
- Various test and validation scripts

### Technology Stack
- **Python 3.11+** for backend logic
- **HTML5/CSS3/JavaScript** for frontend
- **File-based storage** (JSONL format)
- **HTTP/REST** communication
- **Thread-based concurrency**

---

## Go Core System (Infrastructure Layer)

### Location
- `cmd/` - Entry points
- `internal/` - Core business logic
- `pkg/` - Public APIs

### Core Components

#### 1. CLI Tools (`cmd/`)
- `stratavore/` - Main client application
- `stratavore-agent/` - Agent wrapper
- `stratavored/` - Daemon/server

#### 2. Daemon (`internal/daemon/`)
- HTTP/gRPC servers
- Runner lifecycle management
- Session orchestration
- Real-time coordination

#### 3. Storage (`internal/storage/`)
- PostgreSQL integration with pgvector
- Transactional outbox pattern
- Persistent state management
- Data consistency guarantees

#### 4. Messaging (`internal/messaging/`)
- RabbitMQ event distribution
- Outbox-based reliable delivery
- Event-driven architecture
- Async coordination

#### 5. Observability (`internal/observability/`)
- Prometheus metrics
- Structured logging
- Health monitoring
- Performance tracking

### Technology Stack
- **Go 1.24+** for high-performance services
- **PostgreSQL + pgvector** for data persistence
- **RabbitMQ** for message queuing
- **gRPC/HTTP** for service communication
- **Prometheus** for metrics
- **Docker** for containerization

---

## System Integration

### Communication Patterns

1. **Python → Go**: HTTP/gRPC API calls
2. **Go → Python**: File-based state updates
3. **Browser ↔ Python**: WebSocket-like polling
4. **Agents ↔ Core**: Event-driven messaging

### Data Flow

```
Browser UI → Python Server → Go Daemon → Infrastructure
     ↑              ↓              ↓              ↓
   Display ← JSON Files ← PostgreSQL ← RabbitMQ
```

### Key Integration Points

1. **State Synchronization**
   - Python reads from JSONL files
   - Go writes to PostgreSQL
   - Bridge processes maintain consistency

2. **Agent Lifecycle**
   - Python `agent_manager.py` handles local agents
   - Go daemon manages system-level runners
   - Both coordinate through shared storage

3. **User Interface**
   - Python serves the web interface
   - Go provides backend APIs
   - Real-time updates via polling

---

## Architectural Benefits

### Python/HTML Layer
- **Rapid Development**: Quick UI prototyping
- **Flexibility**: Easy to modify interfaces
- **Accessibility**: Browser-based access
- **Agent Management**: Complex workflow orchestration

### Go Core Layer
- **Performance**: High-throughput operations
- **Reliability**: Strong typing and error handling
- **Scalability**: Efficient resource utilization
- **Production-Ready**: Enterprise-grade infrastructure

### Separation of Concerns
- **UI Logic**: Isolated in Python/HTML
- **Business Logic**: Centralized in Go
- **Data Persistence**: Robust Go-based storage
- **User Experience**: Rich web interface

---

## Deployment Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Browser UI    │←── │  Python Server │←── │   Go Daemon    │
│   (HTML/JS)     │    │   (server.py)  │    │  (stratavored) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ↓                        ↓
                       ┌─────────────────┐    ┌─────────────────┐
                       │  File Storage   │    │  PostgreSQL     │
                       │  (JSONL files)  │    │  + pgvector     │
                       └─────────────────┘    └─────────────────┘
                                                        │
                                                        ↓
                                               ┌─────────────────┐
                                               │   RabbitMQ      │
                                               │   (Events)      │
                                               └─────────────────┘
```

---

## Conclusion

The Python/HTML components provide the **human-computer interface** and **agent workflow management**, making Stratavore accessible and user-friendly. The Go core delivers the **industrial-strength infrastructure** needed for production workloads, with robust data persistence, messaging, and observability.

This hybrid architecture allows for rapid UI development while maintaining enterprise-grade performance and reliability in the core system.