#!/usr/bin/env python3
"""
Status Handler - API endpoints for data retrieval and aggregation
Handles /api/status, /api/health, and /api/agents endpoints
"""

import time
import json
from .base_handler import BaseHandler

class StatusHandler(BaseHandler):
    """Handler for status and data retrieval endpoints"""
    
    def __init__(self):
        super().__init__()
        self.server_start_time = time.time()
    
    def handle_status(self, handler):
        """Handle GET /api/status - Main data aggregation endpoint"""
        try:
            self._log_request(handler, "Processing status request")
            
            # Load all data sources
            jobs = self._load_jsonl(self.jobs_file)
            
            # Load progress data
            progress = {}
            try:
                with open(self.progress_file, "r", encoding="utf-8") as fh:
                    raw_progress = fh.read().strip()
                    if raw_progress.startswith("{") or raw_progress.startswith("["):
                        progress = json.loads(raw_progress)
            except (FileNotFoundError, json.JSONDecodeError):
                pass
            
            # Load time sessions
            time_sessions = self._load_jsonl(self.time_sessions_file)
            
            # Load agent data
            agents = self._load_prefixed_jsonl(self.agents_file)
            agent_todos = self._load_prefixed_jsonl_list(self.agent_todos_file)
            
            # Calculate metrics
            overview = self._calculate_overview_metrics(jobs, time_sessions)
            priority_breakdown = self._calculate_priority_breakdown(jobs)
            
            # Build response
            response_data = {
                "jobs": jobs,
                "progress": progress,
                "time_sessions": time_sessions,
                "agents": agents,
                "agent_todos": agent_todos,
                "overview": overview,
                "priority_breakdown": priority_breakdown,
                "timestamp": time.time(),
                "status": "success",
                "performance": {
                    "data_load_time_ms": self._measure_data_load_time(),
                    "active_agents": len(agents),
                    "active_sessions": len([s for s in time_sessions if s.get('status') == 'active'])
                }
            }
            
            self._respond_success(handler, response_data, "Data loaded successfully")
            self._log_request(handler, f"Status successful: {len(jobs)} jobs, {len(agents)} agents")
            
        except Exception as exc:
            error_msg = f"Status request failed: {str(exc)}"
            self._log_request(handler, error_msg, "ERROR")
            self._respond_error(handler, 500, error_msg, {"exception_type": type(exc).__name__})
    
    def handle_health(self, handler):
        """Handle GET /api/health - System health check"""
        try:
            self._log_request(handler, "Processing health check")
            
            # Calculate uptime
            uptime = time.time() - self.server_start_time
            
            # Check data file accessibility
            file_checks = {
                'jobs_file': self._check_file_accessible(self.jobs_file),
                'progress_file': self._check_file_accessible(self.progress_file),
                'time_sessions_file': self._check_file_accessible(self.time_sessions_file),
                'agents_file': self._check_file_accessible(self.agents_file),
                'agent_todos_file': self._check_file_accessible(self.agent_todos_file)
            }
            
            # Check agent manager availability
            agent_manager_healthy = self._check_agent_manager_health()
            
            # System metrics
            system_metrics = {
                'uptime_seconds': uptime,
                'uptime_formatted': self._format_uptime(uptime),
                'file_accessibility': file_checks,
                'agent_manager_healthy': agent_manager_healthy,
                'memory_usage_mb': self._get_memory_usage(),
                'disk_space_mb': self._get_disk_space()
            }
            
            # Determine overall health
            all_files_healthy = all(check['accessible'] for check in file_checks.values())
            overall_healthy = all_files_healthy and agent_manager_healthy
            
            response_data = {
                "status": "healthy" if overall_healthy else "unhealthy",
                "timestamp": time.time(),
                "uptime": uptime,
                "system_metrics": system_metrics,
                "checks": {
                    "files_accessible": all_files_healthy,
                    "agent_manager": agent_manager_healthy
                }
            }
            
            status_code = 200 if overall_healthy else 503
            self._respond_json(handler, status_code, response_data)
            self._log_request(handler, f"Health check: {response_data['status']}")
            
        except Exception as exc:
            error_msg = f"Health check failed: {str(exc)}"
            self._log_request(handler, error_msg, "ERROR")
            self._respond_error(handler, 500, error_msg)
    
    def handle_agents(self, handler):
        """Handle GET /api/agents - Agent-focused data with summary"""
        try:
            self._log_request(handler, "Processing agents request")
            
            # Load agent data
            agents = self._load_prefixed_jsonl(self.agents_file)
            agent_todos = self._load_prefixed_jsonl_list(self.agent_todos_file)
            
            # Calculate agent summary statistics
            summary = {
                "total": len(agents),
                "by_status": {},
                "by_personality": {},
                "working_count": 0,
                "idle_count": 0,
                "error_count": 0
            }
            
            for agent_id, agent_data in agents.items():
                # Status breakdown
                status = agent_data.get("status", "unknown")
                summary["by_status"][status] = summary["by_status"].get(status, 0) + 1
                
                if status == "working":
                    summary["working_count"] += 1
                elif status == "idle":
                    summary["idle_count"] += 1
                elif status == "error":
                    summary["error_count"] += 1
                
                # Personality breakdown
                personality = agent_data.get("personality", "unknown")
                summary["by_personality"][personality] = summary["by_personality"].get(personality, 0) + 1
            
            # Calculate todo statistics
            todo_stats = {
                "total": len(agent_todos),
                "completed": len([t for t in agent_todos if t.get("status") == "completed"]),
                "in_progress": len([t for t in agent_todos if t.get("status") == "in_progress"]),
                "pending": len([t for t in agent_todos if t.get("status") == "pending"]),
                "cancelled": len([t for t in agent_todos if t.get("status") == "cancelled"])
            }
            
            # Agent performance metrics
            performance_metrics = {}
            for agent_id, agent_data in agents.items():
                metrics = agent_data.get("metrics", {})
                if metrics:
                    performance_metrics[agent_id] = {
                        "tasks_completed": metrics.get("tasks_completed", 0),
                        "total_work_time": metrics.get("total_work_time", 0),
                        "average_task_time": metrics.get("average_task_time", 0)
                    }
            
            response_data = {
                "status": "success",
                "timestamp": time.time(),
                "agents": agents,
                "agent_todos": agent_todos,
                "summary": summary,
                "todo_stats": todo_stats,
                "performance_metrics": performance_metrics,
                "agent_utilization": {
                    "utilization_rate": round((summary["working_count"] / summary["total"]) * 100, 1) if summary["total"] > 0 else 0
                }
            }
            
            self._respond_success(handler, response_data, f"Agent data loaded: {summary['total']} agents")
            self._log_request(handler, f"Agents request successful: {summary['total']} agents, {todo_stats['total']} todos")
            
        except Exception as exc:
            error_msg = f"Agents request failed: {str(exc)}"
            self._log_request(handler, error_msg, "ERROR")
            self._respond_error(handler, 500, error_msg)
    
    def _measure_data_load_time(self):
        """Measure the time it takes to load all data files"""
        start_time = time.time()
        
        # Simulate loading all data files
        self._load_jsonl(self.jobs_file)
        self._load_prefixed_jsonl(self.agents_file)
        self._load_jsonl(self.time_sessions_file)
        self._load_prefixed_jsonl_list(self.agent_todos_file)
        
        return round((time.time() - start_time) * 1000, 2)  # Convert to milliseconds
    
    def _check_file_accessible(self, file_path):
        """Check if a file is accessible and readable"""
        try:
            with open(file_path, 'r') as f:
                f.read(1)  # Try to read first byte
            return {"accessible": True, "error": None}
        except Exception as exc:
            return {"accessible": False, "error": str(exc)}
    
    def _check_agent_manager_health(self):
        """Check if agent manager is importable and functional"""
        try:
            manager = self._get_agent_manager()
            # Try to get basic info from manager
            return hasattr(manager, 'agents') and hasattr(manager, 'load_agent_data')
        except Exception:
            return False
    
    def _format_uptime(self, seconds):
        """Format uptime in human readable format"""
        days = int(seconds // 86400)
        hours = int((seconds % 86400) // 3600)
        minutes = int((seconds % 3600) // 60)
        
        if days > 0:
            return f"{days}d {hours}h {minutes}m"
        elif hours > 0:
            return f"{hours}h {minutes}m"
        else:
            return f"{minutes}m"
    
    def _get_memory_usage(self):
        """Get current memory usage in MB (simplified)"""
        try:
            import psutil
            process = psutil.Process()
            return round(process.memory_info().rss / 1024 / 1024, 2)
        except ImportError:
            return 0  # psutil not available
    
    def _get_disk_space(self):
        """Get available disk space in MB (simplified)"""
        try:
            import shutil
            _, _, free_bytes = shutil.disk_usage(".")
            return round(free_bytes / 1024 / 1024, 2)
        except Exception:
            return 0