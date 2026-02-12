#!/usr/bin/env python3
"""
Meridian Lex Stratavore Integration
High-level command interface for fleet operations
"""

import sys
import os
import json
import time
from typing import List, Dict, Optional, Any
from pathlib import Path

# Add the tools directory to Python path for imports
CURRENT_DIR = Path(__file__).parent
TOOLS_DIR = CURRENT_DIR.parent / "tools"
sys.path.insert(0, str(TOOLS_DIR))

from stratavore_tool import StratavoreTool, AgentPersonality, AgentStatus


class MeridianLexStratavore:
    """
    Meridian Lex's Stratavore Command Interface
    Fleet-grade agent and task management operations
    """
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        """Initialize the command interface"""
        self.tool = StratavoreTool(base_url)
        self.mission_log = []
        
    def log_operation(self, operation: str, details: Dict):
        """Log operation for audit trail"""
        log_entry = {
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "operation": operation,
            "details": details
        }
        self.mission_log.append(log_entry)
        print(f"[{log_entry['timestamp']}] {operation}: {details}")
        
    def fleet_status(self) -> Dict:
        """Get comprehensive fleet status"""
        try:
            status = self.tool.get_system_status()
            agents = self.tool.get_agents()
            
            fleet_report = {
                "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
                "health": status.get("status", "unknown"),
                "total_agents": len(agents.get("agents", {})),
                "active_jobs": len([j for j in status.get("jobs", []) if j.get("status") == "in_progress"]),
                "agent_summary": agents.get("summary", {}),
                "mission_log_size": len(self.mission_log)
            }
            
            self.log_operation("FLEET_STATUS_CHECK", fleet_report)
            return fleet_report
            
        except Exception as e:
            error_msg = f"Failed to get fleet status: {e}"
            self.log_operation("ERROR", {"message": error_msg})
            return {"error": error_msg}
    
    def deploy_agent(self, personality: str, task_description: str = "", immediate_task: bool = False) -> Optional[str]:
        """
        Deploy a new agent to the fleet
        
        Args:
            personality: Agent personality type
            task_description: Optional task to assign immediately
            immediate_task: Whether to create and assign a task immediately
            
        Returns:
            Agent ID if successful, None otherwise
        """
        try:
            personality_enum = AgentPersonality(personality.lower())
            
            # Spawn the agent
            spawn_result = self.tool.spawn_agent(personality_enum)
            
            if spawn_result.get("status") != "success":
                self.log_operation("AGENT_DEPLOY_FAILED", {
                    "personality": personality,
                    "error": spawn_result.get("error", "Unknown error")
                })
                return None
            
            agent_id = spawn_result["agent_id"]
            self.log_operation("AGENT_DEPLOYED", {
                "agent_id": agent_id,
                "personality": personality
            })
            
            # Wait for agent to be ready
            if self.tool.wait_for_agent_ready(agent_id):
                self.log_operation("AGENT_READY", {"agent_id": agent_id})
                
                # Create and assign task if requested
                if immediate_task and task_description:
                    task_id = f"mission-{int(time.time())}"
                    job_def = self.tool.create_job_definition(
                        title=f"Auto-assigned task for {personality}",
                        description=task_description
                    )
                    
                    assign_result = self.tool.assign_task(agent_id, task_id)
                    if assign_result.get("status") == "success":
                        self.log_operation("TASK_ASSIGNED", {
                            "agent_id": agent_id,
                            "task_id": task_id,
                            "description": task_description
                        })
                    else:
                        self.log_operation("TASK_ASSIGN_FAILED", {
                            "agent_id": agent_id,
                            "error": assign_result.get("error", "Unknown error")
                        })
                
                return agent_id
            else:
                self.log_operation("AGENT_READY_TIMEOUT", {"agent_id": agent_id})
                return None
                
        except Exception as e:
            self.log_operation("AGENT_DEPLOY_ERROR", {
                "personality": personality,
                "error": str(e)
            })
            return None
    
    def form_squad(self, composition: Dict[str, int], mission_brief: str) -> Dict:
        """
        Form a specialized squad of agents for a mission
        
        Args:
            composition: Dict of personality -> count (e.g., {"cadet": 2, "senior": 1})
            mission_brief: Mission description for task assignment
            
        Returns:
            Squad formation results
        """
        squad_results = {
            "mission_id": f"mission-{int(time.time())}",
            "composition_requested": composition,
            "agents_deployed": [],
            "agents_failed": [],
            "total_deployed": 0
        }
        
        self.log_operation("SQUAD_FORMATION_INITIATED", {
            "mission_id": squad_results["mission_id"],
            "composition": composition
        })
        
        for personality, count in composition.items():
            for i in range(count):
                agent_id = self.deploy_agent(
                    personality=personality,
                    task_description=f"{mission_brief} (Agent {i+1} of {count})",
                    immediate_task=True
                )
                
                if agent_id:
                    squad_results["agents_deployed"].append({
                        "agent_id": agent_id,
                        "personality": personality,
                        "sequence": i+1
                    })
                    squad_results["total_deployed"] += 1
                else:
                    squad_results["agents_failed"].append({
                        "personality": personality,
                        "sequence": i+1
                    })
        
        self.log_operation("SQUAD_FORMATION_COMPLETE", squad_results)
        return squad_results
    
    def recall_agent(self, agent_id: str, reason: str = "Mission complete") -> bool:
        """
        Recall an agent (mark as completed)
        
        Args:
            agent_id: ID of agent to recall
            reason: Reason for recall
            
        Returns:
            True if successful, False otherwise
        """
        try:
            result = self.tool.complete_task(agent_id, True, reason)
            if result.get("status") == "success":
                self.log_operation("AGENT_RECALLED", {
                    "agent_id": agent_id,
                    "reason": reason
                })
                return True
            else:
                self.log_operation("AGENT_RECALL_FAILED", {
                    "agent_id": agent_id,
                    "error": result.get("error", "Unknown error")
                })
                return False
                
        except Exception as e:
            self.log_operation("AGENT_RECALL_ERROR", {
                "agent_id": agent_id,
                "error": str(e)
            })
            return False
    
    def emergency_stop(self, agent_id: str) -> bool:
        """
        Emergency stop an agent (kill)
        
        Args:
            agent_id: ID of agent to stop
            
        Returns:
            True if successful, False otherwise
        """
        try:
            result = self.tool.kill_agent(agent_id)
            if result.get("status") == "success":
                self.log_operation("EMERGENCY_STOP", {"agent_id": agent_id})
                return True
            else:
                self.log_operation("EMERGENCY_STOP_FAILED", {
                    "agent_id": agent_id,
                    "error": result.get("error", "Unknown error")
                })
                return False
                
        except Exception as e:
            self.log_operation("EMERGENCY_STOP_ERROR", {
                "agent_id": agent_id,
                "error": str(e)
            })
            return False
    
    def get_available_for_duty(self, personality: Optional[str] = None) -> List[str]:
        """
        Get list of agents available for duty (IDLE status)
        
        Args:
            personality: Optional personality filter
            
        Returns:
            List of available agent IDs
        """
        try:
            personality_enum = None
            if personality:
                personality_enum = AgentPersonality(personality.lower())
            
            agent_id = self.tool.find_available_agent(personality_enum)
            
            if agent_id:
                # Find all available agents by scanning
                available_agents = []
                agents_data = self.tool.get_agents()
                agents = agents_data.get("agents", {})
                
                for aid, agent_data in agents.items():
                    if agent_data.get("status") == "idle":
                        if personality is None or agent_data.get("personality") == personality.lower():
                            available_agents.append(aid)
                
                return available_agents
            else:
                return []
                
        except Exception as e:
            self.log_operation("GET_AVAILABLE_ERROR", {"error": str(e)})
            return []
    
    def execute_mission(self, mission_plan: List[Dict]) -> Dict:
        """
        Execute a complex mission with multiple steps
        
        Args:
            mission_plan: List of mission steps
                Each step: {
                    "name": "Step name",
                    "personality": "agent_type",
                    "task": "Task description",
                    "critical": bool,  # Whether mission fails if this step fails
                    "dependencies": []  # List of step indices this depends on
                }
            
        Returns:
            Mission execution report
        """
        mission_report = {
            "mission_id": f"mission-{int(time.time())}",
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "steps_total": len(mission_plan),
            "steps_completed": [],
            "steps_failed": [],
            "critical_failures": [],
            "agents_deployed": [],
            "mission_success": False,
            "execution_time": 0
        }
        
        start_time = time.time()
        
        self.log_operation("MISSION_EXECUTION_STARTED", {
            "mission_id": mission_report["mission_id"],
            "steps": len(mission_plan)
        })
        
        try:
            for i, step in enumerate(mission_plan):
                step_name = step.get("name", f"Step {i+1}")
                
                # Check dependencies
                if "dependencies" in step and step["dependencies"]:
                    deps_met = all(
                        any(step_name == completed_step.get("step_name") 
                            for completed_step in mission_report["steps_completed"])
                        for step_name in step["dependencies"]
                    )
                    
                    if not deps_met:
                        step_failure = {
                            "step_index": i,
                            "step_name": step_name,
                            "error": f"Dependencies not met: {step['dependencies']}"
                        }
                        mission_report["steps_failed"].append(step_failure)
                        
                        if step.get("critical", False):
                            mission_report["critical_failures"].append(step_failure)
                        
                        continue
                
                # Deploy agent for this step
                agent_id = self.deploy_agent(
                    personality=step["personality"],
                    task_description=step["task"],
                    immediate_task=True
                )
                
                if agent_id:
                    mission_report["agents_deployed"].append(agent_id)
                    
                    step_completion = {
                        "step_index": i,
                        "step_name": step_name,
                        "agent_id": agent_id,
                        "personality": step["personality"],
                        "task": step["task"]
                    }
                    mission_report["steps_completed"].append(step_completion)
                    
                else:
                    step_failure = {
                        "step_index": i,
                        "step_name": step_name,
                        "error": "Failed to deploy agent"
                    }
                    mission_report["steps_failed"].append(step_failure)
                    
                    if step.get("critical", False):
                        mission_report["critical_failures"].append(step_failure)
        
        except Exception as e:
            self.log_operation("MISSION_EXECUTION_ERROR", {"error": str(e)})
            mission_report["mission_error"] = str(e)
        
        finally:
            mission_report["execution_time"] = time.time() - start_time
            mission_report["mission_success"] = (
                len(mission_report["critical_failures"]) == 0 and
                len(mission_report["steps_completed"]) > 0
            )
            
            self.log_operation("MISSION_EXECUTION_COMPLETE", {
                "mission_id": mission_report["mission_id"],
                "success": mission_report["mission_success"],
                "steps_completed": len(mission_report["steps_completed"]),
                "steps_failed": len(mission_report["steps_failed"])
            })
        
        return mission_report
    
    def generate_mission_report(self) -> Dict:
        """Generate comprehensive mission and fleet report"""
        fleet_status = self.fleet_status()
        
        report = {
            "report_timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "fleet_status": fleet_status,
            "recent_operations": self.mission_log[-10:],  # Last 10 operations
            "total_operations": len(self.mission_log),
            "recommendations": []
        }
        
        # Add recommendations based on fleet status
        if fleet_status.get("total_agents", 0) < 3:
            report["recommendations"].append("Consider deploying more agents for optimal fleet capacity")
        
        active_jobs = fleet_status.get("active_jobs", 0)
        if active_jobs > fleet_status.get("total_agents", 0):
            report["recommendations"].append("High workload detected - consider deploying additional agents")
        
        return report


def main():
    """Command line interface for Meridian Lex Stratavore operations"""
    if len(sys.argv) < 2:
        print("Meridian Lex Stratavore Command Interface")
        print("Usage: python meridian_lex_stratavore.py <command> [args]")
        print("\nCommands:")
        print("  status                    - Show fleet status")
        print("  deploy <personality>      - Deploy single agent")
        print("  squad <json> <mission>    - Form squad (JSON: {'cadet': 2, 'senior': 1})")
        print("  available [personality]   - List available agents")
        print("  recall <agent_id>        - Recall agent")
        print("  stop <agent_id>           - Emergency stop agent")
        print("  report                    - Generate mission report")
        return
    
    command = sys.argv[1].lower()
    lex = MeridianLexStratavore()
    
    try:
        if command == "status":
            status = lex.fleet_status()
            print(json.dumps(status, indent=2))
            
        elif command == "deploy":
            if len(sys.argv) < 3:
                print("Error: Personality required")
                return
            personality = sys.argv[2]
            agent_id = lex.deploy_agent(personality)
            if agent_id:
                print(f"Agent deployed: {agent_id}")
            else:
                print("Failed to deploy agent")
                
        elif command == "squad":
            if len(sys.argv) < 4:
                print("Error: Composition JSON and mission brief required")
                return
            composition = json.loads(sys.argv[2])
            mission = sys.argv[3]
            squad = lex.form_squad(composition, mission)
            print(json.dumps(squad, indent=2))
            
        elif command == "available":
            personality = sys.argv[2] if len(sys.argv) > 2 else None
            agents = lex.get_available_for_duty(personality)
            print(f"Available agents: {agents}")
            
        elif command == "recall":
            if len(sys.argv) < 3:
                print("Error: Agent ID required")
                return
            agent_id = sys.argv[2]
            success = lex.recall_agent(agent_id)
            print(f"Recall {'successful' if success else 'failed'}")
            
        elif command == "stop":
            if len(sys.argv) < 3:
                print("Error: Agent ID required")
                return
            agent_id = sys.argv[2]
            success = lex.emergency_stop(agent_id)
            print(f"Emergency stop {'successful' if success else 'failed'}")
            
        elif command == "report":
            report = lex.generate_mission_report()
            print(json.dumps(report, indent=2))
            
        else:
            print(f"Unknown command: {command}")
            
    except Exception as e:
        print(f"Command execution failed: {e}")


if __name__ == "__main__":
    main()