import { tool } from "@opencode-ai/plugin"
import path from "path"
import fs from "fs"

export default tool({
  description: "Stratavore agent/task system - Complete management interface for spawning, assigning, monitoring agents and managing jobs",
  args: {
    action: tool.schema.string().describe("Action: 'spawn', 'assign', 'complete', 'status', 'list', 'summary', 'job-status', 'spawn-batch', 'personalities', 'available'"),
    personality: tool.schema.string().optional().describe("Agent personality: cadet, senior, specialist, researcher, debugger, optimizer"),
    agent_id: tool.schema.string().optional().describe("Target agent ID for operations"),
    task_id: tool.schema.string().optional().describe("Task/job ID to assign or complete"),
    success: tool.schema.boolean().optional().describe("Task completion status (true/false, default: true)"),
    notes: tool.schema.string().optional().describe("Notes for task completion or status updates"),
    thought: tool.schema.string().optional().describe("Agent thought log entry"),
    status: tool.schema.string().optional().describe("Agent status: idle, working, paused, completed, error"),
    count: tool.schema.number().optional().describe("Number of agents to spawn (for batch operations)"),
    job_filter: tool.schema.string().optional().describe("Filter jobs by status: pending, in_progress, completed, cancelled")
  },
  async execute(args, context) {
    const { worktree } = context
    
    try {
      switch (args.action) {
        case "spawn":
          return await spawnAgent(worktree, args.personality, args.task_id)
        case "spawn-batch":
          return await spawnBatchAgents(worktree, args.personality, args.count, args.task_id)
        case "assign":
          return await assignTask(worktree, args.agent_id, args.task_id)
        case "complete":
          return await completeTask(worktree, args.agent_id, args.success, args.notes)
        case "status":
          return await updateAgentStatus(worktree, args.agent_id, args.status, args.thought)
        case "list":
          return await listAgents(worktree, args.personality)
        case "summary":
          return await getAgentSummary(worktree)
        case "job-status":
          return await getJobStatus(worktree, args.job_filter)
        case "personalities":
          return await getPersonalities(worktree)
        case "available":
          return await getAvailableAgents(worktree, args.personality)
        default:
          return `ERROR: Unknown action: ${args.action}. Valid actions: spawn, assign, complete, status, list, summary, job-status, spawn-batch, personalities, available`
      }
    } catch (error) {
      return `❌ Error executing ${args.action}: ${error.message}`
    }
  }
})

async function spawnAgent(worktree, personality, taskId) {
  if (!personality) {
    return "❌ Personality required for spawning. Options: cadet, senior, specialist, researcher, debugger, optimizer"
  }
  
  const validPersonalities = ["cadet", "senior", "specialist", "researcher", "debugger", "optimizer"]
  if (!validPersonalities.includes(personality.toLowerCase())) {
    return `❌ Invalid personality. Valid options: ${validPersonalities.join(", ")}`
  }
  
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" spawn ${personality.toLowerCase()} ${taskId || ""}`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function spawnBatchAgents(worktree, personality, count, taskId) {
  if (!personality || !count) {
    return "❌ Both personality and count required for batch spawning"
  }
  
  const results = []
  for (let i = 0; i < count; i++) {
    const result = await spawnAgent(worktree, personality, taskId)
    results.push(result)
  }
  
  return `🚀 Spawned ${count} ${personality} agents:\n${results.join("\n")}`
}

async function assignTask(worktree, agentId, taskId) {
  if (!agentId || !taskId) {
    return "❌ Both agent_id and task_id required for task assignment"
  }
  
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" assign ${agentId} ${taskId}`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function completeTask(worktree, agentId, success = true, notes = "") {
  if (!agentId) {
    return "❌ agent_id required for task completion"
  }
  
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" complete ${agentId} ${success} "${notes}"`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function updateAgentStatus(worktree, agentId, status, thought) {
  if (!agentId || !status) {
    return "❌ Both agent_id and status required for status update"
  }
  
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" status ${agentId} ${status} "${thought || ""}"`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function listAgents(worktree, personality) {
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" list ${personality || ""}`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function getAgentSummary(worktree) {
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" summary`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function getJobStatus(worktree, filter) {
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" summary`
  const result = await executeCommand(command)
  
  if (filter) {
    // Additional filtering logic could be added here
    return `🔍 Jobs filtered by status: ${filter}\n${result.stdout || result.stderr}`
  }
  
  return result.stdout || result.stderr
}

async function getPersonalities(worktree) {
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" personalities`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function getAvailableAgents(worktree, personality) {
  const wrapperPath = path.join(worktree, ".opencode/tools/stratavore_wrapper.py")
  const command = `python "${wrapperPath}" available ${personality || ""}`
  const result = await executeCommand(command)
  return result.stdout || result.stderr
}

async function executeCommand(command) {
  const { execSync } = require("child_process")
  try {
    const stdout = execSync(command, { 
      encoding: "utf8", 
      timeout: 30000,
      env: { ...process.env, PYTHONIOENCODING: "utf-8" }
    })
    return { stdout: stdout.trim(), stderr: null }
  } catch (error) {
    return { stdout: null, stderr: error.stderr || error.message }
  }
}