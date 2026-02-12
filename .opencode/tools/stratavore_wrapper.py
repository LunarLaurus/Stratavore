#!/usr/bin/env python3
"""
Cross-platform wrapper for Stratavore agent manager with proper encoding handling
"""

import sys
import os
import subprocess
import json
from pathlib import Path

def main():
    """Main wrapper function"""
    if len(sys.argv) < 2:
        print("Usage: python stratavore_wrapper.py [command] [args...]")
        return 1
    
    # Find agent manager - go up from .opencode/tools to repo root
    script_dir = Path(__file__).parent.parent.parent  # From .opencode/tools to repo root
    agents_dir = script_dir / "agents"
    agent_manager = agents_dir / "agent_manager.py"
    job_tools = script_dir / "jobs" / "job_tools.py"
    
    if not agent_manager.exists():
        print(f"ERROR: Agent manager not found: {agent_manager}")
        return 1
    
    # Build command
    command = sys.argv[1]
    args = sys.argv[2:]
    
    if command in ["spawn", "assign", "complete", "status", "list", "summary", "personalities", "available", "unstuck-spawning"]:
        cmd = [sys.executable, str(agent_manager), command] + args
    elif command in ["validate", "summary", "ready", "conflicts", "assign", "all"]:
        if not job_tools.exists():
            print(f"ERROR: Job tools not found: {job_tools}")
            return 1
        cmd = [sys.executable, str(job_tools), command] + args
    else:
        print(f"ERROR: Unknown command: {command}")
        return 1
    
    try:
        # Execute with proper encoding
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            encoding='utf-8',
            timeout=30,
            env={**os.environ, 'PYTHONIOENCODING': 'utf-8'}
        )
        
        if result.stdout:
            print(result.stdout, end='')
        
        if result.stderr:
            print(result.stderr, file=sys.stderr)
        
        return result.returncode
        
    except subprocess.TimeoutExpired:
        print("ERROR: Command timed out after 30 seconds")
        return 1
    except Exception as e:
        print(f"ERROR: Error executing command: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())