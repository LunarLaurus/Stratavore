/**
 * Agent Status Component - Agent monitoring and management
 * Displays active agents, their status, and provides management controls
 */

class AgentStatus {
    constructor(container) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.agents = {};
        this.selectedAgent = null;
        this.showControls = false;
        
        this.init();
        this.bindEvents();
    }
    
    init() {
        this.render();
        this.registerComponent();
    }
    
    render() {
        this.container.innerHTML = `
            <div class="card">
                <div class="card-header">
                    <h3>🤖 AGENT STATUS &amp; WORKFLOW</h3>
                    <div class="card-actions">
                        <button class="btn btn--sm btn--secondary" data-action="spawn-agent">🚀 SPAWN AGENT</button>
                        <button class="btn btn--sm btn--outline" data-action="toggle-controls">🎛 AGENT CONTROLS</button>
                    </div>
                </div>
                
                <!-- Agent controls panel (hidden by default) -->
                <div id="agent-controls" class="control-panel" style="display: none;">
                    <h4>🎛 Agent Control Panel</h4>
                    <div class="control-grid">
                        ${this.renderControlForms()}
                    </div>
                    <div id="control-result" class="control-result" style="display: none;"></div>
                </div>
                
                <div class="agents-status-panel scrollable scrollable--lg" id="agents-container">
                    <div class="loading">🔄 Loading agents...</div>
                </div>
            </div>
        `;
    }
    
    renderControlForms() {
        return `
            <!-- Assign task -->
            <div class="control-form">
                <h6>📋 Assign Task</h6>
                <input id="ctrl-assign-agent" type="text" placeholder="Agent ID" class="form-input--sm">
                <input id="ctrl-assign-task" type="text" placeholder="Task ID (e.g. JOB-001)" class="form-input--sm">
                <button class="btn btn--sm" data-action="assign-task">Assign</button>
            </div>
            
            <!-- Complete task -->
            <div class="control-form">
                <h6>✅ Complete Task</h6>
                <input id="ctrl-complete-agent" type="text" placeholder="Agent ID" class="form-input--sm">
                <input id="ctrl-complete-notes" type="text" placeholder="Notes (optional)" class="form-input--sm">
                <select id="ctrl-complete-success" class="form-input--sm">
                    <option value="true">✅ Success</option>
                    <option value="false">❌ Failed</option>
                </select>
                <button class="btn btn--sm" data-action="complete-task">Complete</button>
            </div>
            
            <!-- Update status -->
            <div class="control-form">
                <h6>📡 Update Status</h6>
                <input id="ctrl-status-agent" type="text" placeholder="Agent ID" class="form-input--sm">
                <select id="ctrl-status-value" class="form-input--sm">
                    <option value="idle">😴 Idle</option>
                    <option value="working">🔨 Working</option>
                    <option value="paused">⏸️ Paused</option>
                    <option value="error">❌ Error</option>
                </select>
                <input id="ctrl-status-thought" type="text" placeholder="Thought (optional)" class="form-input--sm">
                <button class="btn btn--sm" data-action="update-status">Update</button>
            </div>
            
            <!-- Kill agent -->
            <div class="control-form">
                <h6>🔴 Kill Agent</h6>
                <input id="ctrl-kill-agent" type="text" placeholder="Agent ID" class="form-input--sm">
                <p style="color: var(--color-text-muted); font-size: var(--font-size-xs);">Marks agent as error state</p>
                <button class="btn btn--sm btn--danger" data-action="kill-agent">Kill</button>
            </div>
        `;
    }
    
    bindEvents() {
        // Subscribe to agents data
        window.dataStore?.subscribe('agents', (agents) => {
            this.agents = agents;
            this.renderAgents();
        }, { immediate: true });
        
        // Subscribe to selected agent
        window.dataStore?.subscribe('selectedAgent', (agentId) => {
            this.selectedAgent = agentId;
            this.updateAgentSelection();
        });
        
        // Container click events
        this.container.addEventListener('click', (e) => {
            if (e.target.dataset.action) {
                this.handleAction(e.target.dataset.action);
            }
            
            // Agent item click
            const agentItem = e.target.closest('.agent-status-item');
            if (agentItem && !e.target.closest('button')) {
                const agentId = agentItem.dataset.agentId;
                this.selectAgent(agentId);
            }
        });
        
        // Listen for agent updates
        window.eventBus?.on('agent:spawned', (event) => {
            this.showNotification(`Agent ${event.data.agentId} spawned successfully`, 'success');
        });
        
        window.eventBus?.on('agent:error', (event) => {
            this.showNotification(`Agent error: ${event.data.message}`, 'error');
        });
    }
    
    renderAgents() {
        const agentsContainer = this.container.querySelector('#agents-container');
        if (!agentsContainer) return;
        
        const agentEntries = Object.entries(this.agents);
        
        if (agentEntries.length === 0) {
            agentsContainer.innerHTML = `
                <div class="empty-state">
                    No agents running. Click 🚀 SPAWN AGENT to start one.
                </div>
            `;
            return;
        }
        
        agentsContainer.innerHTML = agentEntries.map(([agentId, agent]) => 
            this.renderAgent(agentId, agent)
        ).join('');
        
        // Animate agent items
        this.animateAgentItems();
    }
    
    renderAgent(agentId, agent) {
        const statusEmoji = this.getAgentStatusEmoji(agent.status);
        const personalityEmoji = this.getAgentPersonalityEmoji(agent.personality);
        const lastThought = (agent.thoughts && agent.thoughts.length)
            ? agent.thoughts[agent.thoughts.length - 1].thought 
            : 'No recent thoughts';
        const tasksCompleted = (agent.metrics && agent.metrics.tasks_completed) || 0;
        const isSelected = this.selectedAgent === agentId;
        
        return `
            <div class="agent-status-item ${agent.status || ''} ${isSelected ? 'agent-status-item--selected' : ''}" 
                 data-agent-id="${agentId}">
                <div class="agent-header">
                    <span class="agent-emoji">${statusEmoji}</span>
                    <span class="agent-emoji">${personalityEmoji}</span>
                    <span class="agent-id">${agentId}</span>
                    <span class="agent-personality">(${agent.personality || 'unknown'})</span>
                    <button class="agent-use-btn" data-action="copy-to-controls" data-agent-id="${agentId}">
                        use ↗
                    </button>
                </div>
                <div class="agent-details">
                    <div class="agent-detail-row">
                        <span class="agent-detail-label">Task:</span>
                        <span class="agent-detail-value">${agent.current_task || 'No task assigned'}</span>
                    </div>
                    <div class="agent-detail-row">
                        <span class="agent-detail-label">Thought:</span>
                        <span class="agent-detail-value">${this.truncateText(lastThought, 50)}</span>
                    </div>
                    <div class="agent-detail-row">
                        <span class="agent-detail-label">Completed:</span>
                        <span class="agent-detail-value">${tasksCompleted} tasks</span>
                    </div>
                    <div class="agent-detail-row">
                        <span class="agent-detail-label">Since:</span>
                        <span class="agent-detail-value">${agent.created_at ? this.formatDate(agent.created_at) : 'N/A'}</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    handleAction(action) {
        switch (action) {
            case 'spawn-agent':
                this.spawnAgent();
                break;
            case 'toggle-controls':
                this.toggleControls();
                break;
            case 'assign-task':
                this.assignTask();
                break;
            case 'complete-task':
                this.completeTask();
                break;
            case 'update-status':
                this.updateStatus();
                break;
            case 'kill-agent':
                this.killAgent();
                break;
            case 'copy-to-controls':
                const agentId = event.target.dataset.agentId;
                this.copyToControls(agentId);
                break;
        }
    }
    
    async spawnAgent() {
        const personalities = ['cadet', 'senior', 'specialist', 'researcher', 'debugger', 'optimizer'];
        const personality = prompt(
            'Enter agent personality:\n' + personalities.map((p, i) => `  ${i + 1}. ${p}`).join('\n')
        );
        
        if (!personality) return;
        
        const normalized = personality.trim().toLowerCase()
            .replace(/^\d+\.\s*/, '')  // allow "1. cadet" input
            .replace(/\s+/g, '');
        
        try {
            const response = await window.apiClient?.spawnAgent(normalized);
            if (response.ok) {
                this.showNotification(`✅ Agent spawned!\nID: ${response.data.agent_id}\nPersonality: ${response.data.personality}`, 'success');
                window.eventBus?.emit('agent:spawned', response.data);
                setTimeout(() => {
                    window.eventBus?.emit('header:refresh-requested');
                }, 1200);
            } else {
                this.showNotification(`❌ Failed: ${response.data.error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`❌ Network error: ${error.message}`, 'error');
        }
    }
    
    toggleControls() {
        const controlsPanel = this.container.querySelector('#agent-controls');
        this.showControls = !this.showControls;
        controlsPanel.style.display = this.showControls ? 'block' : 'none';
    }
    
    async assignTask() {
        const agentId = this.container.querySelector('#ctrl-assign-agent').value.trim();
        const taskId = this.container.querySelector('#ctrl-assign-task').value.trim();
        
        if (!agentId || !taskId) {
            this.showControlResult('⚠️ Agent ID and Task ID are required', false);
            return;
        }
        
        try {
            const response = await window.apiClient?.assignAgent(agentId, taskId);
            if (response.ok) {
                this.showControlResult(`✅ Assigned ${taskId} → ${agentId}`, true);
                setTimeout(() => {
                    window.eventBus?.emit('header:refresh-requested');
                }, 800);
            } else {
                this.showControlResult(`❌ ${response.data.error}`, false);
            }
        } catch (error) {
            this.showControlResult(`❌ Network error: ${error.message}`, false);
        }
    }
    
    async completeTask() {
        const agentId = this.container.querySelector('#ctrl-complete-agent').value.trim();
        const notes = this.container.querySelector('#ctrl-complete-notes').value.trim();
        const success = this.container.querySelector('#ctrl-complete-success').value === 'true';
        
        if (!agentId) {
            this.showControlResult('⚠️ Agent ID is required', false);
            return;
        }
        
        try {
            const response = await window.apiClient?.completeTask(agentId, success, notes);
            if (response.ok) {
                this.showControlResult(`✅ Task marked ${success ? 'complete' : 'failed'} for ${agentId}`, true);
                setTimeout(() => {
                    window.eventBus?.emit('header:refresh-requested');
                }, 800);
            } else {
                this.showControlResult(`❌ ${response.data.error}`, false);
            }
        } catch (error) {
            this.showControlResult(`❌ Network error: ${error.message}`, false);
        }
    }
    
    async updateStatus() {
        const agentId = this.container.querySelector('#ctrl-status-agent').value.trim();
        const status = this.container.querySelector('#ctrl-status-value').value;
        const thought = this.container.querySelector('#ctrl-status-thought').value.trim() || null;
        
        if (!agentId) {
            this.showControlResult('⚠️ Agent ID is required', false);
            return;
        }
        
        try {
            const response = await window.apiClient?.updateAgentStatus(agentId, status, thought);
            if (response.ok) {
                this.showControlResult(`✅ ${agentId} → ${status}`, true);
                setTimeout(() => {
                    window.eventBus?.emit('header:refresh-requested');
                }, 800);
            } else {
                this.showControlResult(`❌ ${response.data.error}`, false);
            }
        } catch (error) {
            this.showControlResult(`❌ Network error: ${error.message}`, false);
        }
    }
    
    async killAgent() {
        const agentId = this.container.querySelector('#ctrl-kill-agent').value.trim();
        
        if (!agentId) {
            this.showControlResult('⚠️ Agent ID is required', false);
            return;
        }
        
        if (!confirm(`Kill agent ${agentId}?`)) return;
        
        try {
            const response = await window.apiClient?.killAgent(agentId);
            if (response.ok) {
                this.showControlResult(`✅ Agent ${agentId} killed`, true);
                setTimeout(() => {
                    window.eventBus?.emit('header:refresh-requested');
                }, 800);
            } else {
                this.showControlResult(`❌ ${response.data.error}`, false);
            }
        } catch (error) {
            this.showControlResult(`❌ Network error: ${error.message}`, false);
        }
    }
    
    copyToControls(agentId) {
        const fields = ['ctrl-assign-agent', 'ctrl-complete-agent', 'ctrl-status-agent', 'ctrl-kill-agent'];
        fields.forEach(fieldId => {
            const field = this.container.querySelector(`#${fieldId}`);
            if (field) field.value = agentId;
        });
        
        // Show controls panel
        const controlsPanel = this.container.querySelector('#agent-controls');
        if (controlsPanel.style.display === 'none') {
            this.toggleControls();
        }
    }
    
    selectAgent(agentId) {
        this.selectedAgent = agentId;
        window.dataStore?.updateState('selectedAgent', agentId);
        
        const agent = this.agents[agentId];
        window.eventBus?.emit('agent:selected', { agentId, agent });
    }
    
    updateAgentSelection() {
        const agentItems = this.container.querySelectorAll('.agent-status-item');
        agentItems.forEach(item => {
            const isSelected = item.dataset.agentId === this.selectedAgent;
            item.classList.toggle('agent-status-item--selected', isSelected);
        });
    }
    
    showControlResult(message, ok) {
        const resultElement = this.container.querySelector('#control-result');
        if (!resultElement) return;
        
        resultElement.style.display = 'block';
        resultElement.style.background = ok ? 'var(--color-success)' : 'var(--color-error)';
        resultElement.style.color = 'white';
        resultElement.style.border = `1px solid ${ok ? 'var(--color-success)' : 'var(--color-error)'}`;
        resultElement.textContent = message;
        
        setTimeout(() => {
            resultElement.style.display = 'none';
        }, 4000);
    }
    
    showNotification(message, type = 'info') {
        window.eventBus?.emit('notification', { message, type });
    }
    
    getAgentStatusEmoji(status) {
        return {
            idle: '😴',
            working: '🔨',
            spawning: '🚀',
            completed: '✅',
            error: '❌',
            paused: '⏸️'
        }[status] || '❓';
    }
    
    getAgentPersonalityEmoji(personality) {
        return {
            cadet: '👨‍🚀',
            senior: '👴',
            specialist: '🎯',
            researcher: '🔬',
            debugger: '🐛',
            optimizer: '⚡'
        }[personality] || '🤖';
    }
    
    truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }
    
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    }
    
    animateAgentItems() {
        const agentItems = this.container.querySelectorAll('.agent-status-item');
        agentItems.forEach((item, index) => {
            item.style.opacity = '0';
            item.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                item.style.transition = 'all 0.3s ease';
                item.style.opacity = '1';
                item.style.transform = 'translateY(0)';
            }, index * 100);
        });
    }
    
    registerComponent() {
        window.eventBus?.registerComponent('agent-status', this);
    }
    
    // Public methods
    refresh() {
        this.renderAgents();
    }
    
    getSelectedAgent() {
        return this.selectedAgent ? this.agents[this.selectedAgent] : null;
    }
    
    getAgentCount() {
        return Object.keys(this.agents).length;
    }
    
    getWorkingAgents() {
        return Object.values(this.agents).filter(agent => agent.status === 'working');
    }
    
    destroy() {
        window.eventBus?.unregisterComponent('agent-status');
        this.container.innerHTML = '';
    }
}

// Auto-initialize if container exists
document.addEventListener('DOMContentLoaded', () => {
    const agentsContainer = document.querySelector('#agents-status-panel');
    if (agentsContainer) {
        window.agentStatus = new AgentStatus(agentsContainer.parentElement);
    }
});

export default AgentStatus;