/**
 * Overview Panel Component - Statistics dashboard
 * Displays job metrics, completion rates, and tracked hours
 */

class OverviewPanel {
    constructor(container) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.overview = {};
        this.refreshInterval = null;
        
        this.init();
        this.bindEvents();
    }
    
    init() {
        this.render();
        this.registerComponent();
        this.startAutoRefresh();
    }
    
    render() {
        this.container.innerHTML = `
            <div class="card">
                <h3>📊 OVERVIEW</h3>
                <div id="overview-stats" class="overview__stats">
                    <div class="loading">🔄 Loading overview...</div>
                </div>
            </div>
        `;
    }
    
    bindEvents() {
        // Subscribe to overview data changes
        window.dataStore?.subscribe('overview', (overview) => {
            this.overview = overview;
            this.renderOverview();
        }, { immediate: true });
        
        // Listen for refresh events
        window.eventBus?.on('header:refresh-requested', () => {
            this.showLoading();
        });
        
        // Listen for real-time updates
        window.eventBus?.on('state:changed', (event) => {
            if (event.key.startsWith('overview.') || event.key === 'jobs') {
                this.scheduleUpdate();
            }
        });
    }
    
    renderOverview() {
        const statsContainer = this.container.querySelector('#overview-stats');
        if (!statsContainer) return;
        
        if (!this.overview || Object.keys(this.overview).length === 0) {
            statsContainer.innerHTML = '<div class="loading">🔄 Loading overview...</div>';
            return;
        }
        
        const {
            totalJobs = 0,
            pendingJobs = 0,
            inProgressJobs = 0,
            completedJobs = 0,
            completionRate = 0,
            totalTrackedHours = 0
        } = this.overview;
        
        statsContainer.innerHTML = `
            <div class="stats-grid">
                <div class="stat-item">
                    <div class="stat-value">${totalJobs}</div>
                    <div class="stat-label">Total Jobs</div>
                </div>
                <div class="stat-item stat-item--warning">
                    <div class="stat-value">${pendingJobs}</div>
                    <div class="stat-label">🟡 Pending</div>
                </div>
                <div class="stat-item stat-item--info">
                    <div class="stat-value">${inProgressJobs}</div>
                    <div class="stat-label">🟠 In Progress</div>
                </div>
                <div class="stat-item stat-item--success">
                    <div class="stat-value">${completedJobs}</div>
                    <div class="stat-label">🟢 Completed</div>
                </div>
                <div class="stat-item stat-item--highlight">
                    <div class="stat-value">${completionRate}%</div>
                    <div class="stat-label">Completion Rate</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${totalTrackedHours}h</div>
                    <div class="stat-label">📊 Tracked Hours</div>
                </div>
            </div>
        `;
        
        // Animate the numbers
        this.animateNumbers();
    }
    
    animateNumbers() {
        const statValues = this.container.querySelectorAll('.stat-value');
        statValues.forEach((element, index) => {
            const finalValue = element.textContent;
            
            // Skip if not a number or percentage
            if (!/^\d+/.test(finalValue)) return;
            
            const numericValue = parseInt(finalValue);
            element.textContent = '0';
            
            setTimeout(() => {
                this.animateValue(element, 0, numericValue, 500);
            }, index * 100);
        });
    }
    
    animateValue(element, start, end, duration) {
        const startTime = performance.now();
        const isPercentage = element.textContent.includes('%');
        const hasDecimal = element.textContent.includes('.');
        const suffix = isPercentage ? '%' : (hasDecimal ? element.textContent.replace(/[\d.]/g, '') : '');
        
        const updateValue = (currentTime) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            // Easing function
            const easeOutQuart = 1 - Math.pow(1 - progress, 4);
            const currentValue = Math.floor(start + (end - start) * easeOutQuart);
            
            if (hasDecimal && end < 100) {
                element.textContent = (currentValue / 10).toFixed(1) + suffix;
            } else {
                element.textContent = currentValue + suffix;
            }
            
            if (progress < 1) {
                requestAnimationFrame(updateValue);
            }
        };
        
        requestAnimationFrame(updateValue);
    }
    
    showLoading() {
        const statsContainer = this.container.querySelector('#overview-stats');
        if (statsContainer) {
            statsContainer.innerHTML = '<div class="loading">🔄 Updating overview...</div>';
        }
    }
    
    scheduleUpdate() {
        // Debounce rapid updates
        if (this.updateTimeout) {
            clearTimeout(this.updateTimeout);
        }
        
        this.updateTimeout = setTimeout(() => {
            this.renderOverview();
        }, 300);
    }
    
    startAutoRefresh() {
        // Refresh overview every 10 seconds more frequently than full refresh
        this.refreshInterval = setInterval(() => {
            if (this.overview.totalJobs > 0) {
                this.renderOverview();
            }
        }, 10000);
    }
    
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }
    
    registerComponent() {
        window.eventBus?.registerComponent('overview-panel', this);
    }
    
    // Public methods
    refresh() {
        this.showLoading();
        setTimeout(() => this.renderOverview(), 500);
    }
    
    getOverview() {
        return this.overview;
    }
    
    destroy() {
        this.stopAutoRefresh();
        if (this.updateTimeout) {
            clearTimeout(this.updateTimeout);
        }
        window.eventBus?.unregisterComponent('overview-panel');
        this.container.innerHTML = '';
    }
}

// Auto-initialize if container exists
document.addEventListener('DOMContentLoaded', () => {
    const overviewContainer = document.querySelector('#overview');
    if (overviewContainer) {
        window.overviewPanel = new OverviewPanel(overviewContainer);
    }
});

export default OverviewPanel;