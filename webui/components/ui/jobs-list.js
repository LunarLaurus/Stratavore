/**
 * Jobs List Component - Active jobs display with time tracking
 * Shows active jobs sorted by priority with time tracking information
 */

class JobsList {
    constructor(container) {
        this.container = typeof container === 'string' ? document.querySelector(container) : container;
        this.activeJobs = [];
        this.timeSessions = [];
        this.selectedJob = null;
        this.sortBy = 'priority'; // priority, created_at, title
        this.filterBy = 'all'; // all, high, medium, low
        
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
                <div class="jobs-header">
                    <h3>📋 ACTIVE JOBS (<span id="active-job-count">0</span>)</h3>
                    <div class="jobs-controls">
                        <select id="job-filter" class="form-input form-input--sm">
                            <option value="all">All Priorities</option>
                            <option value="high">🔴 High</option>
                            <option value="medium">🟡 Medium</option>
                            <option value="low">🔵 Low</option>
                        </select>
                        <select id="job-sort" class="form-input form-input--sm">
                            <option value="priority">Sort by Priority</option>
                            <option value="created_at">Sort by Created</option>
                            <option value="title">Sort by Title</option>
                        </select>
                    </div>
                </div>
                <div class="job-list scrollable scrollable--lg" id="jobs-container">
                    <div class="loading">🔄 Loading jobs...</div>
                </div>
            </div>
        `;
    }
    
    bindEvents() {
        // Subscribe to active jobs changes
        window.dataStore?.subscribe('activeJobs', (jobs) => {
            this.activeJobs = jobs;
            this.renderJobs();
        }, { immediate: true });
        
        // Subscribe to time sessions for time tracking
        window.dataStore?.subscribe('timeSessions', (sessions) => {
            this.timeSessions = sessions;
            this.renderJobs();
        });
        
        // Filter and sort controls
        this.container.addEventListener('change', (e) => {
            if (e.target.id === 'job-filter') {
                this.filterBy = e.target.value;
                this.renderJobs();
            } else if (e.target.id === 'job-sort') {
                this.sortBy = e.target.value;
                this.renderJobs();
            }
        });
        
        // Job item clicks
        this.container.addEventListener('click', (e) => {
            const jobItem = e.target.closest('.job-item');
            if (jobItem) {
                const jobId = jobItem.dataset.jobId;
                this.selectJob(jobId);
            }
        });
        
        // Listen for real-time updates
        window.eventBus?.on('job:updated', () => {
            this.renderJobs();
        });
    }
    
    renderJobs() {
        const jobsContainer = this.container.querySelector('#jobs-container');
        const countElement = this.container.querySelector('#active-job-count');
        
        if (!jobsContainer) return;
        
        // Filter jobs
        let filteredJobs = this.activeJobs.filter(job => {
            if (this.filterBy === 'all') return true;
            return job.priority === this.filterBy;
        });
        
        // Sort jobs
        filteredJobs.sort((a, b) => {
            switch (this.sortBy) {
                case 'priority':
                    const priorityOrder = { high: 0, medium: 1, low: 2 };
                    return priorityOrder[a.priority] - priorityOrder[b.priority];
                case 'created_at':
                    return new Date(b.created_at) - new Date(a.created_at);
                case 'title':
                    return a.title.localeCompare(b.title);
                default:
                    return 0;
            }
        });
        
        // Update count
        if (countElement) {
            countElement.textContent = filteredJobs.length;
        }
        
        // Render jobs
        if (filteredJobs.length === 0) {
            jobsContainer.innerHTML = '<div class="empty-state">✅ No active jobs matching filter</div>';
            return;
        }
        
        jobsContainer.innerHTML = filteredJobs.map(job => this.renderJob(job)).join('');
        
        // Animate job items
        this.animateJobItems();
    }
    
    renderJob(job) {
        const timeInfo = this.calculateJobTime(job.id);
        const isSelected = this.selectedJob === job.id;
        
        return `
            <div class="job-item job-item--${job.priority} job-item--${job.status?.replace('_', '-') || 'unknown'} ${isSelected ? 'job-item--selected' : ''}" 
                 data-job-id="${job.id}">
                <div class="job-header">
                    <div class="job-title">🏷️ ${job.title || job.id}</div>
                    <div class="job-priority-indicator priority-${job.priority}"></div>
                </div>
                <div class="job-meta">
                    <div class="job-info-row">
                        <span class="job-info-label">ID:</span>
                        <span class="job-info-value">${job.id}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Status:</span>
                        <span class="job-info-value job-status--${job.status}">${(job.status || 'unknown').replace('_', ' ').toUpperCase()}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Priority:</span>
                        <span class="job-info-value priority-${job.priority}">${(job.priority || 'unknown').toUpperCase()}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Agent:</span>
                        <span class="job-info-value">${job.assignee || 'UNASSIGNED'}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Estimate:</span>
                        <span class="job-info-value">${job.estimated_hours != null ? job.estimated_hours + 'h' : 'N/A'}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Actual:</span>
                        <span class="job-info-value job-time">${timeInfo.formattedTime}</span>
                    </div>
                    <div class="job-info-row">
                        <span class="job-info-label">Created:</span>
                        <span class="job-info-value">${job.created_at ? this.formatDate(job.created_at) : 'N/A'}</span>
                    </div>
                    ${this.renderDependencies(job)}
                </div>
                <div class="job-actions">
                    ${job.assignee ? `<button class="btn btn--sm btn--outline" data-action="unassign" data-job-id="${job.id}">Unassign</button>` : ''}
                    <button class="btn btn--sm" data-action="edit" data-job-id="${job.id}">Edit</button>
                </div>
            </div>
        `;
    }
    
    renderDependencies(job) {
        if (!job.dependencies || !Array.isArray(job.dependencies) || job.dependencies.length === 0) {
            return '';
        }
        
        return `
            <div class="job-info-row">
                <span class="job-info-label">🔗 Dependencies:</span>
                <span class="job-info-value">${job.dependencies.join(', ')}</span>
            </div>
        `;
    }
    
    calculateJobTime(jobId) {
        let totalSeconds = 0;
        
        this.timeSessions.filter(s => s.job_id === jobId).forEach(session => {
            if (session.status === 'completed' && session.duration_seconds) {
                totalSeconds += session.duration_seconds;
            } else if (session.status === 'active') {
                totalSeconds += (Date.now() / 1000 - session.start_timestamp - (session.paused_time || 0));
            }
        });
        
        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);
        const seconds = Math.floor(totalSeconds % 60);
        
        return {
            totalSeconds,
            totalHours: totalSeconds / 3600,
            formattedTime: `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
        };
    }
    
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    }
    
    selectJob(jobId) {
        // Update selection
        this.selectedJob = jobId;
        
        // Update data store
        window.dataStore?.updateState('selectedJob', jobId);
        
        // Re-render to show selection
        this.renderJobs();
        
        // Emit selection event
        const job = this.activeJobs.find(j => j.id === jobId);
        window.eventBus?.emit('job:selected', { job });
        
        // Scroll to selected job
        const selectedElement = this.container.querySelector(`[data-job-id="${jobId}"]`);
        if (selectedElement) {
            selectedElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }
    
    animateJobItems() {
        const jobItems = this.container.querySelectorAll('.job-item');
        jobItems.forEach((item, index) => {
            item.style.opacity = '0';
            item.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                item.style.transition = 'all 0.3s ease';
                item.style.opacity = '1';
                item.style.transform = 'translateY(0)';
            }, index * 50);
        });
    }
    
    registerComponent() {
        window.eventBus?.registerComponent('jobs-list', this);
    }
    
    // Public methods
    refresh() {
        this.renderJobs();
    }
    
    getSelectedJob() {
        return this.selectedJob ? this.activeJobs.find(j => j.id === this.selectedJob) : null;
    }
    
    getFilteredJobs() {
        return this.activeJobs.filter(job => {
            if (this.filterBy === 'all') return true;
            return job.priority === this.filterBy;
        });
    }
    
    setFilter(filter) {
        this.filterBy = filter;
        const filterSelect = this.container.querySelector('#job-filter');
        if (filterSelect) filterSelect.value = filter;
        this.renderJobs();
    }
    
    setSort(sort) {
        this.sortBy = sort;
        const sortSelect = this.container.querySelector('#job-sort');
        if (sortSelect) sortSelect.value = sort;
        this.renderJobs();
    }
    
    destroy() {
        window.eventBus?.unregisterComponent('jobs-list');
        this.container.innerHTML = '';
    }
}

// Auto-initialize if container exists
document.addEventListener('DOMContentLoaded', () => {
    const jobsContainer = document.querySelector('#jobs-container');
    if (jobsContainer) {
        window.jobsList = new JobsList(jobsContainer.parentElement);
    }
});

export default JobsList;