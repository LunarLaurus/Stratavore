#!/usr/bin/env python3
"""
Base Handler - Common functionality for all webui API handlers
Provides shared utilities, error handling, and response formatting
"""

import json
import time
import os
from datetime import datetime
from typing import Dict, Any, Optional

class BaseHandler:
    """Base class with common functionality for all API handlers"""
    
    def __init__(self):
        # Resolve paths relative to webui directory
        self.webui_dir = os.path.dirname(os.path.abspath(__file__))
        self.root_dir = os.path.dirname(self.webui_dir)
        
        # Common paths
        self.jobs_dir = os.path.join(self.root_dir, "jobs")
        self.agents_dir = os.path.join(self.root_dir, "agents")
        
        # File paths
        self.jobs_file = os.path.join(self.jobs_dir, "jobs.jsonl")
        self.progress_file = os.path.join(self.jobs_dir, "progress.json")
        self.time_sessions_file = os.path.join(self.jobs_dir, "time_sessions.jsonl")
        self.agents_file = os.path.join(self.agents_dir, "active_agents.jsonl")
        self.agent_todos_file = os.path.join(self.agents_dir, "agent_todos.jsonl")
    
    def _send_cors_headers(self, handler):
        """Write CORS headers (call between send_response and end_headers)."""
        handler.send_header("Access-Control-Allow-Origin", "*")
        handler.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        handler.send_header("Access-Control-Allow-Headers", "Content-Type")
    
    def _read_json_body(self, handler) -> Dict[str, Any]:
        """Read and parse the request body as JSON."""
        content_length = int(handler.headers.get("Content-Length", 0))
        raw = handler.rfile.read(content_length)
        return json.loads(raw.decode("utf-8")) if raw else {}
    
    def _respond_json(self, handler, status: int, payload: Dict[str, Any]):
        """Send JSON response with proper headers."""
        body = json.dumps(payload, indent=2).encode("utf-8")
        handler.send_response(status)
        self._send_cors_headers(handler)
        handler.send_header("Content-Type", "application/json; charset=utf-8")
        handler.send_header("Content-Length", str(len(body)))
        handler.end_headers()
        handler.wfile.write(body)
    
    def _respond_error(self, handler, status: int, message: str, details: Optional[Dict] = None):
        """Send standardized error response."""
        error_payload = {
            "status": "error",
            "error": message,
            "timestamp": time.time()
        }
        if details:
            error_payload["details"] = details
        
        self._respond_json(handler, status, error_payload)
    
    def _respond_success(self, handler, data: Dict[str, Any], message: str = "Success"):
        """Send standardized success response."""
        payload = {
            "status": "success",
            "message": message,
            "timestamp": time.time(),
            **data
        }
        self._respond_json(handler, 200, payload)
    
    def _load_jsonl(self, file_path: str) -> list:
        """Read a JSONL file and return a list of parsed objects. Never raises."""
        results = []
        try:
            with open(file_path, "r", encoding="utf-8") as fh:
                for line in fh:
                    line = line.strip()
                    if line:
                        try:
                            results.append(json.loads(line))
                        except json.JSONDecodeError as exc:
                            print(f"[WARN] Skipping malformed JSONL line in {file_path}: {exc}")
        except FileNotFoundError:
            pass
        except OSError as exc:
            print(f"[WARN] Could not read {file_path}: {exc}")
        return results
    
    def _load_prefixed_jsonl(self, file_path: str) -> dict:
        """
        Read a JSONL file where each line is:  <key> <json_object>
        Returns a dict keyed by the prefix token.
        Handles lines that are pure JSON objects (no prefix) gracefully.
        """
        results = {}
        try:
            with open(file_path, "r", encoding="utf-8") as fh:
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
                            print(f"[WARN] Skipping malformed line in {file_path}: {exc}")
                    else:
                        # Fallback: try to parse the whole line as JSON
                        try:
                            obj = json.loads(line)
                            if isinstance(obj, dict) and "id" in obj:
                                results[obj["id"]] = obj
                        except json.JSONDecodeError:
                            pass
        except FileNotFoundError:
            pass
        except OSError as exc:
            print(f"[WARN] Could not read {file_path}: {exc}")
        return results
    
    def _load_prefixed_jsonl_list(self, file_path: str) -> list:
        """Like _load_prefixed_jsonl but returns a list of the JSON values."""
        results = []
        try:
            with open(file_path, "r", encoding="utf-8") as fh:
                for line in fh:
                    line = line.strip()
                    if not line:
                        continue
                    parts = line.split(" ", 1)
                    if len(parts) == 2:
                        try:
                            results.append(json.loads(parts[1]))
                        except json.JSONDecodeError as exc:
                            print(f"[WARN] Skipping malformed line in {file_path}: {exc}")
                    else:
                        try:
                            obj = json.loads(line)
                            results.append(obj)
                        except json.JSONDecodeError:
                            pass
        except FileNotFoundError:
            pass
        except OSError as exc:
            print(f"[WARN] Could not read {file_path}: {exc}")
        return results
    
    def _get_agent_manager(self):
        """Import and return an AgentManager instance."""
        import sys
        if self.agents_dir not in sys.path:
            sys.path.insert(0, self.agents_dir)
        try:
            from agent_manager import AgentManager
            return AgentManager()
        except ImportError as exc:
            raise RuntimeError(f"Could not import AgentManager: {exc}")
    
    def _calculate_overview_metrics(self, jobs, time_sessions):
        """Calculate overview metrics from jobs and time sessions."""
        total = len(jobs)
        pending = len([j for j in jobs if j.get('status') == 'pending'])
        in_progress = len([j for j in jobs if j.get('status') == 'in_progress'])
        completed = len([j for j in jobs if j.get('status') == 'completed'])
        
        # Calculate tracked hours
        total_seconds = 0
        for session in time_sessions:
            if session.get('status') == 'completed' and session.get('duration_seconds'):
                total_seconds += session['duration_seconds']
            elif session.get('status') == 'active':
                total_seconds += (time.time() - session.get('start_timestamp', time.time()) - session.get('paused_time', 0))
        
        tracked_hours = round(total_seconds / 3600, 2)
        completion_rate = round((completed / total) * 100) if total > 0 else 0
        
        return {
            'totalJobs': total,
            'pendingJobs': pending,
            'inProgressJobs': in_progress,
            'completedJobs': completed,
            'completionRate': completion_rate,
            'totalTrackedHours': tracked_hours
        }
    
    def _calculate_priority_breakdown(self, jobs):
        """Calculate priority breakdown from jobs."""
        active_jobs = [j for j in jobs if j.get('status') != 'completed']
        
        breakdown = {
            'high': len([j for j in active_jobs if j.get('priority') == 'high']),
            'medium': len([j for j in active_jobs if j.get('priority') == 'medium']),
            'low': len([j for j in active_jobs if j.get('priority') == 'low']),
            'total': len(active_jobs)
        }
        
        return breakdown
    
    def _validate_required_fields(self, data: Dict[str, Any], required_fields: list) -> Optional[str]:
        """Validate that required fields are present in data."""
        for field in required_fields:
            if field not in data or not data[field] or str(data[field]).strip() == '':
                return f"Missing required field: {field}"
        return None
    
    def _get_request_metadata(self, handler) -> Dict[str, Any]:
        """Get metadata about the current request for logging."""
        return {
            'timestamp': time.time(),
            'client_address': handler.client_address[0] if hasattr(handler, 'client_address') else 'unknown',
            'method': getattr(handler, 'command', 'unknown'),
            'path': getattr(handler, 'path', 'unknown'),
            'user_agent': handler.headers.get('User-Agent', 'unknown') if hasattr(handler, 'headers') else 'unknown'
        }
    
    def _log_request(self, handler, message: str, level: str = "INFO"):
        """Log request information."""
        metadata = self._get_request_metadata(handler)
        timestamp = datetime.fromtimestamp(metadata['timestamp']).strftime('%Y-%m-%d %H:%M:%S')
        print(f"[{timestamp}] [{level}] {metadata['client_address']} {metadata['method']} {metadata['path']}: {message}")
    
    def handle_options(self, handler):
        """Handle CORS preflight requests."""
        handler.send_response(200)
        self._send_cors_headers(handler)
        handler.end_headers()