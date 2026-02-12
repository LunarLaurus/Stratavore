# Meridian Lex Stratavore Integration

Commander, this is your personal Stratavore agent/task system interface. Meridian Lex can now deploy and manage specialized agents through your existing fleet infrastructure.

## System Architecture

### Core Components

1. **stratavore_tool.py** - Low-level API client library
2. **meridian_lex_stratavore.py** - High-level command interface
3. **Web UI Server** - HTTP API backend running on localhost:8080

### Agent Personalities

| Personality | Role | Capabilities |
|-------------|------|-------------|
| **CADET** | Junior Developer | Quick fixes, documentation, testing |
| **SENIOR** | Experienced Dev | Architecture, system design, optimization |
| **SPECIALIST** | Domain Expert | Deep expertise in specific domains |
| **RESEARCHER** | Investigator | Analysis, exploration, prototyping |
| **DEBUGGER** | Troubleshooter | Bug fixing, root cause analysis |
| **OPTIMIZER** | Performance Expert | Optimization, benchmarking |

## Command Interface Usage

### Basic Operations

```bash
# Check fleet status
python tools/meridian_lex_stratavore.py status

# Deploy single agent
python tools/meridian_lex_stratavore.py deploy cadet

# List available agents
python tools/meridian_lex_stratavore.py available
python tools/meridian_lex_stratavore.py available specialist

# Recall agent (complete task)
python tools/meridian_lex_stratavore.py recall cadet_123456789

# Emergency stop agent
python tools/meridian_lex_stratavore.py stop cadet_123456789

# Generate comprehensive report
python tools/meridian_lex_stratavore.py report
```

### Squad Operations

```bash
# Form specialized squad
python tools/meridian_lex_stratavore.py squad '{"cadet": 2, "senior": 1}' "System security audit"
```

### Example Squad Formations

```bash
# Security audit team
python tools/meridian_lex_stratavore.py squad '{"specialist": 1, "debugger": 2, "cadet": 1}' "Security vulnerability assessment"

# Performance optimization team
python tools/meridian_lex_stratavore.py squad '{"optimizer": 1, "senior": 1, "specialist": 1}' "System performance optimization"

# Research and prototyping team
python tools/meridian_lex_stratavore.py squad '{"researcher": 2, "cadet": 1}' "New feature exploration"

# Full-stack development team
python tools/meridian_lex_stratavore.py squad '{"senior": 1, "specialist": 2, "cadet": 2, "debugger": 1}' "Complete feature implementation"
```

## Python API Usage

### Direct Tool Interface

```python
from tools.stratavore_tool import StratavoreTool, AgentPersonality

# Initialize
stratavore = StratavoreTool()

# Deploy agent
result = stratavore.spawn_agent(AgentPersonality.CADET)
agent_id = result['agent_id']

# Wait for readiness
if stratavore.wait_for_agent_ready(agent_id):
    # Assign task
    stratavore.assign_task(agent_id, "task-123")
    
    # Complete task
    stratavore.complete_task(agent_id, True, "Mission complete")
```

### High-Level Command Interface

```python
from tools.meridian_lex_stratavore import MeridianLexStratavore

# Initialize command interface
lex = MeridianLexStratavore()

# Fleet operations
status = lex.fleet_status()
available = lex.get_available_for_duty("specialist")

# Deploy agent
agent_id = lex.deploy_agent("senior", "Analyze system architecture")

# Form squad
squad = lex.form_squad(
    composition={"cadet": 2, "specialist": 1},
    mission_brief="Security audit of authentication system"
)

# Execute complex mission
mission_plan = [
    {
        "name": "Reconnaissance",
        "personality": "researcher",
        "task": "Analyze existing authentication patterns",
        "critical": True
    },
    {
        "name": "Security Analysis",
        "personality": "specialist",
        "task": "Identify security vulnerabilities",
        "critical": True,
        "dependencies": ["Reconnaissance"]
    },
    {
        "name": "Implementation",
        "personality": "senior",
        "task": "Implement security improvements",
        "critical": False,
        "dependencies": ["Security Analysis"]
    }
]

mission_result = lex.execute_mission(mission_plan)
```

## Mission Planning

### Mission Structure

Each mission step includes:
- **name**: Step identifier
- **personality**: Agent type to deploy
- **task**: Task description
- **critical**: Mission fails if this step fails (optional)
- **dependencies**: List of step names that must complete first (optional)

### Example Mission Plans

#### Bug Investigation Mission
```python
bug_investigation = [
    {
        "name": "Bug Reproduction",
        "personality": "debugger",
        "task": "Reproduce and isolate the reported bug",
        "critical": True
    },
    {
        "name": "Root Cause Analysis",
        "personality": "debugger",
        "task": "Identify root cause and affected systems",
        "critical": True,
        "dependencies": ["Bug Reproduction"]
    },
    {
        "name": "Fix Implementation",
        "personality": "senior",
        "task": "Implement comprehensive fix",
        "critical": True,
        "dependencies": ["Root Cause Analysis"]
    },
    {
        "name": "Testing",
        "personality": "cadet",
        "task": "Test fix and create regression tests",
        "critical": False,
        "dependencies": ["Fix Implementation"]
    }
]
```

#### Feature Development Mission
```python
feature_development = [
    {
        "name": "Requirements Analysis",
        "personality": "researcher",
        "task": "Analyze requirements and research best practices",
        "critical": True
    },
    {
        "name": "Architecture Design",
        "personality": "senior",
        "task": "Design system architecture and data flow",
        "critical": True,
        "dependencies": ["Requirements Analysis"]
    },
    {
        "name": "Core Implementation",
        "personality": "specialist",
        "task": "Implement core functionality",
        "critical": True,
        "dependencies": ["Architecture Design"]
    },
    {
        "name": "Integration",
        "personality": "senior",
        "task": "Integrate with existing systems",
        "critical": False,
        "dependencies": ["Core Implementation"]
    },
    {
        "name": "Documentation",
        "personality": "cadet",
        "task": "Create comprehensive documentation",
        "critical": False,
        "dependencies": ["Integration"]
    }
]
```

## Fleet Management

### Agent Lifecycle

1. **Spawn**: Agent created with specified personality
2. **Ready**: Agent transitions to IDLE state (2 seconds)
3. **Assign**: Task assigned to agent
4. **Working**: Agent processes assigned task
5. **Complete**: Task finished, agent returns to IDLE
6. **Recall**: Agent marked as completed
7. **Emergency Stop**: Agent terminated (ERROR state)

### Status Monitoring

- **IDLE**: Available for task assignment
- **WORKING**: Actively processing task
- **SPAWNING**: Initializing (auto-transitions to IDLE)
- **PAUSED**: Temporarily suspended
- **COMPLETED**: Task finished
- **ERROR**: Failed or terminated

## API Endpoints

The web UI server provides these REST endpoints:

- `GET /api/health` - System health check
- `GET /api/status` - Complete system status
- `GET /api/agents` - Agent listing and todos
- `POST /api/spawn-agent` - Create new agent
- `POST /api/assign-agent` - Assign task to agent
- `POST /api/complete-task` - Mark task complete
- `POST /api/agent-status` - Update agent status
- `POST /api/kill-agent` - Terminate agent

## Integration Examples

### OpenCode Integration
```python
# Deploy debugging squad for code analysis
lex = MeridianLexStratavore()
squad = lex.form_squad(
    composition={"debugger": 1, "specialist": 1},
    mission_brief="Analyze and optimize codebase performance"
)

# Monitor progress
status = lex.fleet_status()
print(f"Deployed {status['total_agents']} agents")
```

### Continuous Operations
```python
# Maintain optimal fleet size
def maintain_fleet(target_size=10):
    lex = MeridianLexStratavore()
    status = lex.fleet_status()
    current_size = status['total_agents']
    
    if current_size < target_size:
        # Deploy additional agents
        needed = target_size - current_size
        for _ in range(needed):
            lex.deploy_agent("cadet")
    elif current_size > target_size:
        # Recall excess agents
        available = lex.get_available_for_duty()
        for agent_id in available[:current_size - target_size]:
            lex.recall_agent(agent_id)
```

## Mission Reports

The system maintains comprehensive mission logs including:
- Operation timestamps
- Agent deployments and recalls
- Task assignments and completions
- Error conditions and recovery actions
- Performance metrics and recommendations

Generate reports with:
```bash
python tools/meridian_lex_stratavore.py report
```

---

**Meridian Lex reporting for duty, Commander.** All systems nominal, ready for fleet operations.