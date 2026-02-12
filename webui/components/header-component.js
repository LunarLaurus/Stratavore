/**
 * Header Component - Application header with status and controls
 */
class HeaderComponent extends BaseComponent {
    constructor(elementId, options = {}) {
        super(elementId, options);
        this.refreshInterval = null;
    }

    setupEventListeners() {
        // Refresh button
        const refreshBtn = this.find('[data-action="refresh"]');
        if (refreshBtn) {
            this.addEventListener(refreshBtn, 'click', () => {
                this.handleRefresh();
            });
        }

        // Health check button
        const healthBtn = this.find('[data-action="health"]');
        if (healthBtn) {
            this.addEventListener(healthBtn, 'click', () => {
                this.handleHealthCheck();
            });
        }

        // Debug toggle button
        const debugBtn = this.find('[data-action="debug"]');
        if (debugBtn) {
            this.addEventListener(debugBtn, 'click', () => {
                this.toggleDebug();
            });
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'F5') {
                e.preventDefault();
                this.handleRefresh();
            } else if (e.ctrlKey && e.key === 'd') {
                e.preventDefault();
                this.toggleDebug();
            }
        });
    }

    setupSubscriptions() {
        // Connection status changes
        this.subscribe('connectionStatus', (status) => {
            this.updateConnectionStatus(status);
        });

        // Data loaded events
        window.StratavoreCore.eventBus.on('data:loaded', (data) => {
            this.updateDataStats(data);
        });

        // Data errors
        window.StratavoreCore.eventBus.on('data:error', (error) => {
            this.updateConnectionStatus('error');
        });

        // Health updates
        window.StratavoreCore.eventBus.on('health:updated', (health) => {
            this.updateHealthStatus(health);
        });
    }

    handleRefresh() {
        this.emit('refresh:requested');
        
        // Show immediate visual feedback
        const refreshBtn = this.find('[data-action="refresh"]');
        if (refreshBtn) {
            refreshBtn.disabled = true;
            refreshBtn.textContent = '🔄 Refreshing...';
        }

        // Re-enable after a delay
        setTimeout(() => {
            if (refreshBtn) {
                refreshBtn.disabled = false;
                refreshBtn.textContent = '🔄 REFRESH DATA';
            }
        }, 2000);
    }

    async handleHealthCheck() {
        const healthBtn = this.find('[data-action="health"]');
        if (healthBtn) {
            healthBtn.disabled = true;
            healthBtn.textContent = '🏥 Checking...';
        }

        try {
            await window.StratavoreCore.dataManager.checkHealth();
        } catch (error) {
            console.error('Health check failed:', error);
        } finally {
            if (healthBtn) {
                healthBtn.disabled = false;
                healthBtn.textContent = '🏥 HEALTH CHECK';
            }
        }
    }

    toggleDebug() {
        const debugBtn = this.find('[data-action="debug"]');
        const debugPanel = document.getElementById('debug-panel');
        
        if (debugPanel) {
            const isVisible = debugPanel.style.display !== 'none';
            debugPanel.style.display = isVisible ? 'none' : 'block';
            
            if (debugBtn) {
                debugBtn.textContent = isVisible ? '🔧 TOGGLE DEBUG' : '🔧 HIDE DEBUG';
            }
        }
    }

    updateConnectionStatus(status) {
        const indicator = this.find('#connection-status');
        const text = this.find('#connection-text');
        
        if (!indicator || !text) return;

        // Remove all status classes
        indicator.className = 'status-indicator';
        
        // Update based on status
        switch (status) {
            case 'online':
                indicator.classList.add('online');
                text.textContent = 'Connected';
                break;
            case 'loading':
                indicator.classList.add('warning');
                text.textContent = 'Loading...';
                break;
            case 'error':
                indicator.classList.add('error');
                text.textContent = 'Connection error';
                break;
            default:
                indicator.classList.add('error');
                text.textContent = 'Offline';
        }
    }

    updateDataStats(data) {
        const text = this.find('#connection-text');
        if (!text || !data) return;

        const jobCount = data.jobs?.length || 0;
        const agentCount = Object.keys(data.agents || {}).length;
        const sessionCount = data.time_sessions?.length || 0;
        
        text.textContent = `Connected (${jobCount} jobs · ${agentCount} agents · ${sessionCount} sessions)`;
    }

    updateHealthStatus(health) {
        const text = this.find('#connection-text');
        if (!text || !health) return;

        if (health.status === 'healthy') {
            const uptime = Math.floor(health.uptime || 0);
            text.textContent = `Server healthy (uptime: ${uptime}s)`;
        } else {
            text.textContent = 'Health check failed';
            this.updateConnectionStatus('error');
        }
    }

    async onMount() {
        // Render header HTML
        this.render(`
            <div class="header">
                <h1>🚀 STRATAVORE JOB TRACKER v4</h1>
                <p>Parallel agent workflows with real-time monitoring</p>
                <div style="margin: 20px 0;">
                    <span class="status-indicator" id="connection-status"></span>
                    <span id="connection-text">Checking connection...</span>
                    <button class="refresh-btn" data-action="refresh">🔄 REFRESH DATA</button>
                    <button class="refresh-btn" data-action="health">🏥 HEALTH CHECK</button>
                    <button class="refresh-btn" data-action="debug">🔧 TOGGLE DEBUG</button>
                </div>
            </div>
        `);

        // Setup event listeners after rendering
        this.setupEventListeners();
        
        // Initialize connection status
        this.updateConnectionStatus('loading');
    }

    onUnmount() {
        this.cleanupEventListeners();
        this.cleanupSubscriptions();
    }
}

// Export for use in modules
window.HeaderComponent = HeaderComponent;