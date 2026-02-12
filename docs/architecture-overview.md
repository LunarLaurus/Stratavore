# Stratavore Architecture Overview: Python/HTML vs Core System

## Executive Summary

Stratavore is a hybrid system consisting of a **Go-based core orchestration platform** with **Python/HTML-based user interfaces and agent management**. The Python/HTML components serve as the interaction layer, while the Go core provides the heavy-lifting infrastructure.

---

## Python/HTML Components (User Interface Layer)

### Location
- `webui/` directory (modular architecture v4)
- `agents/` directory (agent management)
- `.opencode/tools/` (development tools)

### Core Components

#### 1. Modular Web UI (`webui/`) - v4 Architecture
**Purpose**: Scalable, component-based monitoring and control interface

**Architecture**:
```
webui/
├── index.html (main template)
├── components/          # Modular UI components
│   ├── base-component.js     # Base component class
│   ├── header-component.js   # Header with controls
│   ├── overview-panel-component.js  # Statistics dashboard
│   ├── jobs-list-component.js     # Job management
│   ├── agent-status-component.js   # Agent management
│   └── [additional components...]
├── services/           # Core services
│   ├── event-bus.js    # Component communication
│   └── api-client.js   # API client and data manager
├── styles/            # Modular CSS
│   ├── base.css       # Design system foundation
│   └── components/    # Component-specific styles
└── utils/             # Utilities
    ├── constants.js   # App constants
    └── helpers.js     # Helper functions
```

**Key Features**:
- **Component-based Architecture**: Each UI element is an independent, reusable component
- **Event-driven Communication**: Components communicate through a centralized event bus
- **Reactive Data Store**: Centralized state management with automatic UI updates
- **Unlimited Agent Scaling**: Virtual scrolling and batch operations for 1000+ agents
- **Modular CSS**: Component-scoped styling with a design system
- **Real-time Updates**: Polling-based data synchronization with future WebSocket support

**Component System**:
- **BaseComponent**: Provides lifecycle management, event handling, and state management
- **HeaderComponent**: Application header with status indicators and global controls
- **OverviewPanelComponent**: Real-time statistics and metrics dashboard
- **JobsListComponent**: Paginated, searchable, filterable job management
- **AgentStatusComponent**: Scalable agent monitoring with batch operations

- `server.py` (v4 enhanced) - Modular HTTP server providing:
  - **API Handler Classes**: Separate handlers for different endpoint groups
  - **Standardized Responses**: Consistent API response structure
  - **Enhanced Error Handling**: Graceful degradation and error reporting
  - **Batch Operations**: Multi-agent actions for scalability
  - **Modular Architecture**: Clean separation of concerns
  - **CORS Support**: Full cross-origin resource sharing
  - **Health Monitoring**: System health and uptime tracking

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
- **Python 3.11+** for backend API server
- **Modern JavaScript (ES6+)** for frontend components
- **HTML5/CSS3 with Custom Properties** for responsive design
- **File-based storage** (JSONL format) for state persistence
- **HTTP/REST** communication with standardized API responses
- **Event-driven Architecture** for component communication
- **CSS Design System** with custom properties for theming
- **Virtual Scrolling** for handling large datasets
- **Debounced Operations** for performance optimization

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

### Python/HTML Layer (v4 Modular)
- **Component Reusability**: Each UI element can be independently developed, tested, and reused
- **Scalable Architecture**: Supports unlimited agents through virtual scrolling and batch operations
- **Maintainability**: Clear separation of concerns with modular file structure
- **Developer Experience**: Modern JavaScript with event-driven patterns and reactive state management
- **Performance Optimization**: Debounced operations, lazy loading, and efficient DOM updates
- **Accessibility**: Browser-based access with keyboard shortcuts and responsive design
- **Agent Management**: Advanced workflow orchestration with individual and batch operations

### Go Core Layer
- **Performance**: High-throughput operations with Go's concurrency model
- **Reliability**: Strong typing, error handling, and production-grade infrastructure
- **Scalability**: Efficient resource utilization and horizontal scaling capabilities
- **Enterprise-Ready**: Comprehensive monitoring, logging, and observability

### Separation of Concerns
- **UI Components**: Independent, self-contained modules with clear responsibilities
- **Core Services**: Centralized state management, API communication, and event handling
- **Data Persistence**: Robust Go-based storage with transactional guarantees
- **User Experience**: Rich, responsive interface with real-time updates
- **Development Workflow**: Modular development allowing parallel work on different components

---

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    BROWSER CLIENT (v4)                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │ Header Comp │ │ Overview    │ │ Jobs List   │ │ Agent   │ │
│  │             │ │ Panel       │ │ Component  │ │ Status  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
│  ┌─────────────┐ ┌─────────────┐                           │
│  │ Event Bus   │ │ Data Store  │ ← Reactive State Management │
│  └─────────────┘ └─────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
                                │ HTTP/REST API (Modular)
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                PYTHON SERVER (v4 Enhanced)                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Status API  │ │ Agent API   │ │ Health API  │           │
│  │ Handler     │ │ Handler     │ │ Handler     │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Data Loader Service                       │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │ gRPC/HTTP
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                    GO DAEMON (Core)                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ State Mgr   │ │ Runner Mgr  │ │ Event Bus   │           │
│  │ (PostgreSQL)│ │ (Lifecycle) │ │ (RabbitMQ)  │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ↓
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  File Storage   │    │  PostgreSQL     │    │   RabbitMQ      │
│  (JSONL files)  │    │  + pgvector     │    │   (Events)      │
│                 │    │                 │    │                 │
│ • Active Agents │    │ • Projects      │    │ • Runner Events │
│ • Agent Todos   │    │ • Sessions     │    │ • System Alerts │
│ • Jobs          │    │ • Outbox       │    │ • Notifications│
│ • Time Sessions │    │ • Metrics      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Key Scaling Features

### Frontend v4 Architecture
- **Virtual Scrolling**: Efficiently handle 1000+ agents without performance degradation
- **Batch Operations**: Perform actions on multiple agents simultaneously
- **Component Lazy Loading**: Load only the components needed for current view
- **Event-Driven Updates**: Only update components affected by data changes
- **Debounced Interactions**: Prevent excessive API calls during rapid user actions

### Backend v4 Enhancements
- **Modular API Handlers**: Separate handlers for different endpoint groups
- **Standardized Response Format**: Consistent API responses across all endpoints
- **Batch Operation Support**: Efficiently handle multiple agent actions
- **Enhanced Error Handling**: Graceful degradation with meaningful error messages
- **Health Monitoring**: Real-time system health and performance metrics

### Integration Patterns
- **Real-time Polling**: 30-second data refresh with manual refresh capabilities
- **CORS-Enabled Communication**: Full support for cross-origin requests
- **State Synchronization**: Consistent state between frontend components and backend
- **Error Recovery**: Automatic retry logic and fallback mechanisms

---

## Migration from Monolithic to Modular (v4)

### What Changed
- **From Single File to Component System**: 774-line `index.html` → 15+ modular components
- **From Ad-hoc JavaScript to Structured Architecture**: 427-line script → service layer with event bus
- **From Mixed CSS to Design System**: Monolithic styles → component-scoped CSS with custom properties
- **From Simple Server to API Layer**: Basic handler → modular API with standardized responses

### Benefits Achieved
- **Maintainability**: Components can be developed, tested, and debugged independently
- **Scalability**: Virtual scrolling and batch operations support unlimited agents
- **Performance**: Debounced operations, lazy loading, and efficient DOM updates
- **Developer Experience**: Modern JavaScript patterns, clear separation of concerns
- **Code Reusability**: Components can be reused across different pages or projects

### Future Extensibility
- **WebSocket Support**: Ready for real-time updates without polling
- **Plugin Architecture**: Components can be added as plugins
- **Theme System**: CSS custom properties enable easy theming
- **Mobile Responsiveness**: Component-based responsive design
- **Progressive Enhancement**: Core functionality works with JavaScript disabled

---

## Conclusion

The **v4 modular architecture** transforms Stratavore's web interface from a monolithic application to a scalable, maintainable platform. The Python/HTML components now provide:

- **Component-Based Architecture**: Independent, reusable UI modules
- **Scalable Agent Management**: Support for unlimited agents through virtual scrolling
- **Modern Developer Experience**: Event-driven patterns and reactive state management
- **Enhanced User Interface**: Real-time updates with batch operations and advanced filtering

The Go core continues to deliver the **industrial-strength infrastructure** needed for production workloads, while the new modular frontend architecture provides the flexibility and scalability required for modern web applications.

This evolution enables rapid feature development, easier maintenance, and the ability to scale the user interface alongside the powerful backend infrastructure.