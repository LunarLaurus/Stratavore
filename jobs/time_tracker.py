#!/usr/bin/env python3
"""
Granular Time Tracking System
Track job work sessions with second precision
"""

import json
import time
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Optional

class TimeTracker:
    def __init__(self):
        self.sessions_file = "jobs/time_sessions.jsonl"
        self.ensure_file_exists()
    
    def ensure_file_exists(self):
        """Create sessions file if it doesn't exist"""
        try:
            with open(self.sessions_file, 'r') as f:
                pass
        except FileNotFoundError:
            with open(self.sessions_file, 'w') as f:
                f.write("")
    
    def start_session(self, job_id: str, agent: str, description: str = "") -> str:
        """Start a new work session"""
        session_id = f"{job_id}_{int(time.time())}"
        session = {
            "session_id": session_id,
            "job_id": job_id,
            "agent": agent,
            "description": description,
            "status": "active",
            "start_time": datetime.now().isoformat(),
            "start_timestamp": time.time(),
            "end_time": None,
            "end_timestamp": None,
            "duration_seconds": None,
            "paused_time": 0,
            "pauses": [],
            "created_at": datetime.now().isoformat()
        }
        
        self.append_session(session)
        print(f"⏱️  Started session {session_id} for job {job_id}")
        return session_id
    
    def end_session(self, session_id: str, notes: str = "") -> bool:
        """End an active work session"""
        sessions = self.load_sessions()
        
        for i, session in enumerate(sessions):
            if session["session_id"] == session_id and session["status"] == "active":
                session["status"] = "completed"
                session["end_time"] = datetime.now().isoformat()
                session["end_timestamp"] = time.time()
                session["duration_seconds"] = session["end_timestamp"] - session["start_timestamp"] - session["paused_time"]
                session["notes"] = notes
                
                sessions[i] = session
                self.save_sessions(sessions)
                
                duration = timedelta(seconds=int(session["duration_seconds"]))
                print(f"⏹️  Ended session {session_id}")
                print(f"   Duration: {duration}")
                return True
        
        print(f"❌ Active session {session_id} not found")
        return False
    
    def append_session(self, session: Dict):
        """Append a new session to the file"""
        with open(self.sessions_file, 'a') as f:
            f.write(json.dumps(session) + '\n')
    
    def load_sessions(self) -> List[Dict]:
        """Load all sessions from file"""
        try:
            with open(self.sessions_file, 'r') as f:
                content = f.read().strip()
                if not content:
                    return []
                return [json.loads(line) for line in content.split('\n') if line]
        except FileNotFoundError:
            return []
    
    def save_sessions(self, sessions: List[Dict]):
        """Save all sessions back to file"""
        with open(self.sessions_file, 'w') as f:
            for session in sessions:
                f.write(json.dumps(session) + '\n')
    
    def get_active_sessions(self) -> List[Dict]:
        """Get all currently active sessions"""
        sessions = self.load_sessions()
        return [s for s in sessions if s["status"] == "active"]
    
    def calculate_job_time(self, job_id: str) -> Dict:
        """Calculate total time spent on a job"""
        sessions = self.get_job_sessions(job_id)
        
        total_seconds = 0
        completed_sessions = 0
        
        for session in sessions:
            if session["status"] == "completed" and session["duration_seconds"]:
                total_seconds += session["duration_seconds"]
                completed_sessions += 1
            elif session["status"] == "active":
                # Calculate current session time
                current_time = time.time() - session["start_timestamp"] - session["paused_time"]
                total_seconds += current_time
        
        return {
            "job_id": job_id,
            "total_seconds": total_seconds,
            "total_hours": total_seconds / 3600,
            "completed_sessions": completed_sessions,
            "active_sessions": len([s for s in sessions if s["status"] == "active"]),
            "formatted_time": str(timedelta(seconds=int(total_seconds)))
        }
    
    def get_job_sessions(self, job_id: str) -> List[Dict]:
        """Get all sessions for a specific job"""
        sessions = self.load_sessions()
        return [s for s in sessions if s["job_id"] == job_id]

def main():
    """CLI interface for time tracking"""
    tracker = TimeTracker()
    
    if len(sys.argv) < 2:
        print("Usage: python3 time_tracker.py [command]")
        print("Commands:")
        print("  start <job_id> <agent> [description]  - Start work session")
        print("  end <session_id> [notes]            - End work session")
        print("  active                               - Show active sessions")
        print("  job <job_id>                        - Show job time summary")
        print("  all                                  - Show all session stats")
        return
    
    command = sys.argv[1]
    
    if command == "start":
        if len(sys.argv) < 4:
            print("Usage: python3 time_tracker.py start <job_id> <agent> [description]")
            return
        job_id = sys.argv[2]
        agent = sys.argv[3]
        description = sys.argv[4] if len(sys.argv) > 4 else ""
        tracker.start_session(job_id, agent, description)
    
    elif command == "end":
        if len(sys.argv) < 3:
            print("Usage: python3 time_tracker.py end <session_id> [notes]")
            return
        session_id = sys.argv[2]
        notes = sys.argv[3] if len(sys.argv) > 3 else ""
        tracker.end_session(session_id, notes)
    
    elif command == "active":
        active = tracker.get_active_sessions()
        if not active:
            print("📭 No active sessions")
        else:
            print(f"🔄 {len(active)} active sessions:")
            for session in active:
                start_time = datetime.fromisoformat(session["start_time"])
                duration = time.time() - session["start_timestamp"]
                print(f"  {session['session_id']}")
                print(f"    Job: {session['job_id']}")
                print(f"    Agent: {session['agent']}")
                print(f"    Started: {start_time.strftime('%H:%M:%S')}")
                print(f"    Duration: {str(timedelta(seconds=int(duration)))}")
                if session.get("description"):
                    print(f"    Note: {session['description']}")
                print()
    
    elif command == "job":
        if len(sys.argv) < 3:
            print("Usage: python3 time_tracker.py job <job_id>")
            return
        job_id = sys.argv[2]
        time_info = tracker.calculate_job_time(job_id)
        print(f"📊 Time Summary for {job_id}")
        print(f"  Total Time: {time_info['formatted_time']}")
        print(f"  Hours: {time_info['total_hours']:.2f}")
        print(f"  Completed Sessions: {time_info['completed_sessions']}")
        print(f"  Active Sessions: {time_info['active_sessions']}")
    
    elif command == "all":
        sessions = tracker.load_sessions()
        print(f"📈 ALL SESSIONS ({len(sessions)} total)")
        
        # Group by job
        job_times = {}
        for session in sessions:
            job_id = session["job_id"]
            if job_id not in job_times:
                job_times[job_id] = tracker.calculate_job_time(job_id)
        
        for job_id, time_info in job_times.items():
            print(f"\n🏷️  {job_id}")
            print(f"   ⏱️  {time_info['formatted_time']} ({time_info['total_hours']:.2f}h)")
            print(f"   📋 {time_info['completed_sessions']} completed, {time_info['active_sessions']} active")

if __name__ == "__main__":
    main()