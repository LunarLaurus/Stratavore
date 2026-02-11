#!/usr/bin/env python3
"""
Agent Manager - Handles multiple parallel agents with different personalities
Manages agent spawning, task assignment, and status monitoring
"""

import json
import time
import sys
import subprocess
import threading
import uuid
from datetime import datetime
from typing import Dict, List, Optional, Union
from enum import Enum

class AgentStatus(Enum):
    IDLE = "idle"
    WORKING = "working"
    PAUSED = "paused"
    COMPLETED = "completed"
    ERROR = "error"
    SPAWNING = "spawning"

class AgentPersonality(Enum):
    CADET = "cadet"
    SENIOR = "senior"
    SPECIALIST = "specialist"
    RESEARCHER = "researcher"
    DEBUGGER = "debugger"
    OPTIMIZER = "optimizer"

class AgentManager:
    def __init__(self):
        self.agents_file = "agents/active_agents.jsonl"
        self.personalities_file = "agents/agent_personalities.json"
        self.commands_file = "agents/agent_commands.jsonl"
        self.todos_file = "agents/agent_todos.jsonl"
        self._lock = threading.Lock()  # Protects concurrent file writes
        self.ensure_files_exist()
        self.agents = {}
        self.load_agent_data()
    
    def ensure_files_exist(self):
        """Create required files if they don't exist"""
        files = [
            self.agents_file,
            self.personalities_file,
            self.commands_file,
            self.todos_file
        ]
        
        for file_path in files:
            try:
                with open(file_path, 'r') as f:
                    pass
            except FileNotFoundError:
                with open(file_path, 'w') as f:
                    if file_path.endswith('.jsonl'):
                        f.write("")
                    else:
                        f.write("{}")
    
    def load_agent_data(self):
        """Load existing agent data"""
        # Load active agents
        try:
            with open(self.agents_file, 'r') as f:
                content = f.read().strip()
                if content:
                    self.agents = {line.split(' ', 1)[0]: json.loads(line.split(' ', 1)[1]) 
                                   for line in content.split('\n') if line.strip()}
        except FileNotFoundError:
            self.agents = {}
        
        # Load personalities if not exist
        try:
            with open(self.personalities_file, 'r') as f:
                self.personalities = json.load(f)
        except FileNotFoundError:
            self.create_default_personalities()
        
        # Check for stuck spawning agents on startup
        self.unstuck_spawning_agents()
    
    def create_default_personalities(self):
        """Create default agent personalities"""
        personalities = {
            "cadet": {
                "name": "Cadet Agent",
                "description": "Eager junior developer, quick with simple tasks",
                "strengths": ["speed", "enthusiasm", "learning"],
                "specialties": ["quick_fixes", "documentation", "testing"],
                "work_style": "rapid_iteration",
                "communication_style": "brief_and_focused"
            },
            "senior": {
                "name": "Senior Agent",
                "description": "Experienced developer, methodical and thorough",
                "strengths": ["experience", "architecture", "problem_solving"],
                "specialties": ["system_design", "refactoring", "optimization"],
                "work_style": "careful_planning",
                "communication_style": "detailed_and_comprehensive"
            },
            "specialist": {
                "name": "Specialist Agent",
                "description": "Domain expert for specific technologies",
                "strengths": ["expertise", "depth", "precision"],
                "specialties": ["databases", "security", "performance"],
                "work_style": "focused_deep_dive",
                "communication_style": "technical_detailed"
            },
            "researcher": {
                "name": "Researcher Agent",
                "description": "Investigative agent for exploring new solutions",
                "strengths": ["analysis", "exploration", "documentation"],
                "specialties": ["feature_research", "comparative_analysis", "prototyping"],
                "work_style": "systematic_investigation",
                "communication_style": "analytical_and_curious"
            },
            "debugger": {
                "name": "Debugger Agent",
                "description": "Specialized in troubleshooting and fixing issues",
                "strengths": ["troubleshooting", "precision", "persistence"],
                "specialties": ["bug_fixing", "testing", "root_cause_analysis"],
                "work_style": "systematic_debugging",
                "communication_style": "methodical_and_precise"
            },
            "optimizer": {
                "name": "Optimizer Agent",
                "description": "Focuses on improving performance and efficiency",
                "strengths": ["performance", "efficiency", "scalability"],
                "specialties": ["optimization", "refactoring", "benchmarking"],
                "work_style": "iterative_improvement",
                "communication_style": "metrics_focused"
            }
        }
        
        with open(self.personalities_file, 'w') as f:
            json.dump(personalities, f, indent=2)
        
        self.personalities = personalities
    
    def spawn_agent(self, personality: AgentPersonality, task_id: Optional[str] = None) -> str:
        """Spawn a new agent with specified personality"""
        agent_id = f"{personality.value}_{int(time.time())}"
        
        agent_data = {
            "id": agent_id,
            "personality": personality.value,
            "status": AgentStatus.SPAWNING.value,
            "created_at": datetime.now().isoformat(),
            "updated_at": datetime.now().isoformat(),
            "current_task": task_id,
            "completed_tasks": [],
            "active_session_id": None,
            "process_id": None,
            "thoughts": [],
            "metrics": {
                "tasks_completed": 0,
                "time_spent": 0,
                "success_rate": 0
            }
        }
        
        # Save to file
        self.save_agent(agent_id, agent_data)
        
        personality_data = self.personalities.get(personality.value, {"name": "Unknown Agent", "description": "No description", "strengths": [], "specialties": []})
        print(f"🚀 Spawning {personality.value} agent: {agent_id}")
        print(f"   Personality: {personality_data['name']}")
        print(f"   Task: {task_id or 'No task assigned'}")
        
        # Auto-transition to IDLE after a short delay to simulate agent startup
        def complete_spawning():
            time.sleep(2)  # Simulate startup time
            self.update_agent_status(agent_id, AgentStatus.IDLE, "Agent startup completed successfully")
        
        # Run in background thread
        threading.Thread(target=complete_spawning, daemon=True).start()
        
        return agent_id
    
    def save_agent(self, agent_id: str, agent_data: Dict):
        """Save agent data to file (thread-safe)."""
        with self._lock:
            self.agents[agent_id] = agent_data
            # Write atomically: build full content then replace file
            lines = [f"{aid} {json.dumps(data)}\n" for aid, data in self.agents.items()]
            with open(self.agents_file, 'w') as f:
                f.writelines(lines)
    
    def assign_task(self, agent_id: str, task_id: str) -> bool:
        """Assign a task to an agent"""
        if agent_id not in self.agents:
            print(f"❌ Agent {agent_id} not found")
            return False
        
        agent = self.agents[agent_id]
        
        if agent["status"] not in [AgentStatus.IDLE.value, AgentStatus.COMPLETED.value]:
            print(f"❌ Agent {agent_id} is not available (status: {agent['status']})")
            return False
        
        agent["current_task"] = task_id
        agent["status"] = AgentStatus.WORKING.value
        agent["updated_at"] = datetime.now().isoformat()
        
        # Add thought about task assignment
        self.add_agent_thought(agent_id, f"Assigned new task: {task_id}")
        
        # Create TODO entry for task assignment
        self.create_todo_entry(
            agent_id=agent_id,
            title=f"Task Assigned: {task_id}",
            description=f"Agent {agent_id} ({agent['personality']}) started working on task: {task_id}",
            status="in_progress",
            task_id=task_id
        )
        
        self.save_agent(agent_id, agent)
        return True
    
    def update_agent_status(self, agent_id: str, status: AgentStatus, thought: Optional[str] = None):
        """Update agent status and optionally add thought"""
        if agent_id not in self.agents:
            return
        
        agent = self.agents[agent_id]
        old_status = agent["status"]
        agent["status"] = status.value
        agent["updated_at"] = datetime.now().isoformat()
        
        # Log status change
        if old_status != status.value:
            self.add_agent_thought(agent_id, f"Status changed: {old_status} -> {status.value}")
        
        if thought:
            self.add_agent_thought(agent_id, thought)
        
        self.save_agent(agent_id, agent)
    
    def add_agent_thought(self, agent_id: str, thought: str):
        """Add a thought to agent's thought log"""
        if agent_id not in self.agents:
            return
        
        agent = self.agents[agent_id]
        thought_data = {
            "timestamp": datetime.now().isoformat(),
            "thought": thought
        }
        
        agent["thoughts"].append(thought_data)
        
        # Keep only last 50 thoughts to prevent memory bloat
        if len(agent["thoughts"]) > 50:
            agent["thoughts"] = agent["thoughts"][-50:]
        
        self.save_agent(agent_id, agent)
    
    def unstuck_spawning_agents(self):
        """Fix agents stuck in spawning state by transitioning them to idle"""
        stuck_agents = []
        for agent_id, agent_data in self.agents.items():
            if agent_data["status"] == AgentStatus.SPAWNING.value:
                # Check if agent has been spawning for more than 30 seconds
                created_at = datetime.fromisoformat(agent_data["created_at"])
                if (datetime.now() - created_at).total_seconds() > 30:
                    stuck_agents.append(agent_id)
        
        for agent_id in stuck_agents:
            self.update_agent_status(agent_id, AgentStatus.IDLE, "Auto-recovered from stuck spawning state")
            print(f"🔧 Recovered stuck agent {agent_id} from spawning state")
        
        return stuck_agents
    
    def create_todo_entry(self, agent_id: str, title: str, description: str, status: str = "pending", task_id: Optional[str] = None) -> str:
        """Create a TODO entry for agent activity"""
        todo_id = str(uuid.uuid4())
        todo_data = {
            "id": todo_id,
            "agent_id": agent_id,
            "title": title,
            "description": description,
            "status": status,
            "priority": "medium",
            "created_at": datetime.now().isoformat(),
            "updated_at": datetime.now().isoformat(),
            "task_id": task_id,
            "tags": ["agent", "auto-generated"],
            "metadata": {
                "agent_personality": self.agents.get(agent_id, {}).get("personality", "unknown"),
                "event_type": "agent_activity"
            }
        }
        
        # Save to todos file
        with open(self.todos_file, 'a') as f:
            f.write(f"{todo_id} {json.dumps(todo_data)}\n")
        
        # Also integrate with OpenCode TODO system if available
        try:
            import subprocess
            import os
            
            # Check if we're in the OpenCode environment
            if os.path.exists("/home/meridian/strat/Stratavore"):
                # Create a simplified todo for OpenCode integration
                opencode_todo = {
                    "content": f"🤖 {agent_id}: {title}",
                    "status": "pending" if status == "pending" else "in_progress" if status == "in_progress" else "completed",
                    "priority": "medium",
                    "id": f"agent-{todo_id}"
                }
                
                # Write to a shared location that OpenCode can access
                with open("/tmp/opencode_agent_todos.jsonl", 'a') as f:
                    f.write(json.dumps(opencode_todo) + "\n")
                    
        except Exception as e:
            print(f"Note: OpenCode integration not available: {e}")
        
        return todo_id
    
    def update_todo_status(self, todo_id: str, status: str, notes: Optional[str] = None):
        """Update TODO entry status"""
        # Load existing todos
        todos = {}
        try:
            with open(self.todos_file, 'r') as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    parts = line.split(' ', 1)
                    if len(parts) == 2:
                        try:
                            todos[parts[0]] = json.loads(parts[1])
                        except json.JSONDecodeError:
                            pass
        except FileNotFoundError:
            return
        
        if todo_id not in todos:
            return
        
        # Update todo
        todo = todos[todo_id]
        todo["status"] = status
        todo["updated_at"] = datetime.now().isoformat()
        if notes:
            if "metadata" not in todo:
                todo["metadata"] = {}
            todo["metadata"]["completion_notes"] = notes
        
        # Save back to file
        with open(self.todos_file, 'w') as f:
            for tid, tdata in todos.items():
                f.write(f"{tid} {json.dumps(tdata)}\n")
    
    def get_agent_todos(self, agent_id: Optional[str] = None) -> List[Dict]:
        """Get TODO entries, filtered by agent if specified"""
        todos = []
        try:
            with open(self.todos_file, 'r') as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    parts = line.split(' ', 1)
                    if len(parts) != 2:
                        continue
                    try:
                        todo = json.loads(parts[1])
                        if agent_id is None or todo.get("agent_id") == agent_id:
                            todos.append(todo)
                    except json.JSONDecodeError:
                        pass
        except FileNotFoundError:
            pass
        
        # Sort by created_at descending
        todos.sort(key=lambda x: x.get("created_at", ""), reverse=True)
        return todos
    
    def complete_task(self, agent_id: str, success: bool = True, notes: str = ""):
        """Mark agent's current task as completed"""
        if agent_id not in self.agents:
            return False
        
        agent = self.agents[agent_id]
        task_id = agent.get("current_task")
        
        if task_id:
            agent["completed_tasks"].append({
                "task_id": task_id,
                "completed_at": datetime.now().isoformat(),
                "success": success,
                "notes": notes
            })
            
            # Update metrics
            agent["metrics"]["tasks_completed"] += 1
            if success:
                agent["metrics"]["success_rate"] = (
                    (agent["metrics"]["success_rate"] * (agent["metrics"]["tasks_completed"] - 1) + 100) /
                    agent["metrics"]["tasks_completed"]
                )
            
            self.add_agent_thought(agent_id, f"Completed task {task_id}: {'SUCCESS' if success else 'FAILED'}")
            
            # Create TODO entry for task completion
            todo_title = f"Task Completed: {task_id}" if success else f"Task Failed: {task_id}"
            todo_status = "completed" if success else "cancelled"
            self.create_todo_entry(
                agent_id=agent_id,
                title=todo_title,
                description=f"Agent {agent_id} ({agent['personality']}) {'completed' if success else 'failed to complete'} task: {task_id}. Notes: {notes}",
                status=todo_status,
                task_id=task_id
            )
        
        agent["current_task"] = None
        agent["status"] = AgentStatus.IDLE.value
        agent["updated_at"] = datetime.now().isoformat()
        
        self.save_agent(agent_id, agent)
        
        result = "SUCCESS" if success else "FAILED"
        print(f"✅ Marked task {task_id} as {result} for agent {agent_id}")
        return True
    
    def get_available_agents(self, personality: Optional[AgentPersonality] = None) -> List[str]:
        """Get list of available agents (idle or completed)"""
        available = []
        
        for agent_id, agent_data in self.agents.items():
            if agent_data["status"] in [AgentStatus.IDLE.value, AgentStatus.COMPLETED.value]:
                if personality is None or agent_data["personality"] == personality.value:
                    available.append(agent_id)
        
        return available
    
    def get_agent_summary(self) -> Dict:
        """Get summary of all agents"""
        summary = {
            "total_agents": len(self.agents),
            "by_status": {},
            "by_personality": {},
            "working_agents": 0,
            "idle_agents": 0,
            "total_tasks_completed": 0
        }
        
        for agent_data in self.agents.values():
            # Count by status
            status = agent_data["status"]
            summary["by_status"][status] = summary["by_status"].get(status, 0) + 1
            
            # Count by personality
            personality = agent_data["personality"]
            summary["by_personality"][personality] = summary["by_personality"].get(personality, 0) + 1
            
            # Count working/idle
            if status == AgentStatus.WORKING.value:
                summary["working_agents"] += 1
            elif status in [AgentStatus.IDLE.value, AgentStatus.COMPLETED.value]:
                summary["idle_agents"] += 1
            
            # Total tasks completed
            summary["total_tasks_completed"] += agent_data["metrics"]["tasks_completed"]
        
        return summary

def main():
    """CLI interface for agent management"""
    if len(sys.argv) < 2:
        print("Usage: python3 agent_manager.py [command]")
        print("Commands:")
        print("  spawn <personality> [task_id]              - Spawn new agent")
        print("  assign <agent_id> <task_id>               - Assign task to agent")
        print("  status <agent_id> <status> [thought]    - Update agent status")
        print("  complete <agent_id> [success] [notes]     - Complete current task")
        print("  list [personality]                          - List agents")
        print("  available [personality]                      - List available agents")
        print("  summary                                      - Show agent summary")
        print("  personalities                                - Show available personalities")
        return
    
    manager = AgentManager()
    command = sys.argv[1]
    
    if command == "spawn":
        if len(sys.argv) < 3:
            print("Usage: python3 agent_manager.py spawn <personality> [task_id]")
            return
        
        personality_str = sys.argv[2]
        task_id = sys.argv[3] if len(sys.argv) > 3 else None
        
        try:
            personality_enum = AgentPersonality(personality_str.lower())
            agent_id = manager.spawn_agent(personality_enum, task_id)
            print(f"✅ Spawned agent: {agent_id}")
        except ValueError:
            print(f"❌ Invalid personality: {personality_str}")
            print(f"Available personalities: {[p.value for p in AgentPersonality]}")
    
    elif command == "list":
        personality_filter = sys.argv[2] if len(sys.argv) > 2 else None
        
        if personality_filter:
            try:
                personality_enum = AgentPersonality(personality_filter.lower())
                agents = [aid for aid, data in manager.agents.items() 
                          if data["personality"] == personality_enum.value]
                print(f"🤖 {personality_filter} agents ({len(agents)}):")
            except ValueError:
                print(f"❌ Invalid personality: {personality_filter}")
                return
        else:
            agents = list(manager.agents.keys())
            print(f"🤖 All agents ({len(agents)}):")
        
        for agent_id in agents:
            agent = manager.agents[agent_id]
            status_emoji = {
                "idle": "😴",
                "working": "🔨", 
                "spawning": "🚀",
                "completed": "✅",
                "error": "❌"
            }.get(agent["status"], "❓")
            
            print(f"  {status_emoji} {agent_id} ({agent['personality']}) - {agent['status']}")
            if agent.get("current_task"):
                print(f"      📋 Task: {agent['current_task']}")
    
    elif command == "available":
        personality_filter_str = sys.argv[2] if len(sys.argv) > 2 else None
        
        try:
            personality_filter = AgentPersonality(personality_filter_str.lower()) if personality_filter_str else None
            available = manager.get_available_agents(personality_filter)
            
            if available:
                print(f"✅ Available agents ({len(available)}):")
                for agent_id in available:
                    agent = manager.agents[agent_id]
                    print(f"  🤖 {agent_id} ({agent['personality']})")
            else:
                print("📭 No available agents")
        except ValueError:
            print(f"❌ Invalid personality: {personality_filter_str}")
    
    elif command == "summary":
        summary = manager.get_agent_summary()
        print(f"📊 AGENT SUMMARY")
        print(f"Total agents: {summary['total_agents']}")
        print(f"Working: {summary['working_agents']}")
        print(f"Idle: {summary['idle_agents']}")
        print(f"Total tasks completed: {summary['total_tasks_completed']}")
        
        print(f"\nBy status:")
        for status, count in summary["by_status"].items():
            emoji = {"idle": "😴", "working": "🔨", "completed": "✅"}.get(status, "❓")
            print(f"  {emoji} {status}: {count}")
        
        print(f"\nBy personality:")
        for personality, count in summary["by_personality"].items():
            print(f"  🤖 {personality}: {count}")
    
    elif command == "personalities":
        print("🎭 Available Agent Personalities:")
        for personality in AgentPersonality:
            p = manager.personalities.get(personality.value, {})
            print(f"  🎭 {personality.value}")
            print(f"      Name: {p.get('name', 'Unknown')}")
            print(f"      Description: {p.get('description', 'No description')}")
            print(f"      Strengths: {', '.join(p.get('strengths', []))}")
            print(f"      Specialties: {', '.join(p.get('specialties', []))}")
            print()
    
    elif command == "assign":
        if len(sys.argv) < 4:
            print("Usage: python3 agent_manager.py assign <agent_id> <task_id>")
            return
        
        agent_id = sys.argv[2]
        task_id = sys.argv[3]
        
        if manager.assign_task(agent_id, task_id):
            print(f"✅ Assigned task {task_id} to agent {agent_id}")
        else:
            print(f"❌ Failed to assign task")
    
    elif command == "status":
        if len(sys.argv) < 4:
            print("Usage: python3 agent_manager.py status <agent_id> <status> [thought]")
            return
        
        agent_id = sys.argv[2]
        status_str = sys.argv[3]
        thought = sys.argv[4] if len(sys.argv) > 4 else None
        
        try:
            status_enum = AgentStatus(status_str.lower())
            manager.update_agent_status(agent_id, status_enum, thought)
            print(f"✅ Updated {agent_id} status to {status_str}")
            if thought:
                print(f"💭 Thought: {thought}")
        except ValueError:
            print(f"❌ Invalid status: {status_str}")
            print(f"Available statuses: {[s.value for s in AgentStatus]}")
    
    elif command == "complete":
        if len(sys.argv) < 3:
            print("Usage: python3 agent_manager.py complete <agent_id> [success] [notes]")
            return
        
        agent_id = sys.argv[2]
        success = True if len(sys.argv) < 4 else sys.argv[3].lower() == 'true'
        notes = sys.argv[4] if len(sys.argv) > 4 else ""
        
        if manager.complete_task(agent_id, success, notes):
            result = "SUCCESS" if success else "FAILED"
            print(f"✅ Marked task as {result} for agent {agent_id}")
        else:
            print(f"❌ Failed to complete task for agent {agent_id}")
    
    elif command == "unstuck-spawning":
        stuck_agents = manager.unstuck_spawning_agents()
        if stuck_agents:
            print(f"🔧 Unstuck {len(stuck_agents)} agents from spawning state")
        else:
            print("✅ No agents stuck in spawning state")
    
    else:
        print(f"❌ Unknown command: {command}")

if __name__ == "__main__":
    main()