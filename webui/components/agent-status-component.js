/**
 * Agent Status Component - Displays and manages agent status and controls
 */
class AgentStatusComponent extends BaseComponent {
    constructor(elementId, options = {}) {
        super(elementId, options);
        this.selectedAgents = new Set();
        this.showControls = false;
        this.virtualScrollEnabled = false;
        this.visibleRange = { start: 0, end: 50 };
    }

    setupEventListeners() {
        // Agent control buttons
        this.addEventListener(this.element, 'click', (e) => {
            if (e.target.matches('[data-action="spawn-agent"]')) {
                this.handleSpawnAgent();
            } else if (e.target.matches('[data-action="toggle-controls"]')) {
                this.toggleControls();
            } else if (e.target.matches('[data-action="select-agent"]')) {
                this.toggleAgentSelection(e.target.dataset.agentId);
            } else if (e.target.matches('[data-action="copy-agent-id"]')) {
                this.copyAgentId(e.target.dataset.agentId);
            } else if (e.target.matches('[data-action="batch-assign"]')) {
                this.handleBatchAssign();
            } else if (e.target.matches('[data-action="batch-status"]')) {
                this.handleBatchStatusUpdate();
            } else if (e.target.matches('[data-action="batch-kill"]')) {
                this.handleBatchKill();
            }
        });

        // Control form submissions
        const controlForms = this.findAll('.control-form');
        controlForms.forEach(form => {
            this.addEventListener(form, 'submit', (e) => {
                e.preventDefault();
                this.handleControlAction(e.target.dataset.action);
            });
        });

        // Virtual scrolling
        const agentsContainer = this.find('#agents-container');
        if (agentsContainer) {
            this.addEventListener(agentsContainer, 'scroll', 
                this.debounce(() => this.handleVirtualScroll(), 16)
            );
        }
    }

    setupSubscriptions() {
        // Subscribe to agent data changes
        this.subscribe('agents', () => this.renderAgents());
        this.subscribe('agentTodos', () => this.renderAgents());

        // Listen for agent events
        window.StratavoreCore.eventBus.on('agent:spawned', (data) => {
            this.showNotification(`Agent ${data.agentId} spawned!`, 'success');
        });
        window.StratavoreCore.eventBus.on('agent:updated', () => this.renderAgents());
        window.StratavoreCore.eventBus.on('agent:killed', (data) => {
            this.showNotification(`Agent ${data.agentId} terminated`, 'warning');
        });
    }

    handleSpawnAgent() {
        const personalities = ['cadet', 'senior', 'specialist', 'researcher', 'debugger', 'optimizer'];
        const personality = prompt(
            'Enter agent personality:\\n' + personalities.map((p, i) => `  ${i + 1}. ${p}`).join('\\n')
        );
        
        if (!personality) return;
        
        const norm = personality.trim().toLowerCase()
            .replace(/^\\d+\\.\\s*/, '')  // allow "1. cadet" input
            .replace(/\\s+/g, '');
        
        if (!this.isValidPersonality(norm)) {
            this.showNotification('Invalid personality specified', 'error');
            return;
        }

        this.spawnAgent(norm);
    }

    async spawnAgent(personality) {
        try {
            const data = await window.StratavoreCore.apiClient.spawnAgent(personality);
            if (data.status === 'success') {
                this.showNotification(`✅ Agent spawned!\\nID: ${data.agent_id}\\nPersonality: ${data.personality}`, 'success');
                this.emit('agent:spawned', { agentId: data.agent_id, personality: data.personality });
                setTimeout(() => this.emit('refresh:requested'), 1200);
            } else {
                this.showNotification(`❌ Failed: ${data.error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`❌ Network error: ${error.message}`, 'error');
        }
    }

    toggleControls() {
        this.showControls = !this.showControls;
        const controlsPanel = this.find('#agent-controls');
        const toggleBtn = this.find('[data-action="toggle-controls"]');
        
        if (controlsPanel) {
            controlsPanel.style.display = this.showControls ? 'block' : 'none';
        }
        
        if (toggleBtn) {
            toggleBtn.textContent = this.showControls ? '🎛 HIDE CONTROLS' : '🎛 AGENT CONTROLS';
        }
    }

    toggleAgentSelection(agentId) {
        if (this.selectedAgents.has(agentId)) {
            this.selectedAgents.delete(agentId);
        } else {
            this.selectedAgents.add(agentId);
        }
        this.updateSelectionUI();
    }

    copyAgentId(agentId) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(agentId);
            this.showNotification(`Agent ID ${agentId} copied to clipboard`, 'info');
        }
        this.populateControlForms(agentId);
    }

    populateControlForms(agentId) {
        const inputs = this.findAll('[data-agent-input]');
        inputs.forEach(input => {
            if (input.type === 'text') {
                input.value = agentId;
            }
        });
        
        // Show controls if hidden
        if (!this.showControls) {
            this.toggleControls();
        }
    }

    async handleControlAction(action) {
        const agentId = this.find(`[data-agent-input="${action}-agent"]`)?.value?.trim();
        if (!agentId) {
            this.showNotification('Agent ID is required', 'error');
            return;
        }

        try {
            let result;
            switch (action) {
                case 'assign':
                    const taskId = this.find(`[data-agent-input="${action}-task"]`)?.value?.trim();
                    if (!taskId) {
                        this.showNotification('Task ID is required', 'error');
                        return;
                    }
                    result = await window.StratavoreCore.apiClient.assignAgent(agentId, taskId);
                    break;
                    
                case 'complete':
                    const success = this.find(`[data-agent-input="${action}-success"]`)?.value === 'true';
                    const notes = this.find(`[data-agent-input="${action}-notes"]`)?.value?.trim() || '';
                    result = await window.StratavoreCore.apiClient.completeTask(agentId, success, notes);
                    break;
                    
                case 'status':
                    const status = this.find(`[data-agent-input="${action}-status"]`)?.value;
                    const thought = this.find(`[data-agent-input="${action}-thought"]`)?.value?.trim() || null;
                    result = await window.StratavoreCore.apiClient.updateAgentStatus(agentId, status, thought);
                    break;
                    
                case 'kill':
                    if (!confirm(`Kill agent ${agentId}?`)) return;
                    result = await window.StratavoreCore.apiClient.killAgent(agentId);
                    break;
                    
                default:
                    this.showNotification('Unknown action', 'error');
                    return;
            }

            if (result.status === 'success') {
                this.showNotification(`✅ ${action} completed for ${agentId}`, 'success');
                setTimeout(() => this.emit('refresh:requested'), 800);
            } else {
                this.showNotification(`❌ ${result.error}`, 'error');
            }
        } catch (error) {
            this.showNotification(`❌ Network error: ${error.message}`, 'error');
        }
    }

    async handleBatchAssign() {
        const taskId = prompt('Enter task ID for selected agents:');
        if (!taskId || this.selectedAgents.size === 0) return;
        
        try {
            const agentIds = Array.from(this.selectedAgents);
            const result = await window.StratavoreCore.apiClient.batchOperation('assign', agentIds, { taskId });
            
            if (result.status === 'success') {
                this.showNotification(`✅ Task ${taskId} assigned to ${agentIds.length} agents`, 'success');
                this.selectedAgents.clear();
                setTimeout(() => this.emit('refresh:requested'), 800);
            }
        } catch (error) {
            this.showNotification(`❌ Batch assign failed: ${error.message}`, 'error');
        }
    }

    async handleBatchStatusUpdate() {
        const status = prompt('Enter new status for selected agents (idle, working, paused, error):');
        if (!status || this.selectedAgents.size === 0) return;
        
        try {
            const agentIds = Array.from(this.selectedAgents);
            const result = await window.StratavoreCore.apiClient.batchOperation('status', agentIds, { status });
            
            if (result.status === 'success') {
                this.showNotification(`✅ Status updated for ${agentIds.length} agents`, 'success');
                this.selectedAgents.clear();
                setTimeout(() => this.emit('refresh:requested'), 800);
            }
        } catch (error) {
            this.showNotification(`❌ Batch status update failed: ${error.message}`, 'error');
        }
    }

    async handleBatchKill() {
        if (this.selectedAgents.size === 0) {
            this.showNotification('No agents selected', 'error');
            return;
        }
        
        if (!confirm(`Kill ${this.selectedAgents.size} selected agents?`)) return;
        
        try {
            const agentIds = Array.from(this.selectedAgents);
            const result = await window.StratavoreCore.apiClient.batchOperation('kill', agentIds);
            
            if (result.status === 'success') {
                this.showNotification(`✅ ${agentIds.length} agents terminated`, 'warning');
                this.selectedAgents.clear();
                setTimeout(() => this.emit('refresh:requested'), 800);
            }
        } catch (error) {
            this.showNotification(`❌ Batch kill failed: ${error.message}`, 'error');
        }
    }

    handleVirtualScroll() {
        const container = this.find('#agents-container');
        if (!container) return;
        
        const scrollTop = container.scrollTop;
        const containerHeight = container.clientHeight;
        const itemHeight = 120; // Approximate height of agent item
        
        const start = Math.floor(scrollTop / itemHeight);
        const visibleCount = Math.ceil(containerHeight / itemHeight);
        const end = start + visibleCount + 5; // Buffer for smooth scrolling
        
        this.visibleRange = { start: Math.max(0, start - 5), end };
        this.renderAgents();
    }

    updateSelectionUI() {
        this.findAll('[data-action="select-agent"]').forEach(checkbox => {
            const agentId = checkbox.dataset.agentId;
            checkbox.checked = this.selectedAgents.has(agentId);
        });

        const selectionInfo = this.find('#agent-selection-info');
        if (selectionInfo) {
            const count = this.selectedAgents.size;
            selectionInfo.innerHTML = count > 0 
                ? `${count} agent${count !== 1 ? 's' : ''} selected`
                : '';
        }
    }

    getFilteredAgents() {
        const agents = this.getData('agents') || {};
        return Object.entries(agents);
    }

    renderAgents() {
        const agentEntries = this.getFilteredAgents();
        const agentsContainer = this.find('#agents-container');
        const countEl = this.find('#agent-count');
        
        if (countEl) {
            countEl.textContent = agentEntries.length;
        }

        if (!agentsContainer) return;

        if (agentEntries.length === 0) {
            agentsContainer.innerHTML = `
                <div class="no-agents">
                    <p>No agents running. Click 🚀 SPAWN AGENT to start one.</p>
                </div>
            `;
            return;
        }

        // For large datasets, use virtual scrolling
        const shouldVirtualize = agentEntries.length > 100;
        const visibleEntries = shouldVirtualize 
            ? agentEntries.slice(this.visibleRange.start, this.visibleRange.end)
            : agentEntries;

        agentsContainer.innerHTML = `
            ${shouldVirtualize ? `<div style="height: ${this.visibleRange.start * 120}px;"></div>` : ''}
            ${visibleEntries.map(([agentId, agent]) => this.renderAgent(agentId, agent)).join('')}
            ${shouldVirtualize ? `<div style="height: ${(agentEntries.length - this.visibleRange.end) * 120}px;"></div>` : ''}
        `;

        this.updateSelectionUI();
    }

    renderAgent(agentId, agent) {
        const statusEmoji = this.getAgentStatusEmoji(agent.status);
        const personalityEmoji = this.getAgentPersonalityEmoji(agent.personality);
        const lastThought = (agent.thoughts && agent.thoughts.length)
            ? agent.thoughts[agent.thoughts.length - 1].thought 
            : 'No recent thoughts';
        const tasksComplete = (agent.metrics && agent.metrics.tasks_completed) || 0;
        const isSelected = this.selectedAgents.has(agentId);

        return `
            <div class="agent-item ${agent.status || ''} ${isSelected ? 'selected' : ''}" data-agent-id="${agentId}">
                <div class="agent-header">
                    <input type="checkbox" data-action="select-agent" data-agent-id="${agentId}" ${isSelected ? 'checked' : ''}>
                    <span class="agent-emoji">${statusEmoji}</span>
                    <span class="agent-emoji">${personalityEmoji}</span>
                    <strong>${agentId}</strong>
                    <span class="agent-personality">(${agent.personality || 'unknown'})</span>
                    <div class="agent-actions">
                        <button data-action="copy-agent-id" data-agent-id="${agentId}" class="mini-btn">use ↗</button>
                    </div>
                </div>
                <div class="agent-details">
                    <p><strong>Task:</strong> ${agent.current_task || 'No task assigned'}</p>
                    <p><strong>Thought:</strong> ${lastThought}</p>
                    <p><strong>Completed:</strong> ${tasksComplete} tasks</p>
                    <p><strong>Since:</strong> ${this.formatTimestamp(agent.created_at)}</p>
                </div>
            </div>
        `;
    }

    getAgentStatusEmoji(status) {
        const statusEmojis = {
            idle: '😴', working: '🔨', spawning: '🚀', completed: '✅', error: '❌', paused: '⏸️'
        };
        return statusEmojis[status] || '❓';
    }

    getAgentPersonalityEmoji(personality) {
        const personalityEmojis = {
            cadet: '👨‍🚀', senior: '👴', specialist: '🎯', researcher: '🔬', debugger: '🐛', optimizer: '⚡'
        };
        return personalityEmojis[personality] || '🤖';
    }

    isValidPersonality(personality) {
        const validPersonalities = ['cadet', 'senior', 'specialist', 'researcher', 'debugger', 'optimizer'];
        return validPersonalities.includes(personality);
    }

    formatTimestamp(timestamp) {
        if (!timestamp) return 'N/A';
        return new Date(timestamp).toLocaleString();
    }

    showNotification(message, type = 'info') {
        window.StratavoreCore.eventBus.emit('notification:show', { message, type });
    }

    async onMount() {
        this.render(`
            <div class="card">
                <div class="card-header">
                    <h3>🤖 AGENT STATUS &amp; WORKFLOW (<span id="agent-count">0</span>)</h3>
                    <div class="card-controls">
                        <button data-action="spawn-agent" class="primary-btn">🚀 SPAWN AGENT</button>
                        <button data-action="toggle-controls" class="secondary-btn">🎛 AGENT CONTROLS</button>
                    </div>
                </div>
                
                <div id="agent-selection-info" class="selection-info"></div>
                
                <!-- Agent controls panel -->
                <div id="agent-controls" class="control-panel" style="display: none;">
                    <h4>🎛 Agent Control Panel</h4>
                    
                    <div class="control-section">
                        <h5>Individual Controls</h5>
                        <div class="control-grid">
                            <!-- Assign Task -->
                            <form class="control-form" data-action="assign">
                                <h6>📋 Assign Task</h6>
                                <input data-agent-input="assign-agent" type="text" placeholder="Agent ID" required>
                                <input data-agent-input="assign-task" type="text" placeholder="Task ID" required>
                                <button type="submit">Assign</button>
                            </form>
                            
                            <!-- Complete Task -->
                            <form class="control-form" data-action="complete">
                                <h6>✅ Complete Task</h6>
                                <input data-agent-input="complete-agent" type="text" placeholder="Agent ID" required>
                                <select data-agent-input="complete-success">
                                    <option value="true">✅ Success</option>
                                    <option value="false">❌ Failed</option>
                                </select>
                                <input data-agent-input="complete-notes" type="text" placeholder="Notes (optional)">
                                <button type="submit">Complete</button>
                            </form>
                            
                            <!-- Update Status -->
                            <form class="control-form" data-action="status">
                                <h6>📡 Update Status</h6>
                                <input data-agent-input="status-agent" type="text" placeholder="Agent ID" required>
                                <select data-agent-input="status-status">
                                    <option value="idle">😴 Idle</option>
                                    <option value="working">🔨 Working</option>
                                    <option value="paused">⏸️ Paused</option>
                                    <option value="error">❌ Error</option>
                                </select>
                                <input data-agent-input="status-thought" type="text" placeholder="Thought (optional)">
                                <button type="submit">Update</button>
                            </form>
                            
                            <!-- Kill Agent -->
                            <form class="control-form" data-action="kill">
                                <h6>🔴 Kill Agent</h6>
                                <input data-agent-input="kill-agent" type="text" placeholder="Agent ID" required>
                                <button type="submit" class="danger-btn">Kill</button>
                            </form>
                        </div>
                    </div>
                    
                    <div class="control-section">
                        <h5>Batch Operations (selected agents)</h5>
                        <div class="batch-controls">
                            <button data-action="batch-assign" class="batch-btn">📋 Batch Assign</button>
                            <button data-action="batch-status" class="batch-btn">📡 Batch Status</button>
                            <button data-action="batch-kill" class="batch-btn danger">🔴 Batch Kill</button>
                        </div>
                    </div>
                </div>
                
                <div id="agents-container" class="agents-container">
                    <div class="loading">🔄 Loading agents...</div>
                </div>
            </div>
        `);

        this.setupEventListeners();
        this.renderAgents();
    }

    onUnmount() {
        this.cleanupEventListeners();
        this.cleanupSubscriptions();
    }
}

// Export for use in modules
window.AgentStatusComponent = AgentStatusComponent;