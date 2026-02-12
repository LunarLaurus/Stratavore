/**
 * Event Bus for component communication
 * Enables decoupled messaging between UI components
 */
class EventBus {
    constructor() {
        this.events = {};
    }

    /**
     * Subscribe to an event
     * @param {string} event - Event name
     * @param {function} callback - Event handler
     * @param {object} context - Optional context for callback
     */
    on(event, callback, context = null) {
        if (!this.events[event]) {
            this.events[event] = [];
        }
        this.events[event].push({ callback, context });
    }

    /**
     * Unsubscribe from an event
     * @param {string} event - Event name
     * @param {function} callback - Event handler to remove
     */
    off(event, callback) {
        if (!this.events[event]) return;
        this.events[event] = this.events[event].filter(
            handler => handler.callback !== callback
        );
    }

    /**
     * Emit an event
     * @param {string} event - Event name
     * @param {*} data - Event data payload
     */
    emit(event, data = null) {
        if (!this.events[event]) return;
        
        this.events[event].forEach(handler => {
            try {
                if (handler.context) {
                    handler.callback.call(handler.context, data);
                } else {
                    handler.callback(data);
                }
            } catch (error) {
                console.error(`Event handler error for "${event}":`, error);
            }
        });
    }

    /**
     * Clear all event listeners
     */
    clear() {
        this.events = {};
    }
}

/**
 * Data Store for centralized state management
 * Handles data caching and updates
 */
class DataStore {
    constructor() {
        this.state = {
            jobs: [],
            agents: {},
            agentTodos: [],
            timeSessions: [],
            progress: {},
            lastUpdate: null,
            connectionStatus: 'offline'
        };
        this.subscribers = {};
    }

    /**
     * Subscribe to state changes
     * @param {string} key - State key to watch (or null for all changes)
     * @param {function} callback - Change handler
     */
    subscribe(key, callback) {
        if (!this.subscribers[key]) {
            this.subscribers[key] = [];
        }
        this.subscribers[key].push(callback);
    }

    /**
     * Unsubscribe from state changes
     * @param {string} key - State key
     * @param {function} callback - Change handler to remove
     */
    unsubscribe(key, callback) {
        if (!this.subscribers[key]) return;
        this.subscribers[key] = this.subscribers[key].filter(
            handler => handler !== callback
        );
    }

    /**
     * Get state value
     * @param {string} key - State key
     * @returns {*} State value
     */
    get(key) {
        return this.state[key];
    }

    /**
     * Set state value and notify subscribers
     * @param {string} key - State key
     * @param {*} value - New value
     */
    set(key, value) {
        const oldValue = this.state[key];
        this.state[key] = value;
        
        // Notify subscribers to this specific key
        if (this.subscribers[key]) {
            this.subscribers[key].forEach(callback => {
                try {
                    callback(value, oldValue);
                } catch (error) {
                    console.error(`State subscriber error for "${key}":`, error);
                }
            });
        }

        // Notify subscribers to all changes
        if (this.subscribers[null]) {
            this.subscribers[null].forEach(callback => {
                try {
                    callback({ key, value, oldValue });
                } catch (error) {
                    console.error(`Global state subscriber error:`, error);
                }
            });
        }
    }

    /**
     * Update multiple state values atomically
     * @param {object} updates - Object with key-value pairs
     */
    update(updates) {
        Object.entries(updates).forEach(([key, value]) => {
            this.set(key, value);
        });
    }

    /**
     * Get entire state snapshot
     * @returns {object} Current state
     */
    getState() {
        return { ...this.state };
    }
}

// Global instances
const eventBus = new EventBus();
const dataStore = new DataStore();

// Export for use in modules
window.StratavoreCore = {
    eventBus,
    dataStore
};