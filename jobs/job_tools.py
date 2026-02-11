#!/usr/bin/env python3
"""
Job Management System Scripts
Validation, analysis, and automation for Stratavore job tracking
"""

import json
import sys
import os
from datetime import datetime
from typing import Dict, List, Set

def load_jobs() -> List[Dict]:
    """Load all jobs from JSONL file"""
    jobs = []
    try:
        with open('jobs/jobs.jsonl', 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    jobs.append(json.loads(line))
                except json.JSONDecodeError as exc:
                    print(f"[WARN] Skipping malformed job line: {exc}")
    except FileNotFoundError:
        print("❌ jobs.jsonl not found")
    return jobs

def validate_dependencies() -> bool:
    """Validate job dependencies and check for circular dependencies"""
    jobs = load_jobs()
    if not jobs:
        return True
    
    # Build dependency graph
    deps = {}
    for job in jobs:
        job_id = job['id']
        deps[job_id] = job.get('dependencies', [])
    
    # Check for circular dependencies
    def visit(job_id: str, visited: Set[str], path: List[str]) -> bool:
        if job_id in visited:
            print(f"❌ Circular dependency detected: {' → '.join(path)} → {job_id}")
            return False
        if job_id not in deps:
            return True
        
        visited.add(job_id)
        path.append(job_id)
        
        for dep in deps[job_id]:
            if not visit(dep, visited.copy(), path.copy()):
                return False
        
        return True
    
    print("🔍 Validating job dependencies...")
    valid = True
    for job_id in deps:
        if not visit(job_id, set(), []):
            valid = False
    
    if valid:
        print("✅ No circular dependencies found")
    
    # Check for missing dependencies
    print("\n🔍 Checking for missing dependencies...")
    all_job_ids = set(job['id'] for job in jobs)
    missing_deps = []
    for job_id, job_deps in deps.items():
        for dep in job_deps:
            if dep not in all_job_ids:
                missing_deps.append((job_id, dep))
    
    if missing_deps:
        print("❌ Missing dependencies found:")
        for job_id, missing_dep in missing_deps:
            print(f"  - {job_id} depends on {missing_dep} (not found)")
        valid = False
    else:
        print("✅ All dependencies found")
    
    return valid

def show_job_summary():
    """Show summary of current job status"""
    jobs = load_jobs()
    if not jobs:
        print("❌ No jobs found")
        return
    
    print(f"\n📊 JOB SUMMARY ({len(jobs)} total jobs)")
    print("=" * 50)
    
    # Status breakdown
    status_counts = {}
    for job in jobs:
        status = job['status']
        status_counts[status] = status_counts.get(status, 0) + 1
    
    print(f"🟡 Pending: {status_counts.get('pending', 0)}")
    print(f"🟠 In Progress: {status_counts.get('in_progress', 0)}")
    print(f"🟢 Completed: {status_counts.get('completed', 0)}")
    print(f"🔴 Cancelled: {status_counts.get('cancelled', 0)}")
    
    # Priority breakdown (excluding completed)
    print(f"\n🎯 PRIORITY BREAKDOWN (active jobs)")
    print("-" * 30)
    active_jobs = [j for j in jobs if j['status'] != 'completed']
    priority_counts = {}
    for job in active_jobs:
        priority = job['priority']
        priority_counts[priority] = priority_counts.get(priority, 0) + 1
    
    print(f"🔴 High: {priority_counts.get('high', 0)}")
    print(f"🟡 Medium: {priority_counts.get('medium', 0)}")
    print(f"🔵 Low: {priority_counts.get('low', 0)}")
    
    # Agent workload
    print(f"\n🤖 AGENT WORKLOAD")
    print("-" * 20)
    agent_stats = {}
    for job in jobs:
        agent = job.get('assignee', 'unassigned')
        if agent not in agent_stats:
            agent_stats[agent] = {'active': 0, 'completed': 0}
        
        if job['status'] == 'completed':
            agent_stats[agent]['completed'] += 1
        elif job['status'] == 'in_progress':
            agent_stats[agent]['active'] += 1
    
    for agent, stats in agent_stats.items():
        print(f"{agent}: {stats['active']} active, {stats['completed']} completed")

def suggest_next_jobs() -> List[Dict]:
    """Suggest jobs that can be started now (no pending dependencies)"""
    jobs = load_jobs()
    if not jobs:
        return []
    
    # Build dependency map
    completed_jobs = set(job['id'] for job in jobs if job['status'] == 'completed')
    in_progress_jobs = set(job['id'] for job in jobs if job['status'] == 'in_progress')
    pending_jobs = [job for job in jobs if job['status'] == 'pending']
    
    ready_jobs = []
    for job in pending_jobs:
        deps = job.get('dependencies', [])
        if all(dep in completed_jobs for dep in deps):
            ready_jobs.append(job)
    
    # Sort by priority, then by creation date
    priority_order = {'high': 0, 'medium': 1, 'low': 2}
    ready_jobs.sort(key=lambda j: (priority_order.get(j.get('priority', 'low'), 3), j.get('created_at') or ''))
    
    return ready_jobs

def generate_conflict_resolution() -> List[Dict]:
    """Generate conflict resolution suggestions for job scheduling"""
    jobs = load_jobs()
    if not jobs:
        return []
    
    conflicts = []
    
    # Check for agent conflicts (multiple jobs assigned to same agent)
    agent_assignments = {}
    for job in jobs:
        if job['status'] == 'pending' and job.get('assignee'):
            agent = job['assignee']
            if agent not in agent_assignments:
                agent_assignments[agent] = []
            agent_assignments[agent].append(job)
    
    for agent, assigned_jobs in agent_assignments.items():
        if len(assigned_jobs) > 1:
            # Sort by priority and suggest focusing on highest priority
            sorted_jobs = sorted(assigned_jobs, key=lambda j: j['priority'])
            conflicts.append({
                'type': 'agent_overload',
                'agent': agent,
                'conflicts': assigned_jobs,
                'suggestion': f"Focus on {sorted_jobs[0]['title']} ({sorted_jobs[0]['priority']} priority) first",
                'jobs_to_reassign': assigned_jobs[1:]  # Lower priority jobs
            })
    
    return conflicts

def show_ready_jobs():
    """Show jobs that are ready to be started"""
    ready_jobs = suggest_next_jobs()
    
    if not ready_jobs:
        print("\n⏳ No jobs are ready to start (all have unmet dependencies)")
        return
    
    print(f"\n🚀 READY TO START ({len(ready_jobs)} jobs)")
    print("=" * 50)
    
    for i, job in enumerate(ready_jobs, 1):
        print(f"{i}. {job['title']}")
        print(f"   🆔 ID: {job['id']}")
        print(f"   🎯 Priority: {job['priority']}")
        print(f"   📊 Estimate: {job['estimated_hours']}h")
        print(f"   📝 {job['description'][:100]}...")
        print()

def show_conflicts():
    """Show job conflicts and resolution suggestions"""
    conflicts = generate_conflict_resolution()
    
    if not conflicts:
        print("\n✅ No job conflicts detected")
        return
    
    print(f"\n⚠️  JOB CONFLICTS ({len(conflicts)} issues)")
    print("=" * 50)
    
    for conflict in conflicts:
        print(f"\n🔴 {conflict['type'].upper().replace('_', ' ')}")
        print(f"Agent: {conflict['agent']}")
        print(f"Suggestion: {conflict['suggestion']}")
        
        if 'jobs_to_reassign' in conflict:
            print("Jobs to reassign:")
            for job in conflict['jobs_to_reassign']:
                print(f"  - {job['title']} ({job['priority']} priority)")
        print()

def auto_assign_jobs():
    """Automatically assign unassigned jobs to agents based on workload"""
    jobs = load_jobs()
    if not jobs:
        return
    
    # Calculate current agent workload
    agent_workload = {}
    for job in jobs:
        if job.get('assignee') and job['status'] in ['in_progress', 'pending']:
            agent = job['assignee']
            agent_workload[agent] = agent_workload.get(agent, 0) + 1
    
    # Find agents with lowest workload
    min_workload = min(agent_workload.values()) if agent_workload else 0
    available_agents = [agent for agent, workload in agent_workload.items() if workload == min_workload]
    
    # Assign unassigned jobs
    unassigned_jobs = [job for job in jobs if job['status'] == 'pending' and not job.get('assignee')]
    
    if not unassigned_jobs:
        print("✅ All jobs are assigned")
        return
    
    print(f"\n🔄 AUTO-ASSIGNMENT SUGGESTIONS")
    print("=" * 40)
    
    for job in unassigned_jobs[:5]:  # Show top 5
        suggested_agent = available_agents[0] if available_agents else "unassigned"
        print(f"Job: {job['title']}")
        print(f"  Suggest agent: {suggested_agent}")
        print(f"  Priority: {job['priority']}")
        print()

def main():
    """Main CLI interface"""
    if len(sys.argv) < 2:
        print("Usage: python3 job_tools.py [command]")
        print("Commands:")
        print("  validate     - Validate job dependencies")
        print("  summary      - Show job summary")
        print("  ready        - Show jobs ready to start")
        print("  conflicts    - Show job conflicts and resolutions")
        print("  assign       - Suggest automatic job assignments")
        print("  all          - Run all commands")
        return
    
    command = sys.argv[1]
    
    if command == "validate":
        validate_dependencies()
    elif command == "summary":
        show_job_summary()
    elif command == "ready":
        show_ready_jobs()
    elif command == "conflicts":
        show_conflicts()
    elif command == "assign":
        auto_assign_jobs()
    elif command == "all":
        validate_dependencies()
        show_job_summary()
        show_ready_jobs()
        show_conflicts()
        auto_assign_jobs()
    else:
        print(f"❌ Unknown command: {command}")

if __name__ == "__main__":
    main()