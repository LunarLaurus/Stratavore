#!/usr/bin/env python3
"""
Enhanced HTTP server for Stratavore Job Tracker Web UI
Supports parallel agent workflows and agent management
"""

import http.server
import socketserver
import os
import json
import sys
from urllib.parse import urlparse
import time

# Resolve the repo root relative to this file so paths work regardless of
# where the server is launched from.
_HERE = os.path.dirname(os.path.abspath(__file__))
_ROOT = os.path.dirname(_HERE)


def _load_jsonl(path: str) -> list:
    """Read a JSONL file and return a list of parsed objects. Never raises."""
    results = []
    try:
        with open(path, "r", encoding="utf-8") as fh:
            for line in fh:
                line = line.strip()
                if line:
                    try:
                        results.append(json.loads(line))
                    except json.JSONDecodeError as exc:
                        print(f"[WARN] Skipping malformed JSONL line in {path}: {exc}")
    except FileNotFoundError:
        pass
    except OSError as exc:
        print(f"[WARN] Could not read {path}: {exc}")
    return results


def _load_prefixed_jsonl(path: str) -> dict:
    """
    Read a JSONL file where each line is:  <key> <json_object>
    Returns a dict keyed by the prefix token.
    Handles lines that are pure JSON objects (no prefix) gracefully.
    """
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
                        print(f"[WARN] Skipping malformed line in {path}: {exc}")
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
        print(f"[WARN] Could not read {path}: {exc}")
    return results


def _load_prefixed_jsonl_list(path: str) -> list:
    """Like _load_prefixed_jsonl but returns a list of the JSON values."""
    results = []
    try:
        with open(path, "r", encoding="utf-8") as fh:
            for line in fh:
                line = line.strip()
                if not line:
                    continue
                parts = line.split(" ", 1)
                if len(parts) == 2:
                    try:
                        results.append(json.loads(parts[1]))
                    except json.JSONDecodeError as exc:
                        print(f"[WARN] Skipping malformed line in {path}: {exc}")
                else:
                    try:
                        obj = json.loads(line)
                        results.append(obj)
                    except json.JSONDecodeError:
                        pass
    except FileNotFoundError:
        pass
    except OSError as exc:
        print(f"[WARN] Could not read {path}: {exc}")
    return results


class JobTrackerHandler(http.server.SimpleHTTPRequestHandler):

    def _send_cors_headers(self):
        """Write CORS headers (call between send_response and end_headers)."""
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")

    def do_GET(self):
        parsed_path = urlparse(self.path)

        if parsed_path.path == "/api/status":
            self._handle_api_status()
        elif parsed_path.path == "/api/health":
            self._handle_api_health()
        else:
            # Let SimpleHTTPRequestHandler serve static files.
            # Do NOT call send_response / send_header before this – the parent
            # manages its own response lifecycle.
            super().do_GET()

    def do_POST(self):
        parsed_path = urlparse(self.path)

        if parsed_path.path == "/api/spawn-agent":
            self._handle_spawn_agent()
        else:
            self.send_error(404, f"Unknown endpoint: {parsed_path.path}")

    def do_OPTIONS(self):
        """Handle CORS preflight requests."""
        self.send_response(200)
        self._send_cors_headers()
        self.end_headers()

    # --- API handlers ---------------------------------------------------------

    def _handle_api_status(self):
        try:
            jobs = _load_jsonl(os.path.join(_ROOT, "jobs", "jobs.jsonl"))

            progress = {}
            progress_path = os.path.join(_ROOT, "jobs", "progress.json")
            try:
                with open(progress_path, "r", encoding="utf-8") as fh:
                    progress = json.load(fh)
            except (FileNotFoundError, json.JSONDecodeError) as exc:
                print(f"[WARN] Could not read progress.json: {exc}")

            time_sessions = _load_jsonl(
                os.path.join(_ROOT, "jobs", "time_sessions.jsonl")
            )
            agents = _load_prefixed_jsonl(
                os.path.join(_ROOT, "agents", "active_agents.jsonl")
            )
            agent_todos = _load_prefixed_jsonl_list(
                os.path.join(_ROOT, "agents", "agent_todos.jsonl")
            )

            response = {
                "jobs": jobs,
                "progress": progress,
                "time_sessions": time_sessions,
                "agents": agents,
                "agent_todos": agent_todos,
                "timestamp": time.time(),
                "status": "success",
            }
            self._respond_json(200, response)

        except Exception as exc:
            self._respond_json(
                500,
                {"status": "error", "error": str(exc), "timestamp": time.time()},
            )

    def _handle_api_health(self):
        self._respond_json(
            200,
            {
                "status": "healthy",
                "timestamp": time.time(),
                "uptime": time.time() - _START_TIME,
            },
        )

    def _handle_spawn_agent(self):
        try:
            content_length = int(self.headers.get("Content-Length", 0))
            raw = self.rfile.read(content_length)
            data = json.loads(raw.decode("utf-8"))

            personality = data.get("personality")
            if not personality:
                self._respond_json(
                    400, {"status": "error", "error": "No personality specified"}
                )
                return

            agents_dir = os.path.join(_ROOT, "agents")
            if agents_dir not in sys.path:
                sys.path.insert(0, agents_dir)

            from agent_manager import AgentManager, AgentPersonality  # type: ignore

            manager = AgentManager()
            try:
                personality_enum = AgentPersonality(personality.lower())
            except ValueError:
                self._respond_json(
                    400,
                    {"status": "error", "error": f"Invalid personality: {personality}"},
                )
                return

            agent_id = manager.spawn_agent(personality_enum)
            self._respond_json(
                200,
                {
                    "status": "success",
                    "agent_id": agent_id,
                    "personality": personality,
                },
            )

        except json.JSONDecodeError as exc:
            self._respond_json(400, {"status": "error", "error": f"Bad JSON: {exc}"})
        except Exception as exc:
            self._respond_json(500, {"status": "error", "error": str(exc)})

    # --- Helpers --------------------------------------------------------------

    def _respond_json(self, status: int, payload: dict):
        body = json.dumps(payload, indent=2).encode("utf-8")
        self.send_response(status)
        self._send_cors_headers()
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] {format % args}")


# Global server start time (used by /api/health)
_START_TIME = time.time()


def start_server(port: int = 8080):
    """Start the web UI server."""
    # Serve static files from the webui directory
    os.chdir(_HERE)

    with socketserver.TCPServer(("", port), JobTrackerHandler) as httpd:
        httpd.allow_reuse_address = True
        print(f"Stratavore Job Tracker UI running on http://localhost:{port}")
        print("API endpoints: /api/status, /api/health, /api/spawn-agent")
        print("Press Ctrl+C to stop")
        print(f"Server started at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nServer stopped")


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Stratavore Web UI server")
    parser.add_argument(
        "--port", type=int, default=8080, help="Port to listen on (default: 8080)"
    )
    args = parser.parse_args()
    start_server(args.port)
