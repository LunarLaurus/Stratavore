/**
 * Header Component - Top navigation and controls
 * Handles connection status, refresh controls, and global actions
 */

class Header {
    constructor(container) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.connectionStatus = 'checking';
        this.lastUpdate = null;
        this.debugMode = false;
        
        this.init();
        this.bindEvents();
    }
    
    init() {
        this.render();
        this.registerComponent();
    }
    
    render() {
        this.container.innerHTML = `
            <div class="header">
                <h1>🚀 STRATAVORE JOB TRACKER v3</h1>
                <p>Parallel agent workflows with real-time monitoring</p>
                <div class="header__controls">
                    <span class="status-indicator" id="connection-status"></span>
                    <span id="connection-text">Checking connection...</span>
                    <button class="btn btn--sm" data-action="refresh-data">🔄 REFRESH DATA</button>
                    <button class="btn btn--sm btn--secondary" data-action="health-check">🏥 HEALTH CHECK</button>
                    <button class="btn btn--sm btn--outline" data-action="toggle-debug">🔧 TOGGLE DEBUG</button>
                </div>
            </div>
        `;
    }
    
    bindEvents() {
        // Button click handlers
        this.container.addEventListener('click', (e) => {
            if (e.target.dataset.action) {
                this.handleAction(e.target.dataset.action);
            }
        });
        
        // Subscribe to connection status changes
        window.dataStore?.subscribe('connectionStatus', (status) => {
            this.updateConnectionStatus(status);
        });
        
        // Subscribe to last update changes
        window.dataStore?.subscribe('lastUpdate', (timestamp) => {
            this.lastUpdate = timestamp;
            this.updateLastUpdated();
        });
        
        // Subscribe to debug mode changes
        window.dataStore?.subscribe('debugMode', (debugMode) => {
            this.debugMode = debugMode;
            this.updateDebugButton();
        });
    }
    
    handleAction(action) {
        switch (action) {
            case 'refresh-data':
                this.refreshData();
                break;
            case 'health-check':
                this.healthCheck();
                break;
            case 'toggle-debug':
                this.toggleDebug();
                break;
        }
    }
    
    async refreshData() {
        window.eventBus?.emit('header:refresh-requested');
        
        try {
            const response = await window.apiClient?.getStatus();
            if (response.ok) {
                window.dataStore?.updateFromAPI(response.data);
                this.showConnectionStatus('connected', `Loaded ${response.data.jobs?.length || 0} jobs`);
            }
        } catch (error) {
            this.showConnectionStatus('error', 'Failed to load data');
            window.eventBus?.emit('header:error', { error: error.message });
        }
    }
    
    async healthCheck() {
        try {
            const response = await window.apiClient?.getHealth();
            if (response.ok) {
                const uptime = Math.floor(response.data.uptime);
                this.showConnectionStatus('connected', `Server healthy (uptime: ${uptime}s)`);
                window.eventBus?.emit('header:health-check', response.data);
            }
        } catch (error) {
            this.showConnectionStatus('error', 'Health check failed');
            window.eventBus?.emit('header:error', { error: error.message });
        }
    }
    
    toggleDebug() {
        const newDebugMode = !this.debugMode;
        window.dataStore?.updateState('debugMode', newDebugMode);
        window.eventBus?.emit('header:debug-toggled', { debugMode: newDebugMode });
    }
    
    updateConnectionStatus(status) {
        this.connectionStatus = status;
        const indicator = this.container.querySelector('#connection-status');
        const text = this.container.querySelector('#connection-text');
        
        if (indicator && text) {
            indicator.className = `status-indicator status-indicator--${status}`;
            
            switch (status) {
                case 'connected':
                    text.textContent = 'Connected';
                    break;
                case 'error':
                    text.textContent = 'Connection error';
                    break;
                case 'checking':
                    text.textContent = 'Checking connection...';
                    break;
                default:
                    text.textContent = 'Offline';
            }
        }
    }
    
    showConnectionStatus(status, message) {
        this.updateConnectionStatus(status);
        const text = this.container.querySelector('#connection-text');
        if (text) text.textContent = message;
    }
    
    updateLastUpdated() {
        window.eventBus?.emit('header:last-updated', { timestamp: this.lastUpdate });
    }
    
    updateDebugButton() {
        const button = this.container.querySelector('[data-action="toggle-debug"]');
        if (button) {
            button.classList.toggle('btn--secondary', this.debugMode);
            button.classList.toggle('btn--outline', !this.debugMode);
        }
    }
    
    registerComponent() {
        window.eventBus?.registerComponent('header', this);
    }
    
    // Public methods
    setConnectionStatus(status, message) {
        this.showConnectionStatus(status, message);
    }
    
    getConnectionStatus() {
        return this.connectionStatus;
    }
    
    destroy() {
        window.eventBus?.unregisterComponent('header');
        this.container.innerHTML = '';
    }
}

// Auto-initialize if container exists
document.addEventListener('DOMContentLoaded', () => {
    const headerContainer = document.querySelector('.header-container');
    if (headerContainer) {
        window.headerComponent = new Header(headerContainer);
    }
});

export default Header;