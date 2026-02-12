/**
 * API Client for backend communication
 * Handles all HTTP requests to the Stratavore API
 */
class APIClient {
    constructor() {
        this.baseURL = '';
        this.requestTimeout = 30000; // 30 seconds
    }

    /**
     * Make HTTP request with error handling
     * @param {string} endpoint - API endpoint
     * @param {object} options - Request options
     * @returns {Promise} Response data
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), this.requestTimeout);

            const response = await fetch(url, {
                ...config,
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            
            if (data.status === 'error') {
                throw new Error(data.error || 'API error');
            }

            return data;
        } catch (error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    /**
     * GET request
     * @param {string} endpoint - API endpoint
     * @returns {Promise} Response data
     */
    async get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    }

    /**
     * POST request
     * @param {string} endpoint - API endpoint
     * @param {object} data - Request body data
     * @returns {Promise} Response data
     */
    async post(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    // API Methods

    /**
     * Get comprehensive status data
     * @returns {Promise} Status data with jobs, agents, time sessions
     */
    async getStatus() {
        return this.get('/api/status');
    }

    /**
     * Get system health information
     * @returns {Promise} Health data
     */
    async getHealth() {
        return this.get('/api/health');
    }

    /**
     * Get agent-focused data
     * @returns {Promise} Agent data with summary
     */
    async getAgents() {
        return this.get('/api/agents');
    }

    /**
     * Spawn a new agent
     * @param {string} personality - Agent personality type
     * @param {string} taskId - Optional task ID
     * @returns {Promise} Agent creation result
     */
    async spawnAgent(personality, taskId = null) {
        return this.post('/api/spawn-agent', { personality, task_id: taskId });
    }

    /**
     * Assign task to agent
     * @param {string} agentId - Agent ID
     * @param {string} taskId - Task ID
     * @returns {Promise} Assignment result
     */
    async assignAgent(agentId, taskId) {
        return this.post('/api/assign-agent', { agent_id: agentId, task_id: taskId });
    }

    /**
     * Complete agent task
     * @param {string} agentId - Agent ID
     * @param {boolean} success - Task success status
     * @param {string} notes - Optional completion notes
     * @returns {Promise} Completion result
     */
    async completeTask(agentId, success = true, notes = '') {
        return this.post('/api/complete-task', { 
            agent_id: agentId, 
            success, 
            notes 
        });
    }

    /**
     * Update agent status
     * @param {string} agentId - Agent ID
     * @param {string} status - New status
     * @param {string} thought - Optional thought/status note
     * @returns {Promise} Status update result
     */
    async updateAgentStatus(agentId, status, thought = null) {
        return this.post('/api/agent-status', { 
            agent_id: agentId, 
            status, 
            thought 
        });
    }

    /**
     * Kill/terminate an agent
     * @param {string} agentId - Agent ID
     * @returns {Promise} Kill result
     */
    async killAgent(agentId) {
        return this.post('/api/kill-agent', { agent_id: agentId });
    }

    /**
     * Batch operation for multiple agents
     * @param {string} operation - Operation type
     * @param {array} agentIds - Array of agent IDs
     * @param {object} params - Additional parameters
     * @returns {Promise} Batch operation results
     */
    async batchOperation(operation, agentIds, params = {}) {
        return this.post('/api/batch-operation', {
            operation,
            agent_ids: agentIds,
            ...params
        });
    }
}

/**
 * Real-time data manager
 * Handles polling and WebSocket connections
 */
class DataManager {
    constructor() {
        this.apiClient = new APIClient();
        this.pollingInterval = null;
        this.pollingFrequency = 30000; // 30 seconds
        this.isPolling = false;
    }

    /**
     * Start polling for data updates
     */
    startPolling() {
        if (this.isPolling) return;
        
        this.isPolling = true;
        this.pollingInterval = setInterval(async () => {
            try {
                await this.loadData();
            } catch (error) {
                console.error('Polling error:', error);
                window.StratavoreCore.dataStore.set('connectionStatus', 'error');
            }
        }, this.pollingFrequency);
        
        // Initial load
        this.loadData();
    }

    /**
     * Stop polling for data updates
     */
    stopPolling() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
        }
        this.isPolling = false;
    }

    /**
     * Load fresh data from API
     */
    async loadData() {
        try {
            window.StratavoreCore.dataStore.set('connectionStatus', 'loading');
            
            const data = await this.apiClient.getStatus();
            
            // Update data store with fresh data
            window.StratavoreCore.dataStore.update({
                jobs: data.jobs || [],
                agents: data.agents || {},
                agentTodos: data.agent_todos || [],
                timeSessions: data.time_sessions || [],
                progress: data.progress || {},
                lastUpdate: data.timestamp,
                connectionStatus: 'online'
            });

            // Emit data loaded event
            window.StratavoreCore.eventBus.emit('data:loaded', data);
            
        } catch (error) {
            console.error('Failed to load data:', error);
            window.StratavoreCore.dataStore.set('connectionStatus', 'error');
            window.StratavoreCore.eventBus.emit('data:error', error);
        }
    }

    /**
     * Check system health
     */
    async checkHealth() {
        try {
            const health = await this.apiClient.getHealth();
            window.StratavoreCore.eventBus.emit('health:updated', health);
            return health;
        } catch (error) {
            console.error('Health check failed:', error);
            window.StratavoreCore.dataStore.set('connectionStatus', 'error');
            return null;
        }
    }

    /**
     * Set polling frequency
     * @param {number} frequency - Polling frequency in milliseconds
     */
    setPollingFrequency(frequency) {
        this.pollingFrequency = frequency;
        if (this.isPolling) {
            this.stopPolling();
            this.startPolling();
        }
    }
}

// Global instances
const apiClient = new APIClient();
const dataManager = new DataManager();

// Export for use in modules
window.StratavoreCore.apiClient = apiClient;
window.StratavoreCore.dataManager = dataManager;