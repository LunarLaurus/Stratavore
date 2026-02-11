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
import threading
import time

class JobTrackerHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        parsed_path = urlparse(self.path)
        
        # CORS headers for local development
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        
        if parsed_path.path == '/api/status':
            # Serve JSON API for live data
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            try:
                # Load jobs data with error handling
                jobs = []
                try:
                    with open('../jobs/jobs.jsonl', 'r') as f:
                        content = f.read().strip()
                        if content:
                            jobs = [json.loads(line) for line in content.split('\n') if line.strip()]
                except (FileNotFoundError, json.JSONDecodeError) as e:
                    print(f"Error loading jobs: {e}")
                
                # Load progress data
                progress = {}
                try:
                    with open('../jobs/progress.json', 'r') as f:
                        progress = json.load(f)
                except (FileNotFoundError, json.JSONDecodeError) as e:
                    print(f"Error loading progress: {e}")
                
                # Load time sessions data
                time_sessions = []
                try:
                    with open('../jobs/time_sessions.jsonl', 'r') as f:
                        content = f.read().strip()
                        if content:
                            time_sessions = [json.loads(line) for line in content.split('\n') if line.strip()]
                except (FileNotFoundError, json.JSONDecodeError) as e:
                    print(f"Error loading time sessions: {e}")
                
                # Load agent data
                agents = {}
                try:
                    with open('../agents/active_agents.jsonl', 'r') as f:
                        content = f.read().strip()
                        if content:
                            agents = {line.split(' ', 1)[0]: json.loads(line.split(' ', 1)[1]) 
                                    for line in content.split('\n') if line.strip()}
                except (FileNotFoundError, json.JSONDecodeError) as e:
                    print(f"Error loading agents: {e}")
                
                # Load agent todos for observability
                agent_todos = []
                try:
                    with open('../agents/agent_todos.jsonl', 'r') as f:
                        content = f.read().strip()
                        if content:
                            agent_todos = [json.loads(line.split(' ', 1)[1]) 
                                         for line in content.split('\n') if line.strip()]
                except (FileNotFoundError, json.JSONDecodeError) as e:
                    print(f"Error loading agent todos: {e}")
                
                response = {
                    'jobs': jobs,
                    'progress': progress,
                    'time_sessions': time_sessions,
                    'agents': agents,
                    'agent_todos': agent_todos,
                    'timestamp': time.time(),
                    'status': 'success'
                }
                
                self.wfile.write(json.dumps(response, indent=2).encode())
            except Exception as e:
                error_response = {
                    'status': 'error',
                    'error': str(e),
                    'timestamp': time.time()
                }
                self.wfile.write(json.dumps(error_response).encode())
        elif parsed_path.path == '/api/health':
            # Health check endpoint
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            health_response = {
                'status': 'healthy',
                'timestamp': time.time(),
                'uptime': time.time() - start_time
            }
            self.wfile.write(json.dumps(health_response).encode())
        else:
            # Serve static files
            return super().do_GET()
    
    def do_POST(self):
        parsed_path = urlparse(self.path)
        
        # CORS headers for local development
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        
        if parsed_path.path == '/api/spawn-agent':
            # Handle agent spawning
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            try:
                content_length = int(self.headers.get('Content-Length', 0))
                post_data = self.rfile.read(content_length).decode('utf-8')
                data = json.loads(post_data)
                
                personality = data.get('personality')
                if not personality:
                    self.wfile.write(json.dumps({'status': 'error', 'error': 'No personality specified'}).encode())
                    return
                
                # Import agent manager
                sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'agents'))
                try:
                    from agent_manager import AgentManager, AgentPersonality
                    manager = AgentManager()
                    personality_enum = AgentPersonality(personality.lower())
                    agent_id = manager.spawn_agent(personality_enum)
                    self.wfile.write(json.dumps({
                        'status': 'success', 
                        'agent_id': agent_id,
                        'personality': personality
                    }).encode())
                except ValueError:
                    self.wfile.write(json.dumps({
                        'status': 'error', 
                        'error': f'Invalid personality: {personality}'
                    }).encode())
                    
            except Exception as e:
                self.wfile.write(json.dumps({
                    'status': 'error', 
                    'error': str(e)
                }).encode())
        else:
            # Default GET handler
            self.do_GET()

    def do_OPTIONS(self):
        # Handle preflight requests
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        self.end_headers()

    def log_message(self, format, *args):
        # Custom logging with timestamps
        timestamp = time.strftime('%Y-%m-%d %H:%M:%S')
        print(f"[{timestamp}] {format % args}")

# Global start time
start_time = time.time()

def start_server():
    """Start web UI server"""
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    
    PORT = 8080
    with socketserver.TCPServer(("", PORT), JobTrackerHandler) as httpd:
        print(f"🚀 Stratavore Job Tracker UI v3 running on http://localhost:{PORT}")
        print("📊 Enhanced with parallel agent system")
        print("🔧 API endpoints: /api/status, /api/health, /api/spawn-agent")
        print("⏹️  Press Ctrl+C to stop server")
        print(f"⏰ Server started at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n👋 Server stopped")

if __name__ == "__main__":
    start_server()