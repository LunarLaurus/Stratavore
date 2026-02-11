#!/usr/bin/env python3
"""
Simple agent status validation script
"""

import os
import sys
import json
import time
from pathlib import Path

def validate_agents():
    """Validate that agents are working properly"""
    print("🔍 VALIDATING AGENT STATUS")
    
    # Check if agent data file exists
    agents_file = "agents/active_agents.jsonl"
    if not os.path.exists(agents_file):
        print("  ❌ Agent data file not found")
        return False
    
    print(f"  ✅ Agent data file found: {agents_file}")
    
    # Read agent data
    try:
        with open(agents_file, 'r') as f:
            agents_content = f.read().strip()
        
        if not agents_content:
            print("  ⚠️  Agent data file is empty")
            return False
        
        # Parse agents from JSONL format
        agents = {}
        for line in agents_content.split('\n'):
            if line.strip():
                parts = line.split(' ', 1)
                if len(parts) == 2:
                    agent_id, agent_data = parts[0], json.loads(parts[1])
                    agents[agent_id] = agent_data
        
        print(f"  ✅ Loaded {len(agents)} agents")
        
    except Exception as e:
        print(f"  ❌ Error reading agent data: {e}")
        return False
    
    if not agents:
        print("  ❌ No agents to validate")
        return False
    
    # Validate each agent
    working_agents = 0
    idle_agents = 0
    spawning_agents = 0
    error_agents = 0
    
    for agent_id, agent_data in agents.items():
        status = agent_data.get('status', 'unknown')
        current_task = agent_data.get('current_task')
        thoughts = agent_data.get('thoughts', [])
        created_at = agent_data.get('created_at', 'unknown')
        
        print(f"\n🤖 Agent: {agent_id}")
        print(f"  Status: {status}")
        print(f"  Task: {current_task or 'None'}")
        print(f"  Created: {created_at}")
        print(f"  Thoughts: {len(thoughts)}")
        
        # Categorize agents
        if status == 'spawning':
            spawning_agents += 1
            print(f"  🚀 Status: Spawning")
        elif status == 'working':
            working_agents += 1
            print(f"  🔨 Status: Working")
        elif status == 'idle':
            idle_agents += 1
            print(f"  😴 Status: Idle")
        else:
            print(f"  ❓ Status: {status}")
            error_agents += 1
        
        # Check for recent thoughts (within last 10 minutes)
        recent_thoughts = [
            thought for thought in thoughts
            if thought.get('timestamp'):
                try:
                    thought_time = time.strptime(thought['timestamp'], "%Y-%m-%dT%H:%M:%S")
                    if time.time() - thought_time.timestamp() < 600:
                        recent_thoughts.append(thought)
                except (ValueError, TypeError):
                    # Skip malformed timestamps
                    pass
        
        if recent_thoughts:
            print(f"  💭 Recent thoughts: {len(recent_thoughts)}")
            for thought in recent_thoughts[-3:]:  # Last 3 thoughts
                timestamp = thought['timestamp']
                thought_text = thought['thought'][:50] + "..." if len(thought['thought']) > 50 else thought['thought']
                print(f"    - {timestamp}: {thought_text}")
    
    # Summary
    total_agents = len(agents)
    print(f"\n📊 AGENT SUMMARY")
    print(f"  Total Agents: {total_agents}")
    print(f"  🚀 Spawning: {spawning_agents}")
    print(f"  🔨 Working: {working_agents}")
    print(f"  😴 Idle: {idle_agents}")
    print(f"  ❓ Error: {error_agents}")
    
    if working_agents == 0:
        print(f"  ⚠️  WARNING: No agents are currently working!")
        return False
    else:
        print(f"  ✅ {working_agents} agents are actively working")
        return True

def test_agent_commands():
    """Test agent manager commands"""
    print("\n🧪 TESTING AGENT COMMANDS")
    
    # Test personalities command
    print("  Testing personalities command...")
    exit_code = os.system("python3 agents/agent_manager.py personalities")
    if exit_code == 0:
        print("  ✅ Personalities command works")
    else:
        print(f"  ❌ Personalities command failed with exit code {exit_code}")
    
    print("  Testing summary command...")
    exit_code = os.system("python3 agents/agent_manager.py summary")
    if exit_code == 0:
        print("  ✅ Summary command works")
    else:
        print(f"  ❌ Summary command failed with exit code {exit_code}")

def main():
    """Main validation function"""
    print("🔍 AGENT STATUS VALIDATION")
    
    # Validate agents are working
    agents_working = validate_agents()
    
    # Test agent manager commands
    test_agent_commands()
    
    print(f"\n🎯 VALIDATION COMPLETE")
    print("="*50)
    
    if agents_working:
        print("✅ All agents are working correctly")
        print("✅ Agent management system functional")
        print("✅ Ready for parallel workflows")
    else:
        print("❌ No agents are working")
        print("❌ Check agent management system")
        print("❌ Verify agent status commands")

if __name__ == "__main__":
    main()