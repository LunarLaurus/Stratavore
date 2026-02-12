/**
 * Data Store - Centralized state management for webui
 * Provides reactive state management with event integration
 * 
 * Features:
 * - Immutable state updates
 * - Reactive subscriptions
 * - Local storage persistence
 * - Data validation
 * - Performance optimization
 */

class DataStore {
    constructor() {
        this.state = this.getInitialState();
        this.subscribers = new Map();
        this.middleware = [];
        this.cache = new Map();
        this.persistenceKey = 'webui-state';
        this.lastUpdate = Date.now();
        
        // Subscribe to event bus for external updates
        window.eventBus?.on('data:update', (event) => {
            this.updateState(event.data.key, event.data.value, event.data.source);
        });
        
        // Load persisted state
        this.loadPersistedState();
        
        // Set up auto-persistence
        this.setupPersistence();
    }

    getInitialState() {
        return {
            // Jobs data
            jobs: [],
            activeJobs: [],
            completedJobs: [],
            
            // Agent data  
            agents: {},
            agentTodos: [],
            agentMetrics: {},
            
            // Time tracking
            timeSessions: [],
            activeSessions: [],
            
            // UI state
            connectionStatus: 'checking',
            lastUpdate: null,
            debugMode: false,
            selectedAgent: null,
            selectedJob: null,
            
            // Metrics and analytics
            overview: {
                totalJobs: 0,
                pendingJobs: 0,
                inProgressJobs: 0,
                completedJobs: 0,
                completionRate: 0,
                totalTrackedHours: 0
            },
            
            priorityBreakdown: {
                high: 0,
                medium: 0,
                low: 0,
                total: 0
            },
            
            // System state
            serverStatus: 'unknown',
            serverUptime: 0,
            apiStatus: 'unknown'
        };
    }

    /**
     * Get current state (immutable copy)
     */
    getState() {
        return JSON.parse(JSON.stringify(this.state));
    }

    /**
     * Get specific state property
     */
    get(key) {
        const keys = key.split('.');
        let value = this.state;
        
        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                return undefined;
            }
        }
        
        return value;
    }

    /**
     * Update state property immutably
     */
    updateState(key, value, source = 'user') {
        const oldValue = this.get(key);
        
        // Validate the update
        if (!this.validateUpdate(key, value, oldValue)) {
            throw new Error(`Invalid state update for ${key}`);
        }
        
        // Apply middleware
        const processedValue = this.applyMiddleware(key, value, oldValue);
        
        // Create new state immutably
        const newState = this.setStateProperty(key, processedValue);
        
        // Update state
        this.state = newState;
        this.lastUpdate = Date.now();
        
        // Notify subscribers
        this.notifySubscribers(key, processedValue, oldValue, source);
        
        // Emit event
        window.eventBus?.emit('state:changed', {
            key,
            value: processedValue,
            oldValue,
            source,
            timestamp: this.lastUpdate
        });
        
        // Cache invalidation
        this.invalidateCache(key);
        
        return true;
    }

    /**
     * Batch update multiple state properties
     */
    batchUpdate(updates, source = 'batch') {
        const changes = [];
        
        for (const { key, value } of updates) {
            const oldValue = this.get(key);
            
            if (this.validateUpdate(key, value, oldValue)) {
                const processedValue = this.applyMiddleware(key, value, oldValue);
                this.state = this.setStateProperty(key, processedValue);
                
                changes.push({
                    key,
                    value: processedValue,
                    oldValue
                });
            }
        }
        
        this.lastUpdate = Date.now();
        
        // Notify all subscribers for all changes
        changes.forEach(({ key, value, oldValue }) => {
            this.notifySubscribers(key, value, oldValue, source);
            this.invalidateCache(key);
        });
        
        // Emit batch change event
        window.eventBus?.emit('state:batch-changed', {
            changes,
            source,
            timestamp: this.lastUpdate
        });
        
        return changes;
    }

    /**
     * Subscribe to state changes
     */
    subscribe(key, callback, options = {}) {
        const {
            immediate = false,
            deep = false,
            transformer = null
        } = options;
        
        if (!this.subscribers.has(key)) {
            this.subscribers.set(key, []);
        }
        
        const subscription = {
            callback,
            deep,
            transformer,
            createdAt: Date.now()
        };
        
        this.subscribers.get(key).push(subscription);
        
        // Immediately call with current value if requested
        if (immediate) {
            const value = this.get(key);
            const transformedValue = transformer ? transformer(value) : value;
            callback(transformedValue, null, 'immediate');
        }
        
        // Return unsubscribe function
        return () => {
            const subscribers = this.subscribers.get(key);
            if (subscribers) {
                const index = subscribers.indexOf(subscription);
                if (index !== -1) {
                    subscribers.splice(index, 1);
                }
            }
        };
    }

    /**
     * Update state from API response
     */
    updateFromAPI(apiData) {
        const updates = [];
        
        // Process jobs
        if (apiData.jobs) {
            const activeJobs = apiData.jobs.filter(j => j.status !== 'completed');
            const completedJobs = apiData.jobs.filter(j => j.status === 'completed');
            
            updates.push(
                { key: 'jobs', value: apiData.jobs },
                { key: 'activeJobs', value: activeJobs },
                { key: 'completedJobs', value: completedJobs }
            );
        }
        
        // Process agents
        if (apiData.agents) {
            updates.push({ key: 'agents', value: apiData.agents });
        }
        
        if (apiData.agent_todos) {
            updates.push({ key: 'agentTodos', value: apiData.agent_todos });
        }
        
        // Process time sessions
        if (apiData.time_sessions) {
            const activeSessions = apiData.time_sessions.filter(s => s.status === 'active');
            
            updates.push(
                { key: 'timeSessions', value: apiData.time_sessions },
                { key: 'activeSessions', value: activeSessions }
            );
        }
        
        // Process metrics
        if (apiData.jobs) {
            const overview = this.calculateOverview(apiData.jobs, apiData.time_sessions || []);
            const priorityBreakdown = this.calculatePriorityBreakdown(apiData.jobs);
            
            updates.push(
                { key: 'overview', value: overview },
                { key: 'priorityBreakdown', value: priorityBreakdown }
            );
        }
        
        // Update timestamps
        if (apiData.timestamp) {
            updates.push(
                { key: 'lastUpdate', value: apiData.timestamp },
                { key: 'connectionStatus', value: 'connected' }
            );
        }
        
        return this.batchUpdate(updates, 'api');
    }

    /**
     * Calculate overview metrics
     */
    calculateOverview(jobs, timeSessions) {
        const total = jobs.length;
        const pending = jobs.filter(j => j.status === 'pending').length;
        const inProgress = jobs.filter(j => j.status === 'in_progress').length;
        const completed = jobs.filter(j => j.status === 'completed').length;
        
        const totalSeconds = timeSessions.reduce((s, session) => {
            if (session.status === 'completed' && session.duration_seconds) {
                return s + session.duration_seconds;
            } else if (session.status === 'active') {
                return s + (Date.now() / 1000 - session.start_timestamp - (session.paused_time || 0));
            }
            return s;
        }, 0);
        
        return {
            totalJobs: total,
            pendingJobs: pending,
            inProgressJobs: inProgress,
            completedJobs: completed,
            completionRate: total > 0 ? Math.round((completed / total) * 100) : 0,
            totalTrackedHours: Number((totalSeconds / 3600).toFixed(2))
        };
    }

    /**
     * Calculate priority breakdown
     */
    calculatePriorityBreakdown(jobs) {
        const active = jobs.filter(j => j.status !== 'completed');
        
        return {
            high: active.filter(j => j.priority === 'high').length,
            medium: active.filter(j => j.priority === 'medium').length,
            low: active.filter(j => j.priority === 'low').length,
            total: active.length
        };
    }

    /**
     * Add middleware for state transformations
     */
    addMiddleware(middleware) {
        this.middleware.push(middleware);
    }

    /**
     * Apply middleware to state updates
     */
    applyMiddleware(key, value, oldValue) {
        let processedValue = value;
        
        for (const middleware of this.middleware) {
            try {
                processedValue = middleware(key, processedValue, oldValue);
            } catch (error) {
                console.error('Middleware error:', error);
            }
        }
        
        return processedValue;
    }

    /**
     * Validate state updates
     */
    validateUpdate(key, value, oldValue) {
        // Basic type validation
        if (oldValue !== undefined && typeof value !== typeof oldValue) {
            console.warn(`Type mismatch for ${key}: expected ${typeof oldValue}, got ${typeof value}`);
        }
        
        // Custom validation rules
        switch (key) {
            case 'jobs':
            case 'agentTodos':
            case 'timeSessions':
                return Array.isArray(value);
                
            case 'agents':
            case 'overview':
            case 'priorityBreakdown':
                return typeof value === 'object' && value !== null;
                
            case 'connectionStatus':
                return ['checking', 'connected', 'error', 'offline'].includes(value);
                
            default:
                return true;
        }
    }

    /**
     * Set nested state property immutably
     */
    setStateProperty(key, value) {
        const keys = key.split('.');
        const newState = JSON.parse(JSON.stringify(this.state));
        
        let current = newState;
        for (let i = 0; i < keys.length - 1; i++) {
            if (!(keys[i] in current)) {
                current[keys[i]] = {};
            }
            current = current[keys[i]];
        }
        
        current[keys[keys.length - 1]] = value;
        return newState;
    }

    /**
     * Notify subscribers of state changes
     */
    notifySubscribers(key, value, oldValue, source) {
        // Notify exact matches
        if (this.subscribers.has(key)) {
            this.subscribers.get(key).forEach(sub => {
                try {
                    const transformedValue = sub.transformer ? sub.transformer(value) : value;
                    sub.callback(transformedValue, oldValue, source);
                } catch (error) {
                    console.error('Subscriber callback error:', error);
                }
            });
        }
        
        // Notify deep subscribers (subscribers to parent keys)
        for (const [subKey, subscribers] of this.subscribers) {
            if (key.startsWith(subKey + '.') && subscribers.some(s => s.deep)) {
                subscribers.filter(s => s.deep).forEach(sub => {
                    try {
                        const transformedValue = sub.transformer ? sub.transformer(this.get(subKey)) : this.get(subKey);
                        sub.callback(transformedValue, oldValue, source);
                    } catch (error) {
                        console.error('Deep subscriber error:', error);
                    }
                });
            }
        }
    }

    /**
     * Cache management
     */
    getFromCache(key) {
        return this.cache.get(key);
    }

    setCache(key, value, ttl = 30000) {
        this.cache.set(key, {
            value,
            expires: Date.now() + ttl
        });
    }

    invalidateCache(key) {
        // Clear specific key and all parent keys
        for (const cacheKey of this.cache.keys()) {
            if (cacheKey === key || cacheKey.startsWith(key + '.')) {
                this.cache.delete(cacheKey);
            }
        }
    }

    /**
     * Persistence to localStorage
     */
    setupPersistence() {
        // Save state changes
        this.subscribe('*', (value, oldValue, source) => {
            if (source !== 'persistence') {
                this.saveToStorage();
            }
        }, { deep: true });
        
        // Periodic cleanup
        setInterval(() => {
            this.cleanupStorage();
        }, 60000);
    }

    saveToStorage() {
        try {
            const toPersist = {
                uiState: {
                    debugMode: this.state.debugMode,
                    selectedAgent: this.state.selectedAgent,
                    selectedJob: this.state.selectedJob
                },
                lastSave: Date.now()
            };
            
            localStorage.setItem(this.persistenceKey, JSON.stringify(toPersist));
        } catch (error) {
            console.warn('Failed to persist state:', error);
        }
    }

    loadPersistedState() {
        try {
            const persisted = localStorage.getItem(this.persistenceKey);
            if (persisted) {
                const data = JSON.parse(persisted);
                
                if (data.uiState) {
                    Object.assign(this.state, data.uiState);
                }
            }
        } catch (error) {
            console.warn('Failed to load persisted state:', error);
        }
    }

    cleanupStorage() {
        try {
            const persisted = localStorage.getItem(this.persistenceKey);
            if (persisted) {
                const data = JSON.parse(persisted);
                
                // Clean up old data (older than 24 hours)
                if (data.lastSave && Date.now() - data.lastSave > 24 * 60 * 60 * 1000) {
                    localStorage.removeItem(this.persistenceKey);
                }
            }
        } catch (error) {
            // Ignore cleanup errors
        }
    }

    /**
     * Get store statistics
     */
    getStats() {
        return {
            stateSize: JSON.stringify(this.state).length,
            subscriberCount: Array.from(this.subscribers.values()).flat().length,
            cacheSize: this.cache.size,
            lastUpdate: this.lastUpdate,
            middlewareCount: this.middleware.length
        };
    }
}

// Global data store instance
window.dataStore = new DataStore();

export default DataStore;