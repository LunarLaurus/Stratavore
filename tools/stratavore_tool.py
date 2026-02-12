#!/usr/bin/env python3
"""
Stratavore Agent Tool Interface for Meridian Lex
Provides programmatic access to the Stratavore agent/task system
"""

import requests
import json
import time
import uuid
from typing import Dict, List, Optional, Any
from enum import Enum
from dataclasses import dataclass
import os


class AgentPersonality(Enum):
    """Available agent personalities in the Stratavore system"""
    CADET = "cadet"
    SENIOR = "senior"
    SPECIALIST = "specialist"
    RESEARCHER = "researcher"
    DEBUGGER = "debugger"
    OPTIMIZER = "optimizer"


class AgentStatus(Enum):
    """Agent status states"""
    SPAWNING = "spawning"
    IDLE = "idle"
    WORKING = "working"
    PAUSED = "paused"
    COMPLETED = "completed"
    ERROR = "error"


@dataclass
class Job:
    """Job data structure"""
    id: str
    title: str
    description: str
    status: str = "pending"
    priority: str = "medium"
    created_at: str = ""
    updated_at: str = ""
    assignee: Optional[str] = None
    labels: Optional[List[str]] = None
    estimated_hours: float = 0
    actual_hours: float = 0
    dependencies: Optional[List[str]] = None
    deliverables: Optional[List[str]] = None

    def __post_init__(self):
        if self.labels is None:
            self.labels = []
        if self.dependencies is None:
            self.dependencies = []
        if self.deliverables is None:
            self.deliverables = []
        if not self.created_at:
            self.created_at = time.strftime("%Y-%m-%dT%H:%M:%SZ")
        if not self.updated_at:
            self.updated_at = self.created_at


@dataclass
class Agent:
    """Agent data structure"""
    id: str
    personality: str
    status: str
    created_at: str
    updated_at: str
    current_task: Optional[str] = None
    completed_tasks: Optional[List[str]] = None
    active_session_id: Optional[str] = None
    process_id: Optional[str] = None
    thoughts: Optional[List[Dict]] = None
    metrics: Optional[Dict] = None

    def __post_init__(self):
        if self.completed_tasks is None:
            self.completed_tasks = []
        if self.thoughts is None:
            self.thoughts = []
        if self.metrics is None:
            self.metrics = {"tasks_completed": 0, "time_spent": 0, "success_rate": 0}


class StratavoreTool:
    """
    Meridian Lex's interface to the Stratavore Agent/Task System
    Provides high-level operations for agent and task management
    """

    def __init__(self, base_url: str = "http://localhost:8080", timeout: int = 30):
        """
        Initialize the Stratavore tool interface
        
        Args:
            base_url: Base URL of the Stratavore web UI server
            timeout: Request timeout in seconds
        """
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        
    def _make_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
        """
        Make HTTP request to the Stratavore API
        
        Args:
            method: HTTP method (GET, POST)
            endpoint: API endpoint path
            data: Request payload for POST requests
            
        Returns:
            Response data as dictionary
            
        Raises:
            requests.RequestException: On network/HTTP errors
        """
        url = f"{self.base_url}{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        try:
            if method.upper() == "GET":
                response = self.session.get(url, headers=headers, timeout=self.timeout)
            elif method.upper() == "POST":
                response = self.session.post(url, headers=headers, json=data, timeout=self.timeout)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
            response.raise_for_status()
            return response.json()
            
        except requests.exceptions.RequestException as e:
            raise requests.RequestException(f"API request failed: {e}")

    def health_check(self) -> Dict:
        """
        Check the health status of the Stratavore server
        
        Returns:
            Health status information
        """
        return self._make_request("GET", "/api/health")

    def get_system_status(self) -> Dict:
        """
        Get complete system status including jobs, agents, and progress
        
        Returns:
            Complete system status data
        """
        return self._make_request("GET", "/api/status")

    def get_agents(self) -> Dict:
        """
        Get all agents and their todos
        
        Returns:
            Agent data with summary statistics
        """
        return self._make_request("GET", "/api/agents")

    def spawn_agent(self, personality: AgentPersonality, task_id: Optional[str] = None) -> Dict:
        """
        Spawn a new agent with specified personality
        
        Args:
            personality: Agent personality type
            task_id: Optional task to assign immediately
            
        Returns:
            Spawn result with agent_id
        """
        data = {"personality": personality.value}
        if task_id:
            data["task_id"] = task_id
            
        return self._make_request("POST", "/api/spawn-agent", data)

    def assign_task(self, agent_id: str, task_id: str) -> Dict:
        """
        Assign a task to an agent
        
        Args:
            agent_id: ID of the agent
            task_id: ID of the task to assign
            
        Returns:
            Assignment result
        """
        data = {"agent_id": agent_id, "task_id": task_id}
        return self._make_request("POST", "/api/assign-agent", data)

    def complete_task(self, agent_id: str, success: bool = True, notes: str = "") -> Dict:
        """
        Mark an agent's current task as completed
        
        Args:
            agent_id: ID of the agent
            success: Whether the task completed successfully
            notes: Optional completion notes
            
        Returns:
            Completion result
        """
        data = {"agent_id": agent_id, "success": success, "notes": notes}
        return self._make_request("POST", "/api/complete-task", data)

    def update_agent_status(self, agent_id: str, status: AgentStatus, thought: str = "") -> Dict:
        """
        Update an agent's status with optional thought log
        
        Args:
            agent_id: ID of the agent
            status: New status
            thought: Optional thought/memo
            
        Returns:
            Status update result
        """
        data = {"agent_id": agent_id, "status": status.value, "thought": thought}
        return self._make_request("POST", "/api/agent-status", data)

    def kill_agent(self, agent_id: str) -> Dict:
        """
        Terminate an agent (marks as ERROR status)
        
        Args:
            agent_id: ID of the agent to terminate
            
        Returns:
            Termination result
        """
        data = {"agent_id": agent_id}
        return self._make_request("POST", "/api/kill-agent")

    def wait_for_agent_ready(self, agent_id: str, timeout_seconds: int = 30) -> bool:
        """
        Wait for an agent to transition from SPAWNING to IDLE
        
        Args:
            agent_id: ID of the agent to monitor
            timeout_seconds: Maximum time to wait
            
        Returns:
            True if agent became ready, False if timeout
        """
        start_time = time.time()
        while time.time() - start_time < timeout_seconds:
            try:
                agents_data = self.get_agents()
                agents = agents_data.get("agents", {})
                agent = agents.get(agent_id)
                
                if agent and agent.get("status") == "idle":
                    return True
                elif agent and agent.get("status") == "error":
                    return False
                    
                time.sleep(1)
            except requests.RequestException:
                time.sleep(1)
                continue
                
        return False

    def find_available_agent(self, personality: Optional[AgentPersonality] = None) -> Optional[str]:
        """
        Find an available IDLE agent, optionally filtered by personality
        
        Args:
            personality: Optional personality filter
            
        Returns:
            Agent ID if available, None otherwise
        """
        try:
            agents_data = self.get_agents()
            agents = agents_data.get("agents", {})
            
            for agent_id, agent_data in agents.items():
                if agent_data.get("status") == "idle":
                    if personality is None or agent_data.get("personality") == personality.value:
                        return agent_id
                        
            return None
        except requests.RequestException:
            return None

    def create_job_definition(self, title: str, description: str, **kwargs) -> Dict:
        """
        Create a job definition dictionary
        
        Args:
            title: Job title
            description: Job description
            **kwargs: Additional job attributes
            
        Returns:
            Job definition as dictionary
        """
        job_id = kwargs.get('id', f"job-{time.strftime('%Y-%m-%d-%H%M%S')}")
        
        job = {
            "id": job_id,
            "title": title,
            "description": description,
            "status": kwargs.get("status", "pending"),
            "priority": kwargs.get("priority", "medium"),
            "created_at": kwargs.get("created_at", time.strftime("%Y-%m-%dT%H:%M:%SZ")),
            "updated_at": kwargs.get("updated_at", time.strftime("%Y-%m-%dT%H:%M:%SZ")),
            "assignee": kwargs.get("assignee"),
            "labels": kwargs.get("labels", []),
            "estimated_hours": kwargs.get("estimated_hours", 0),
            "actual_hours": kwargs.get("actual_hours", 0),
            "dependencies": kwargs.get("dependencies", []),
            "deliverables": kwargs.get("deliverables", [])
        }
        
        return job

    def get_agent_summary(self) -> Dict:
        """
        Get a concise summary of all agents and their states
        
        Returns:
            Summary dictionary with agent counts and details
        """
        try:
            agents_data = self.get_agents()
            return agents_data.get("summary", {})
        except requests.RequestException as e:
            return {"error": str(e)}

    def execute_workflow(self, workflow_steps: List[Dict]) -> Dict:
        """
        Execute a multi-step workflow with agent spawning and task assignment
        
        Args:
            workflow_steps: List of workflow step dictionaries
                Each step should have:
                - personality: AgentPersonality or string
                - task_description: Task description
                - task_title: Task title
                - dependencies: List of step indices this step depends on
                - optional task_id and agent_id
            
        Returns:
            Workflow execution results
        """
        results = {
            "workflow_id": str(uuid.uuid4()),
            "steps_completed": [],
            "steps_failed": [],
            "agents_spawned": [],
            "total_execution_time": 0
        }
        
        start_time = time.time()
        
        try:
            for i, step in enumerate(workflow_steps):
                # Check dependencies
                if "dependencies" in step:
                    deps_met = all(dep_idx in results["steps_completed"] 
                                  for dep_idx in step["dependencies"])
                    if not deps_met:
                        results["steps_failed"].append({
                            "step_index": i,
                            "error": "Dependencies not met",
                            "dependencies": step["dependencies"]
                        })
                        continue
                
                # Spawn agent
                personality = step["personality"]
                if isinstance(personality, str):
                    personality = AgentPersonality(personality.lower())
                
                spawn_result = self.spawn_agent(personality)
                if spawn_result.get("status") != "success":
                    results["steps_failed"].append({
                        "step_index": i,
                        "error": "Failed to spawn agent",
                        "spawn_result": spawn_result
                    })
                    continue
                
                agent_id = spawn_result["agent_id"]
                results["agents_spawned"].append(agent_id)
                
                # Wait for agent to be ready
                if not self.wait_for_agent_ready(agent_id):
                    results["steps_failed"].append({
                        "step_index": i,
                        "error": "Agent failed to become ready",
                        "agent_id": agent_id
                    })
                    continue
                
                # Create task if not provided
                task_id = step.get("task_id")
                if not task_id:
                    job_def = self.create_job_definition(
                        title=step["task_title"],
                        description=step["task_description"]
                    )
                    task_id = job_def["id"]
                
                # Assign task
                assign_result = self.assign_task(agent_id, task_id)
                if assign_result.get("status") != "success":
                    results["steps_failed"].append({
                        "step_index": i,
                        "error": "Failed to assign task",
                        "agent_id": agent_id,
                        "task_id": task_id
                    })
                    continue
                
                results["steps_completed"].append({
                    "step_index": i,
                    "agent_id": agent_id,
                    "task_id": task_id,
                    "personality": personality.value
                })
        
        except Exception as e:
            results["workflow_error"] = str(e)
        
        finally:
            results["total_execution_time"] = time.time() - start_time
        
        return results


# Convenience functions for common operations
def quick_spawn_agent(personality: str, base_url: str = "http://localhost:8080") -> Optional[str]:
    """
    Quick function to spawn an agent and return its ID
    
    Args:
        personality: Agent personality type
        base_url: Stratavore server URL
        
    Returns:
        Agent ID if successful, None otherwise
    """
    try:
        tool = StratavoreTool(base_url)
        result = tool.spawn_agent(AgentPersonality(personality.lower()))
        return result.get("agent_id") if result.get("status") == "success" else None
    except Exception:
        return None


def get_system_health(base_url: str = "http://localhost:8080") -> bool:
    """
    Quick health check of the Stratavore system
    
    Args:
        base_url: Stratavore server URL
        
    Returns:
        True if system is healthy, False otherwise
    """
    try:
        tool = StratavoreTool(base_url)
        health = tool.health_check()
        return health.get("status") == "healthy"
    except Exception:
        return False


if __name__ == "__main__":
    # Example usage and testing
    print("Stratavore Tool Interface - Meridian Lex")
    print("=" * 50)
    
    # Initialize tool
    stratavore = StratavoreTool()
    
    # Health check
    print("Checking system health...")
    health = stratavore.health_check()
    print(f"System status: {health.get('status', 'unknown')}")
    
    # Get agent summary
    print("\nGetting agent summary...")
    summary = stratavore.get_agent_summary()
    print(f"Total agents: {summary.get('total', 0)}")
    
    # Spawn a test agent
    print("\nSpawning CADET agent...")
    spawn_result = stratavore.spawn_agent(AgentPersonality.CADET)
    print(f"Spawn result: {spawn_result}")
    
    if spawn_result.get("status") == "success":
        agent_id = spawn_result["agent_id"]
        print(f"Agent spawned: {agent_id}")
        
        # Wait for readiness
        print("Waiting for agent to be ready...")
        if stratavore.wait_for_agent_ready(agent_id):
            print("Agent is ready!")
            
            # Assign a task
            task_id = f"test-task-{int(time.time())}"
            print(f"Assigning task: {task_id}")
            assign_result = stratavore.assign_task(agent_id, task_id)
            print(f"Assignment result: {assign_result}")
        else:
            print("Agent failed to become ready")