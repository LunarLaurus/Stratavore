#!/usr/bin/env python3
"""
Simple agent manager test without emoji issues
"""

import sys
import os
from pathlib import Path

def test_agent_system():
    """Test basic agent system functionality"""
    repo_root = Path(__file__).parent.parent.parent
    agents_file = repo_root / "agents" / "active_agents.jsonl"
    
    print("Stratavore Agent System Test")
    print("=" * 40)
    
    # Check if agents file exists
    if not agents_file.exists():
        print(f"ERROR: Agents file not found: {agents_file}")
        return False
    
    print(f"Agents file: {agents_file}")
    
    # Count agents
    try:
        with open(agents_file, 'r', encoding='utf-8') as f:
            content = f.read().strip()
            if not content:
                print("No agents found")
                return True
                
            lines = [line for line in content.split('\n') if line.strip()]
            print(f"Total agents: {len(lines)}")
            
            # Simple status breakdown
            status_counts = {}
            for line in lines:
                parts = line.split(' ', 1)
                if len(parts) == 2:
                    try:
                        import json
                        agent_data = json.loads(parts[1])
                        status = agent_data.get('status', 'unknown')
                        status_counts[status] = status_counts.get(status, 0) + 1
                    except:
                        pass
            
            for status, count in status_counts.items():
                print(f"  {status}: {count}")
            
            return True
            
    except Exception as e:
        print(f"Error reading agents: {e}")
        return False

if __name__ == "__main__":
    success = test_agent_system()
    sys.exit(0 if success else 1)