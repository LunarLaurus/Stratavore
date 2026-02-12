#!/usr/bin/env python3
"""
Enhanced Modular HTTP Server for Stratavore Job Tracker Web UI v4
Supports parallel agent workflows with unlimited scaling through modular architecture
"""

import http.server
import socketserver
import os
import json
import sys
import time
from urllib.parse import urlparse

# Resolve the repo root relative to this file so paths work regardless of
# where the server is launched from.
_HERE = os.path.dirname(os.path.abspath(__file__))
_ROOT = os.path.dirname(_HERE)

# Import modular handlers
from backend.handlers.base_handler import BaseHandler
from backend.handlers.status_handler import StatusHandler

class ModularJobTrackerHandler(http.server.SimpleHTTPRequestHandler):
    """Modular HTTP handler with clean separation of concerns"""

    def __init__(self, *args, **kwargs):
        # Initialize handler instances
        self.base_handler = BaseHandler()
        self.status_handler = StatusHandler()
        
        super().__init__(*args, **kwargs)

    def do_GET(self):
        """Handle GET requests with modular routing"""
        parsed_path = urlparse(self.path)

        if parsed_path.path == "/api/status":
            self._handle_status_endpoint()
        elif parsed_path.path == "/api/health":
            self._handle_health_endpoint()
        elif parsed_path.path == "/api/agents":
            self._handle_agents_endpoint()
        elif parsed_path.path == "/api/metrics":
            self._handle_metrics_endpoint()
        elif parsed_path.path.startswith("/api/"):
            self._handle_unknown_endpoint("GET", parsed_path.path)
        else:
            # Serve static files for webui
            super().do_GET()

    def do_POST(self):
        """Handle POST requests with modular routing"""
        parsed_path = urlparse(self.path)

        if parsed_path.path == "/api/spawn-agent":
            self._handle_spawn_agent_endpoint()
        elif parsed_path.path == "/api/assign-agent":
            self._handle_assign_agent_endpoint()
        elif parsed_path.path == "/api/complete-task":
            self._handle_complete_task_endpoint()
        elif parsed_path.path == "/api/agent-status":
            self._handle_agent_status_endpoint()
        elif parsed_path.path == "/api/kill-agent":
            self._handle_kill_agent_endpoint()
        elif parsed_path.path.startswith("/api/"):
            self._handle_unknown_endpoint("POST", parsed_path.path)
        else:
            self.send_error(404, f"Unknown endpoint: {parsed_path.path}")

    def do_OPTIONS(self):
        """Handle CORS preflight requests"""
        self.base_handler.handle_options(self)

    # --- Endpoint Handlers ----------------------------------------------

    def _handle_status_endpoint(self):
        """Delegate to status handler"""
        try:
            self.status_handler.handle_status(self)
        except Exception as exc:
            self.base_handler._respond_error(self, 500, f"Status endpoint error: {str(exc)}")

    def _handle_health_endpoint(self):
        """Delegate to status handler"""
        try:
            self.status_handler.handle_health(self)
        except Exception as exc:
            self.base_handler._respond_error(self, 500, f"Health endpoint error: {str(exc)}")

    def _handle_agents_endpoint(self):
        """Delegate to status handler"""
        try:
            self.status_handler.handle_agents(self)
        except Exception as exc:
            self.base_handler._respond_error(self, 500, f"Agents endpoint error: {str(exc)}")

    def _handle_metrics_endpoint(self):
        """Handle API metrics and performance data"""
        try:
            # Calculate server metrics
            metrics = {
                "status": "success",
                "timestamp": time.time(),
                "server": {
                    "uptime_seconds": time.time() - self.status_handler.server_start_time,
                    "python_version": sys.version,
                    "platform": os.name
                },
                "api": {
                    "endpoints_count": 6,
                    "handlers_count": 2,
                    "architecture": "modular"
                },
                "performance": {
                    "avg_response_time_ms": self._calculate_avg_response_time(),
                    "total_requests": getattr(self, '_request_count', 0)
                },
                "resources": {
                    "memory_usage_mb": self._get_memory_usage(),
                    "disk_space_mb": self._get_disk_space()
                }
            }
            
            self.base_handler._respond_success(self, metrics, "Metrics retrieved successfully")
            
        except Exception as exc:
            self.base_handler._respond_error(self, 500, f"Metrics endpoint error: {str(exc)}")

    def _handle_spawn_agent_endpoint(self):
        """Handle agent spawning requests"""
        try:
            data = self.base_handler._read_json_body(self)
            personality = data.get("personality")
            task_id = data.get("task_id", None)
            
            if not personality:
                self.base_handler._respond_error(self, 400, "No personality specified")
                return

            # Validate personality
            try:
                manager = self.base_handler._get_agent_manager()
                from agent_manager import AgentPersonality
                personality_enum = AgentPersonality(personality.lower())
            except (ImportError, ValueError) as exc:
                valid = ["cadet", "senior", "specialist", "researcher", "debugger", "optimizer"]
                self.base_handler._respond_error(self, 400, f"Invalid personality {personality!r}. Valid: {valid}")
                return

            agent_id = manager.spawn_agent(personality_enum, task_id)
            self.base_handler._respond_success(self, {
                "agent_id": agent_id, 
                "personality": personality
            }, f"Agent {agent_id} spawned successfully")

        except json.JSONDecodeError as exc:
            self.base_handler._respond_error(self, 400, f"Bad JSON: {exc}")
        except Exception as exc:
            self.base_handler._respond_error(self, 500, str(exc))

    def _handle_assign_agent_endpoint(self):
        """Handle agent task assignment"""
        try:
            data = self.base_handler._read_json_body(self)
            agent_id = data.get("agent_id", "").strip()
            task_id = data.get("task_id", "").strip()
            
            error_msg = self.base_handler._validate_required_fields(data, ["agent_id", "task_id"])
            if error_msg:
                self.base_handler._respond_error(self, 400, error_msg)
                return
            
            manager = self.base_handler._get_agent_manager()
            success = manager.assign_task(agent_id, task_id)
            
            if success:
                self.base_handler._respond_success(self, {"agent_id": agent_id, "task_id": task_id}, 
                                                 f"Task {task_id} assigned to agent {agent_id}")
            else:
                self.base_handler._respond_error(self, 400, 
                    f"Could not assign task - agent {agent_id!r} not found or not available")

        except Exception as exc:
            self.base_handler._respond_error(self, 500, str(exc))

    def _handle_complete_task_endpoint(self):
        """Handle agent task completion"""
        try:
            data = self.base_handler._read_json_body(self)
            agent_id = data.get("agent_id", "").strip()
            
            if not agent_id:
                self.base_handler._respond_error(self, 400, "agent_id is required")
                return
            
            success = bool(data.get("success", True))
            notes = data.get("notes", "")
            
            manager = self.base_handler._get_agent_manager()
            ok = manager.complete_task(agent_id, success, notes)
            
            if ok:
                self.base_handler._respond_success(self, {"agent_id": agent_id, "success": success},
                                                 f"Task marked {success and 'complete' or 'failed'} for {agent_id}")
            else:
                self.base_handler._respond_error(self, 400, f"Agent {agent_id!r} not found")

        except Exception as exc:
            self.base_handler._respond_error(self, 500, str(exc))

    def _handle_agent_status_endpoint(self):
        """Handle agent status updates"""
        try:
            data = self.base_handler._read_json_body(self)
            agent_id = data.get("agent_id", "").strip()
            status_str = data.get("status", "").strip()
            
            error_msg = self.base_handler._validate_required_fields(data, ["agent_id", "status"])
            if error_msg:
                self.base_handler._respond_error(self, 400, error_msg)
                return

            # Validate status
            try:
                manager = self.base_handler._get_agent_manager()
                from agent_manager import AgentStatus
                status_enum = AgentStatus(status_str.lower())
            except (ImportError, ValueError):
                valid = ["idle", "working", "paused", "completed", "error", "spawning"]
                self.base_handler._respond_error(self, 400, f"Invalid status {status_str!r}. Valid: {valid}")
                return

            thought = data.get("thought", None)
            manager.update_agent_status(agent_id, status_enum, thought)
            self.base_handler._respond_success(self, {"agent_id": agent_id, "new_status": status_str},
                                             f"Agent {agent_id} status updated to {status_str}")

        except Exception as exc:
            self.base_handler._respond_error(self, 500, str(exc))

    def _handle_kill_agent_endpoint(self):
        """Handle agent termination"""
        try:
            data = self.base_handler._read_json_body(self)
            agent_id = data.get("agent_id", "").strip()
            
            if not agent_id:
                self.base_handler._respond_error(self, 400, "agent_id is required")
                return

            manager = self.base_handler._get_agent_manager()
            if agent_id not in manager.agents:
                self.base_handler._respond_error(self, 404, f"Agent {agent_id!r} not found")
                return
            
            from agent_manager import AgentStatus
            manager.update_agent_status(agent_id, AgentStatus.ERROR, "Killed via web UI")
            self.base_handler._respond_success(self, {"agent_id": agent_id}, 
                                             f"Agent {agent_id} terminated successfully")

        except Exception as exc:
            self.base_handler._respond_error(self, 500, str(exc))

    def _handle_unknown_endpoint(self, method, path):
        """Handle unknown API endpoints"""
        available_endpoints = [
            "GET /api/status",
            "GET /api/health", 
            "GET /api/agents",
            "GET /api/metrics",
            "POST /api/spawn-agent",
            "POST /api/assign-agent",
            "POST /api/complete-task",
            "POST /api/agent-status",
            "POST /api/kill-agent"
        ]
        
        self.base_handler._respond_error(self, 404, f"Unknown endpoint: {method} {path}. Available: {', '.join(available_endpoints)}")

    # --- Utility Methods ------------------------------------------------

    def _calculate_avg_response_time(self):
        """Calculate average response time (simplified)"""
        # In a real implementation, this would track actual response times
        return round(time.time() % 100, 2)  # Mock value for demo

    def _get_memory_usage(self):
        """Get current memory usage in MB"""
        try:
            import psutil
            process = psutil.Process()
            return round(process.memory_info().rss / 1024 / 1024, 2)
        except ImportError:
            return 0  # psutil not available

    def _get_disk_space(self):
        """Get available disk space in MB"""
        try:
            import shutil
            _, _, free_bytes = shutil.disk_usage(".")
            return round(free_bytes / 1024 / 1024, 2)
        except Exception:
            return 0

    def log_message(self, format, *args):
        """Enhanced logging with timestamps and request tracking"""
        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        
        # Track request count
        if not hasattr(self, '_request_count'):
            self._request_count = 0
        self._request_count += 1
        
        # Enhanced log format
        client_ip = self.client_address[0] if hasattr(self, 'client_address') else 'unknown'
        method = getattr(self, 'command', 'unknown')
        path = getattr(self, 'path', 'unknown')
        
        print(f"[{timestamp}] [{self._request_count}] {client_ip} {method} {path} - {format % args}")


# Global server configuration
_START_TIME = time.time()
_SERVER_INFO = {
    "version": "4.0.0",
    "architecture": "modular", 
    "start_time": _START_TIME,
    "features": ["unlimited_scaling", "modular_handlers", "real_time_updates", "error_boundaries"]
}


def start_server(port: int = 8080):
    """Start the modular web UI server"""
    # Serve static files from the webui directory
    os.chdir(_HERE)
    
    # Configure server
    handler = ModularJobTrackerHandler
    
    with socketserver.TCPServer(("", port), handler) as httpd:
        httpd.allow_reuse_address = True
        
        print(f"🚀 Stratavore Job Tracker UI v4 (Modular) running on http://localhost:{port}")
        print(f"📡 Architecture: {_SERVER_INFO['architecture']}")
        print(f"🔧 Features: {', '.join(_SERVER_INFO['features'])}")
        print("🌐 API endpoints:")
        print("   GET  /api/status - Main data aggregation")
        print("   GET  /api/health - System health check")
        print("   GET  /api/agents - Agent status and metrics")
        print("   GET  /api/metrics - Performance metrics")
        print("   POST /api/spawn-agent - Create new agent")
        print("   POST /api/assign-agent - Assign task to agent")
        print("   POST /api/complete-task - Mark task complete")
        print("   POST /api/agent-status - Update agent status")
        print("   POST /api/kill-agent - Terminate agent")
        print("⚡ Press Ctrl+C to stop")
        print(f"🕐 Server started at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n🛑 Server stopped gracefully")


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Stratavore Web UI Server v4 (Modular)")
    parser.add_argument(
        "--port", type=int, default=8080, help="Port to listen on (default: 8080)"
    )
    args = parser.parse_args()
    
    start_server(args.port)