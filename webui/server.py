#!/usr/bin/env python3
"""
Modular HTTP server for Stratavore Job Tracker Web UI v4
Supports parallel agent workflows and enhanced API structure
"""

import http.server
import socketserver
import os
import json
import sys
import logging
import time
from urllib.parse import urlparse, parse_qs
from dataclasses import dataclass
from typing import Dict, List, Optional, Any
import threading

# Resolve the repo root relative to this file
_HERE = os.path.dirname(os.path.abspath(__file__))
_ROOT = os.path.dirname(_HERE)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global variable for server start time
_server_start_time = None


@dataclass
class APIResponse:
    """Standardized API response structure"""
    status: str
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    timestamp: Optional[float] = None
    
    def __post_init__(self):
        if self.timestamp is None:
            self.timestamp = time.time()


class DataLoader:
    """Handles all data loading operations"""
    
    @staticmethod
    def load_jsonl(path: str) -> List[Dict]:
        """Read a JSONL file and return a list of parsed objects"""
        results = []
        try:
            with open(path, "r", encoding="utf-8") as fh:
                for line in fh:
                    line = line.strip()
                    if line:
                        try:
                            results.append(json.loads(line))
                        except json.JSONDecodeError as exc:
                            logger.warning(f"Skipping malformed JSONL line in {path}: {exc}")
        except FileNotFoundError:
            logger.debug(f"File not found: {path}")
        except OSError as exc:
            logger.warning(f"Could not read {path}: {exc}")
        return results

    @staticmethod
    def load_prefixed_jsonl(path: str) -> Dict[str, Dict]:
        """Read a JSONL file with prefixed keys"""
        results = {}
        try:
            with open(path, "r", encoding="utf-8") as fh:
                for line in fh:
                    line = line.strip()
                    if not line:
                        continue
                    parts = line.split(" ", 1)
                    if len(parts) == 2:
                        key, payload = parts
                        try:
                            results[key] = json.loads(payload)
                        except json.JSONDecodeError as exc:
                            logger.warning(f"Skipping malformed line in {path}: {exc}")
                    else:
                        try:
                            obj = json.loads(line)
                            if isinstance(obj, dict) and "id" in obj:
                                results[obj["id"]] = obj
                        except json.JSONDecodeError:
                            pass
        except FileNotFoundError:
            logger.debug(f"File not found: {path}")
        except OSError as exc:
            logger.warning(f"Could not read {path}: {exc}")
        return results


class BaseAPIHandler:
    """Base class for API endpoint handlers"""
    
    def __init__(self, data_loader: DataLoader):
        self.data_loader = data_loader
    
    def handle_request(self, handler, path: str, query_params: Dict[str, str], body: Dict[str, Any]) -> APIResponse:
        """Handle API request - to be implemented by subclasses"""
        raise NotImplementedError("Subclasses must implement handle_request")


class StatusAPIHandler(BaseAPIHandler):
    """Handle status-related API endpoints"""
    
    def handle_request(self, handler, path: str, query_params: Dict[str, str], body: Dict[str, Any]) -> APIResponse:
        """Return comprehensive status data"""
        try:
            jobs = self.data_loader.load_jsonl(os.path.join(_ROOT, "jobs", "jobs.jsonl"))
            
            # Load progress data
            progress = {}
            progress_path = os.path.join(_ROOT, "jobs", "progress.json")
            try:
                with open(progress_path, "r", encoding="utf-8") as fh:
                    raw_progress = fh.read().strip()
                    if raw_progress.startswith("{") or raw_progress.startswith("["):
                        progress = json.loads(raw_progress)
            except (FileNotFoundError, json.JSONDecodeError):
                pass

            time_sessions = self.data_loader.load_jsonl(
                os.path.join(_ROOT, "jobs", "time_sessions.jsonl")
            )
            agents = self.data_loader.load_prefixed_jsonl(
                os.path.join(_ROOT, "agents", "active_agents.jsonl")
            )
            agent_todos = self.data_loader.load_jsonl(
                os.path.join(_ROOT, "agents", "agent_todos.jsonl")
            )

            data = {
                "jobs": jobs,
                "progress": progress,
                "time_sessions": time_sessions,
                "agents": agents,
                "agent_todos": agent_todos,
            }
            
            return APIResponse(status="success", data=data)
            
        except Exception as exc:
            logger.error(f"Error in status API: {exc}")
            return APIResponse(status="error", error=str(exc))


class HealthAPIHandler(BaseAPIHandler):
    """Handle health check endpoints"""
    
    def handle_request(self, handler, path: str, query_params: Dict[str, str], body: Dict[str, Any]) -> APIResponse:
        """Return system health information"""
        global _server_start_time
        uptime = time.time() - _server_start_time if _server_start_time else 0
        
        data = {
            "status": "healthy",
            "uptime": uptime,
            "version": "4.0.0",
            "components": {
                "data_loader": "healthy",
                "api_handlers": "healthy",
                "file_system": "healthy"
            }
        }
        return APIResponse(status="success", data=data)


class AgentsAPIHandler(BaseAPIHandler):
    """Handle agent-related API endpoints"""
    
    def handle_request(self, handler, path: str, query_params: Dict[str, str], body: Dict[str, Any]) -> APIResponse:
        """Return agent-focused data with summary"""
        try:
            agents = self.data_loader.load_prefixed_jsonl(
                os.path.join(_ROOT, "agents", "active_agents.jsonl")
            )
            agent_todos = self.data_loader.load_jsonl(
                os.path.join(_ROOT, "agents", "agent_todos.jsonl")
            )
            
            # Compute summary statistics
            summary = {
                "total": len(agents),
                "by_status": {},
                "by_personality": {},
            }
            for agent in agents.values():
                status = agent.get("status", "unknown")
                personality = agent.get("personality", "unknown")
                summary["by_status"][status] = summary["by_status"].get(status, 0) + 1
                summary["by_personality"][personality] = summary["by_personality"].get(personality, 0) + 1

            data = {
                "agents": agents,
                "agent_todos": agent_todos,
                "summary": summary,
            }
            
            return APIResponse(status="success", data=data)
            
        except Exception as exc:
            logger.error(f"Error in agents API: {exc}")
            return APIResponse(status="error", error=str(exc))


class AgentActionAPIHandler(BaseAPIHandler):
    """Handle agent action endpoints (spawn, assign, complete, etc.)"""
    
    def handle_request(self, handler, path: str, query_params: Dict[str, str], body: Dict[str, Any]) -> APIResponse:
        """Handle agent actions based on the path"""
        try:
            # Import agent manager
            agents_dir = os.path.join(_ROOT, "agents")
            if agents_dir not in sys.path:
                sys.path.insert(0, agents_dir)
            
            try:
                # Import agent_manager dynamically
                agent_manager = __import__('agent_manager')
                AgentManager = getattr(agent_manager, 'AgentManager')
                AgentPersonality = getattr(agent_manager, 'AgentPersonality')
                AgentStatus = getattr(agent_manager, 'AgentStatus')
                manager = AgentManager()
                # Store enum classes for fallback usage
                manager.AgentPersonality = AgentPersonality
                manager.AgentStatus = AgentStatus
            except ImportError as import_error:
                logger.error(f"Could not import agent_manager: {import_error}")
                # Create a dummy manager for error handling
                class DummyManager:
                    def spawn_agent(self, *args, **kwargs):
                        raise Exception("Agent manager not available")
                    def assign_task(self, *args, **kwargs):
                        return False
                    def complete_task(self, *args, **kwargs):
                        return False
                    def update_agent_status(self, *args, **kwargs):
                        raise Exception("Agent manager not available")
                    @property
                    def agents(self):
                        return {}
                manager = DummyManager()
            
            # Determine action based on path
            if "spawn-agent" in path:
                return self._handle_spawn_agent(manager, body)
            elif "assign-agent" in path:
                return self._handle_assign_agent(manager, body)
            elif "complete-task" in path:
                return self._handle_complete_task(manager, body)
            elif "agent-status" in path:
                return self._handle_agent_status(manager, body)
            elif "kill-agent" in path:
                return self._handle_kill_agent(manager, body)
            elif "batch-operation" in path:
                return self._handle_batch_operation(manager, body)
            else:
                return APIResponse(status="error", error="Unknown agent action")
                
        except ImportError as exc:
            logger.error(f"Could not import agent_manager: {exc}")
            return APIResponse(status="error", error="Agent manager not available")
        except Exception as exc:
            logger.error(f"Error in agent action API: {exc}")
            return APIResponse(status="error", error=str(exc))
    
    def _handle_spawn_agent(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle agent spawning"""
        personality = body.get("personality")
        if not personality:
            return APIResponse(status="error", error="No personality specified")

        try:
            AgentPersonality = manager.AgentPersonality
            personality_enum = AgentPersonality(personality.lower())
        except (ValueError, AttributeError):
            # Fallback if AgentPersonality not available
            valid = ['cadet', 'senior', 'specialist', 'researcher', 'debugger', 'optimizer']
            return APIResponse(status="error", error=f"Invalid personality {personality!r}. Valid: {valid}")

        task_id = body.get("task_id")
        try:
            agent_id = manager.spawn_agent(personality_enum, task_id)
        except Exception as exc:
            return APIResponse(status="error", error=f"Failed to spawn agent: {str(exc)}")
        
        data = {"agent_id": agent_id, "personality": personality}
        return APIResponse(status="success", data=data)
    
    def _handle_assign_agent(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle agent task assignment"""
        agent_id = body.get("agent_id", "").strip()
        task_id = body.get("task_id", "").strip()
        
        if not agent_id or not task_id:
            return APIResponse(status="error", error="agent_id and task_id are required")
        
        try:
            success = manager.assign_task(agent_id, task_id)
        except Exception as exc:
            return APIResponse(status="error", error=f"Failed to assign task: {str(exc)}")
            
        if success:
            data = {"agent_id": agent_id, "task_id": task_id}
            return APIResponse(status="success", data=data)
        else:
            return APIResponse(status="error", error=f"Could not assign task - agent {agent_id!r} not found or not available")
    
    def _handle_complete_task(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle task completion"""
        agent_id = body.get("agent_id", "").strip()
        if not agent_id:
            return APIResponse(status="error", error="agent_id is required")
        
        success = bool(body.get("success", True))
        notes = body.get("notes", "")
        
        try:
            ok = manager.complete_task(agent_id, success, notes)
        except Exception as exc:
            return APIResponse(status="error", error=f"Failed to complete task: {str(exc)}")
            
        if ok:
            data = {"agent_id": agent_id, "success": success}
            return APIResponse(status="success", data=data)
        else:
            return APIResponse(status="error", error=f"Agent {agent_id!r} not found")
    
    def _handle_agent_status(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle agent status updates"""
        agent_id = body.get("agent_id", "").strip()
        status_str = body.get("status", "").strip()
        
        if not agent_id or not status_str:
            return APIResponse(status="error", error="agent_id and status are required")

        try:
            AgentStatus = manager.AgentStatus
            status_enum = AgentStatus(status_str.lower())
        except (ValueError, AttributeError):
            # Fallback if AgentStatus not available
            valid = ['idle', 'working', 'paused', 'error', 'spawning', 'completed']
            return APIResponse(status="error", error=f"Invalid status {status_str!r}. Valid: {valid}")

        thought = body.get("thought")
        try:
            manager.update_agent_status(agent_id, status_enum, thought)
        except Exception as exc:
            return APIResponse(status="error", error=f"Failed to update status: {str(exc)}")
        
        data = {"agent_id": agent_id, "new_status": status_str}
        return APIResponse(status="success", data=data)
    
    def _handle_kill_agent(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle agent termination"""
        agent_id = body.get("agent_id", "").strip()
        if not agent_id:
            return APIResponse(status="error", error="agent_id is required")

        # Check if agent exists (may need to adapt based on manager structure)
        try:
            if hasattr(manager, 'agents') and agent_id not in manager.agents:
                return APIResponse(status="error", error=f"Agent {agent_id!r} not found")
            
            AgentStatus = manager.AgentStatus
            manager.update_agent_status(agent_id, AgentStatus.ERROR, "Killed via web UI")
        except (AttributeError, Exception) as exc:
            return APIResponse(status="error", error=f"Failed to kill agent: {str(exc)}")
        
        data = {"agent_id": agent_id}
        return APIResponse(status="success", data=data)
    
    def _handle_batch_operation(self, manager: Any, body: Dict[str, Any]) -> APIResponse:
        """Handle batch operations on multiple agents"""
        operation = body.get("operation")
        agent_ids = body.get("agent_ids", [])
        
        if not operation or not agent_ids:
            return APIResponse(status="error", error="operation and agent_ids are required")
        
        results = []
        
        for agent_id in agent_ids:
            try:
                if operation == "assign":
                    task_id = body.get("task_id")
                    success = manager.assign_task(agent_id, task_id)
                    results.append({"agent_id": agent_id, "success": success})
                    
                elif operation == "status":
                    status = body.get("status")
                    try:
                        AgentStatus = manager.AgentStatus
                        status_enum = AgentStatus(status)
                        manager.update_agent_status(agent_id, status_enum)
                        results.append({"agent_id": agent_id, "success": True})
                    except (ValueError, AttributeError):
                        results.append({"agent_id": agent_id, "success": False, "error": "Invalid status"})
                    
                elif operation == "kill":
                    try:
                        AgentStatus = manager.AgentStatus
                        manager.update_agent_status(agent_id, AgentStatus.ERROR, "Batch kill operation")
                        results.append({"agent_id": agent_id, "success": True})
                    except AttributeError:
                        results.append({"agent_id": agent_id, "success": False, "error": "Cannot kill agent"})
                    
                else:
                    results.append({"agent_id": agent_id, "success": False, "error": "Unknown operation"})
                    
            except Exception as exc:
                results.append({"agent_id": agent_id, "success": False, "error": str(exc)})
        
        data = {"operation": operation, "results": results}
        return APIResponse(status="success", data=data)


class StratavoreRequestHandler(http.server.SimpleHTTPRequestHandler):
    """Enhanced HTTP request handler with modular API routing"""
    
    def __init__(self, *args, **kwargs):
        self.data_loader = DataLoader()
        
        # Initialize API handlers
        self.api_handlers = {
            "/api/status": StatusAPIHandler(self.data_loader),
            "/api/health": HealthAPIHandler(self.data_loader),
            "/api/agents": AgentsAPIHandler(self.data_loader),
            "/api/spawn-agent": AgentActionAPIHandler(self.data_loader),
            "/api/assign-agent": AgentActionAPIHandler(self.data_loader),
            "/api/complete-task": AgentActionAPIHandler(self.data_loader),
            "/api/agent-status": AgentActionAPIHandler(self.data_loader),
            "/api/kill-agent": AgentActionAPIHandler(self.data_loader),
            "/api/batch-operation": AgentActionAPIHandler(self.data_loader),
        }
        
        super().__init__(*args, **kwargs)

    def _send_cors_headers(self):
        """Write CORS headers"""
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")

    def do_GET(self):
        parsed_path = urlparse(self.path)
        
        if parsed_path.path.startswith("/api/"):
            self._handle_api_request("GET", parsed_path)
        else:
            # Serve static files
            super().do_GET()

    def do_POST(self):
        parsed_path = urlparse(self.path)
        
        if parsed_path.path.startswith("/api/"):
            self._handle_api_request("POST", parsed_path)
        else:
            self.send_error(404, f"Unknown endpoint: {parsed_path.path}")

    def do_OPTIONS(self):
        """Handle CORS preflight requests"""
        self.send_response(200)
        self._send_cors_headers()
        self.end_headers()

    def _handle_api_request(self, method: str, parsed_path):
        """Route API requests to appropriate handlers"""
        try:
            # Parse query parameters
            query_params = dict(parse_qs(parsed_path.query))
            # Flatten single-value parameters
            query_params = {k: v[0] if len(v) == 1 else v for k, v in query_params.items()}
            
            # Read request body for POST requests
            body = {}
            if method == "POST":
                content_length = int(self.headers.get("Content-Length", 0))
                if content_length > 0:
                    raw_body = self.rfile.read(content_length)
                    try:
                        body = json.loads(raw_body.decode("utf-8"))
                    except json.JSONDecodeError as exc:
                        self._send_json_response(400, APIResponse(
                            status="error", 
                            error=f"Bad JSON: {exc}"
                        ).__dict__)
                        return
            
            # Find matching handler
            handler = None
            for path_pattern, api_handler in self.api_handlers.items():
                if parsed_path.path == path_pattern or parsed_path.path.startswith(path_pattern):
                    handler = api_handler
                    break
            
            if not handler:
                self.send_error(404, f"Unknown API endpoint: {parsed_path.path}")
                return
            
            # Handle the request
            response = handler.handle_request(self, parsed_path.path, query_params, body)
            self._send_json_response(200 if response.status == "success" else 400, response.__dict__)
            
        except Exception as exc:
            logger.error(f"Unhandled error in API request: {exc}")
            self._send_json_response(500, APIResponse(
                status="error", 
                error="Internal server error"
            ).__dict__)

    def _send_json_response(self, status_code: int, data: Dict[str, Any]):
        """Send JSON response"""
        body = json.dumps(data, indent=2).encode("utf-8")
        self.send_response(status_code)
        self._send_cors_headers()
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        """Override log message for cleaner output"""
        logger.info(format % args)


class StratavoreServer:
    """Main server class with enhanced configuration"""
    
    def __init__(self, port: int = 8080, host: str = ""):
        self.port = port
        self.host = host
        self.httpd = None
        
    def start(self):
        """Start the web server"""
        try:
            # Change to webui directory for static file serving
            os.chdir(_HERE)
            
            # Create HTTP server with custom handler
            self.httpd = socketserver.TCPServer((self.host, self.port), StratavoreRequestHandler)
            self.httpd.allow_reuse_address = True
            
            # Store start time for health checks using module-level variable
            global _server_start_time
            _server_start_time = time.time()
            
            logger.info(f"Stratavore Job Tracker UI v4 running on http://{self.host or 'localhost'}:{self.port}")
            logger.info("API endpoints: /api/status, /api/health, /api/agents, /api/spawn-agent, etc.")
            logger.info("Press Ctrl+C to stop")
            
            # Start server in main thread
            self.httpd.serve_forever()
            
        except KeyboardInterrupt:
            logger.info("Server stopped by user")
            self.stop()
        except Exception as exc:
            logger.error(f"Failed to start server: {exc}")
            raise
    
    def stop(self):
        """Stop the web server"""
        if self.httpd:
            logger.info("Stopping server...")
            self.httpd.shutdown()
            self.httpd.server_close()
            logger.info("Server stopped")


def main():
    """Main entry point"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Stratavore Web UI server v4")
    parser.add_argument(
        "--port", type=int, default=8080, help="Port to listen on (default: 8080)"
    )
    parser.add_argument(
        "--host", type=str, default="", help="Host to bind to (default: all interfaces)"
    )
    parser.add_argument(
        "--debug", action="store_true", help="Enable debug logging"
    )
    
    args = parser.parse_args()
    
    # Configure logging level
    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)
        logger.debug("Debug mode enabled")
    
    # Create and start server
    server = StratavoreServer(port=args.port, host=args.host)
    server.start()


if __name__ == "__main__":
    main()