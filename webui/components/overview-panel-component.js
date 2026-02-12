/**
 * Overview Panel Component - Shows summary statistics
 */
class OverviewPanelComponent extends BaseComponent {
    constructor(elementId, options = {}) {
        super(elementId, options);
        this.lastStats = null;
    }

    setupSubscriptions() {
        // Subscribe to data changes
        this.subscribe('jobs', () => this.updateOverview());
        this.subscribe('agents', () => this.updateOverview());
        this.subscribe('timeSessions', () => this.updateOverview());
        this.subscribe('progress', () => this.updateOverview());

        // Listen for data loaded events
        window.StratavoreCore.eventBus.on('data:loaded', () => {
            this.updateOverview();
        });
    }

    updateOverview() {
        const jobs = this.getData('jobs') || [];
        const agents = this.getData('agents') || {};
        const timeSessions = this.getData('timeSessions') || [];
        
        const stats = this.calculateStats(jobs, agents, timeSessions);
        
        if (JSON.stringify(stats) === JSON.stringify(this.lastStats)) {
            return; // No changes to render
        }
        
        this.lastStats = stats;
        this.renderOverview(stats);
    }

    calculateStats(jobs, agents, timeSessions) {
        const totalJobs = jobs.length;
        const pendingJobs = jobs.filter(j => j.status === 'pending').length;
        const inProgressJobs = jobs.filter(j => j.status === 'in_progress').length;
        const completedJobs = jobs.filter(j => j.status === 'completed').length;
        
        const agentCount = Object.keys(agents).length;
        const workingAgents = Object.values(agents).filter(a => a.status === 'working').length;
        const idleAgents = Object.values(agents).filter(a => a.status === 'idle').length;
        
        const totalTrackedHours = timeSessions.reduce((sum, s) => 
            sum + (s.duration_seconds || 0) / 3600, 0
        );

        const completionRate = totalJobs > 0 ? Math.round((completedJobs / totalJobs) * 100) : 0;

        return {
            totalJobs,
            pendingJobs,
            inProgressJobs,
            completedJobs,
            completionRate,
            agentCount,
            workingAgents,
            idleAgents,
            totalTrackedHours: totalTrackedHours.toFixed(2)
        };
    }

    renderOverview(stats) {
        const overviewEl = this.find('#overview-stats');
        if (!overviewEl) return;

        overviewEl.innerHTML = `
            <div class="overview-grid">
                <div class="stat-item">
                    <div class="stat-value">${stats.totalJobs}</div>
                    <div class="stat-label">Total Jobs</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value stat-pending">${stats.pendingJobs}</div>
                    <div class="stat-label">🟡 Pending</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value stat-progress">${stats.inProgressJobs}</div>
                    <div class="stat-label">🟠 In Progress</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value stat-completed">${stats.completedJobs}</div>
                    <div class="stat-label">🟢 Completed</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${stats.completionRate}%</div>
                    <div class="stat-label">Completion Rate</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${stats.agentCount}</div>
                    <div class="stat-label">🤖 Agents</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value stat-working">${stats.workingAgents}</div>
                    <div class="stat-label">🔨 Working</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value stat-idle">${stats.idleAgents}</div>
                    <div class="stat-label">😴 Idle</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${stats.totalTrackedHours}h</div>
                    <div class="stat-label">📊 Tracked Hours</div>
                </div>
            </div>
        `;
    }

    async onMount() {
        this.render(`
            <div class="card">
                <h3>📊 OVERVIEW</h3>
                <div id="overview-stats">
                    <div class="loading">🔄 Loading overview...</div>
                </div>
            </div>
        `);

        // Initial update
        this.updateOverview();
    }

    onUnmount() {
        this.cleanupSubscriptions();
    }
}

// Export for use in modules
window.OverviewPanelComponent = OverviewPanelComponent;