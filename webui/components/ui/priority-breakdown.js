/**
 * Priority Breakdown Component - Active job priority distribution
 * Shows high/medium/low priority counts and visual breakdown
 */

class PriorityBreakdown {
    constructor(container) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.breakdown = {};
        this.chartInstance = null;
        
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
                <h3>🎯 PRIORITY BREAKDOWN</h3>
                <div id="priority-stats" class="priority__stats">
                    <div class="loading">🔄 Loading priorities...</div>
                </div>
                <div id="priority-chart" class="priority__chart"></div>
            </div>
        `;
    }
    
    bindEvents() {
        // Subscribe to priority breakdown changes
        window.dataStore?.subscribe('priorityBreakdown', (breakdown) => {
            this.breakdown = breakdown;
            this.renderBreakdown();
        }, { immediate: true });
        
        // Listen for job changes that might affect priorities
        window.eventBus?.on('state:changed', (event) => {
            if (event.key === 'activeJobs' || event.key === 'jobs') {
                this.scheduleUpdate();
            }
        });
    }
    
    renderBreakdown() {
        const statsContainer = this.container.querySelector('#priority-stats');
        const chartContainer = this.container.querySelector('#priority-chart');
        
        if (!statsContainer || !chartContainer) return;
        
        if (!this.breakdown || Object.keys(this.breakdown).length === 0) {
            statsContainer.innerHTML = '<div class="loading">🔄 Loading priorities...</div>';
            chartContainer.innerHTML = '';
            return;
        }
        
        const { high = 0, medium = 0, low = 0, total = 0 } = this.breakdown;
        
        // Render priority stats
        statsContainer.innerHTML = `
            <div class="priority-list">
                <div class="priority-item priority-item--high">
                    <div class="priority-indicator"></div>
                    <div class="priority-info">
                        <div class="priority-count">${high}</div>
                        <div class="priority-label">🔴 High Priority</div>
                    </div>
                    <div class="priority-percentage">${this.getPercentage(high, total)}%</div>
                </div>
                <div class="priority-item priority-item--medium">
                    <div class="priority-indicator"></div>
                    <div class="priority-info">
                        <div class="priority-count">${medium}</div>
                        <div class="priority-label">🟡 Medium Priority</div>
                    </div>
                    <div class="priority-percentage">${this.getPercentage(medium, total)}%</div>
                </div>
                <div class="priority-item priority-item--low">
                    <div class="priority-indicator"></div>
                    <div class="priority-info">
                        <div class="priority-count">${low}</div>
                        <div class="priority-label">🔵 Low Priority</div>
                    </div>
                    <div class="priority-percentage">${this.getPercentage(low, total)}%</div>
                </div>
            </div>
        `;
        
        // Render chart
        this.renderChart();
        
        // Animate numbers
        this.animateNumbers();
    }
    
    renderChart() {
        const chartContainer = this.container.querySelector('#priority-chart');
        if (!chartContainer) return;
        
        const { high = 0, medium = 0, low = 0 } = this.breakdown;
        const total = high + medium + low;
        
        if (total === 0) {
            chartContainer.innerHTML = '<div class="chart-empty">No active jobs</div>';
            return;
        }
        
        // Create simple bar chart
        const maxCount = Math.max(high, medium, low);
        
        chartContainer.innerHTML = `
            <div class="chart-bars">
                <div class="chart-bar" data-priority="high">
                    <div class="chart-bar-fill" style="width: ${(high / maxCount) * 100}%"></div>
                    <div class="chart-bar-label">${high}</div>
                </div>
                <div class="chart-bar" data-priority="medium">
                    <div class="chart-bar-fill" style="width: ${(medium / maxCount) * 100}%"></div>
                    <div class="chart-bar-label">${medium}</div>
                </div>
                <div class="chart-bar" data-priority="low">
                    <div class="chart-bar-fill" style="width: ${(low / maxCount) * 100}%"></div>
                    <div class="chart-bar-label">${low}</div>
                </div>
            </div>
        `;
        
        // Animate bars
        setTimeout(() => {
            const bars = chartContainer.querySelectorAll('.chart-bar-fill');
            bars.forEach(bar => {
                bar.style.transition = 'width 0.8s ease-out';
            });
        }, 100);
    }
    
    getPercentage(count, total) {
        return total > 0 ? Math.round((count / total) * 100) : 0;
    }
    
    animateNumbers() {
        const counts = this.container.querySelectorAll('.priority-count');
        counts.forEach((element, index) => {
            const finalValue = parseInt(element.textContent);
            element.textContent = '0';
            
            setTimeout(() => {
                this.animateValue(element, 0, finalValue, 600);
            }, index * 150);
        });
    }
    
    animateValue(element, start, end, duration) {
        const startTime = performance.now();
        
        const updateValue = (currentTime) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            const easeOutQuart = 1 - Math.pow(1 - progress, 4);
            const currentValue = Math.floor(start + (end - start) * easeOutQuart);
            
            element.textContent = currentValue;
            
            if (progress < 1) {
                requestAnimationFrame(updateValue);
            }
        };
        
        requestAnimationFrame(updateValue);
    }
    
    scheduleUpdate() {
        if (this.updateTimeout) {
            clearTimeout(this.updateTimeout);
        }
        
        this.updateTimeout = setTimeout(() => {
            this.renderBreakdown();
        }, 300);
    }
    
    registerComponent() {
        window.eventBus?.registerComponent('priority-breakdown', this);
    }
    
    // Public methods
    refresh() {
        this.renderBreakdown();
    }
    
    getBreakdown() {
        return this.breakdown;
    }
    
    getHighestPriority() {
        const { high, medium, low } = this.breakdown;
        if (high > 0) return 'high';
        if (medium > 0) return 'medium';
        if (low > 0) return 'low';
        return null;
    }
    
    destroy() {
        if (this.updateTimeout) {
            clearTimeout(this.updateTimeout);
        }
        window.eventBus?.unregisterComponent('priority-breakdown');
        this.container.innerHTML = '';
    }
}

// Auto-initialize if container exists
document.addEventListener('DOMContentLoaded', () => {
    const priorityContainer = document.querySelector('#priority-breakdown');
    if (priorityContainer) {
        window.priorityBreakdown = new PriorityBreakdown(priorityContainer);
    }
});

export default PriorityBreakdown;