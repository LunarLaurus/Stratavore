/**
 * Jobs List Component - Displays and manages job listings
 */
class JobsListComponent extends BaseComponent {
    constructor(elementId, options = {}) {
        super(elementId, options);
        this.currentPage = 1;
        this.pageSize = 20;
        this.sortBy = 'priority';
        this.sortOrder = 'asc';
        this.filterStatus = 'active'; // active, all, completed
        this.selectedJobs = new Set();
    }

    setupEventListeners() {
        // Pagination controls
        this.addEventListener(this.element, 'click', (e) => {
            if (e.target.matches('[data-action="prev-page"]')) {
                this.prevPage();
            } else if (e.target.matches('[data-action="next-page"]')) {
                this.nextPage();
            } else if (e.target.matches('[data-action="sort"]')) {
                this.setSort(e.target.dataset.sort);
            } else if (e.target.matches('[data-action="filter"]')) {
                this.setFilter(e.target.dataset.filter);
            } else if (e.target.matches('[data-action="select-job"]')) {
                this.toggleJobSelection(e.target.dataset.jobId);
            } else if (e.target.matches('[data-action="refresh-jobs"]')) {
                this.refreshJobs();
            }
        });

        // Real-time search
        const searchInput = this.find('#job-search');
        if (searchInput) {
            this.addEventListener(searchInput, 'input', 
                this.debounce((e) => this.handleSearch(e.target.value), 300)
            );
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'f') {
                e.preventDefault();
                searchInput?.focus();
            }
        });
    }

    setupSubscriptions() {
        // Subscribe to job data changes
        this.subscribe('jobs', () => this.renderJobs());
        this.subscribe('timeSessions', () => this.renderJobs());

        // Listen for job events
        window.StratavoreCore.eventBus.on('job:created', () => this.refreshJobs());
        window.StratavoreCore.eventBus.on('job:updated', () => this.refreshJobs());
        window.StratavoreCore.eventBus.on('job:completed', () => this.refreshJobs());
    }

    handleSearch(query) {
        this.searchQuery = query.toLowerCase();
        this.currentPage = 1;
        this.renderJobs();
    }

    setSort(sortBy) {
        if (this.sortBy === sortBy) {
            this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
        } else {
            this.sortBy = sortBy;
            this.sortOrder = 'asc';
        }
        this.currentPage = 1;
        this.renderJobs();
    }

    setFilter(filter) {
        this.filterStatus = filter;
        this.currentPage = 1;
        this.selectedJobs.clear();
        this.renderJobs();
    }

    toggleJobSelection(jobId) {
        if (this.selectedJobs.has(jobId)) {
            this.selectedJobs.delete(jobId);
        } else {
            this.selectedJobs.add(jobId);
        }
        this.updateSelectionUI();
    }

    updateSelectionUI() {
        this.findAll('[data-action="select-job"]').forEach(checkbox => {
            const jobId = checkbox.dataset.jobId;
            checkbox.checked = this.selectedJobs.has(jobId);
        });

        const selectionInfo = this.find('#selection-info');
        if (selectionInfo) {
            const count = this.selectedJobs.size;
            selectionInfo.innerHTML = count > 0 
                ? `${count} job${count !== 1 ? 's' : ''} selected`
                : '';
        }
    }

    prevPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.renderJobs();
        }
    }

    nextPage() {
        const totalPages = Math.ceil(this.getFilteredJobs().length / this.pageSize);
        if (this.currentPage < totalPages) {
            this.currentPage++;
            this.renderJobs();
        }
    }

    refreshJobs() {
        this.emit('refresh:requested');
    }

    getFilteredJobs() {
        let jobs = this.getData('jobs') || [];
        
        // Apply status filter
        switch (this.filterStatus) {
            case 'active':
                jobs = jobs.filter(j => j.status !== 'completed');
                break;
            case 'completed':
                jobs = jobs.filter(j => j.status === 'completed');
                break;
            // 'all' shows all jobs
        }

        // Apply search filter
        if (this.searchQuery) {
            jobs = jobs.filter(j => 
                (j.title && j.title.toLowerCase().includes(this.searchQuery)) ||
                (j.id && j.id.toLowerCase().includes(this.searchQuery)) ||
                (j.assignee && j.assignee.toLowerCase().includes(this.searchQuery))
            );
        }

        return jobs;
    }

    getSortedJobs(jobs) {
        const priorityOrder = { high: 0, medium: 1, low: 2 };
        
        return jobs.sort((a, b) => {
            let aValue, bValue;
            
            switch (this.sortBy) {
                case 'priority':
                    aValue = priorityOrder[a.priority] || 3;
                    bValue = priorityOrder[b.priority] || 3;
                    break;
                case 'status':
                    aValue = a.status || '';
                    bValue = b.status || '';
                    break;
                case 'assignee':
                    aValue = a.assignee || '';
                    bValue = b.assignee || '';
                    break;
                case 'created_at':
                    aValue = new Date(a.created_at || 0);
                    bValue = new Date(b.created_at || 0);
                    break;
                case 'estimated_hours':
                    aValue = a.estimated_hours || 0;
                    bValue = b.estimated_hours || 0;
                    break;
                default:
                    aValue = a[this.sortBy] || '';
                    bValue = b[this.sortBy] || '';
            }
            
            if (aValue < bValue) return this.sortOrder === 'asc' ? -1 : 1;
            if (aValue > bValue) return this.sortOrder === 'asc' ? 1 : -1;
            return 0;
        });
    }

    getPaginatedJobs(jobs) {
        const startIndex = (this.currentPage - 1) * this.pageSize;
        const endIndex = startIndex + this.pageSize;
        return jobs.slice(startIndex, endIndex);
    }

    renderJobs() {
        const jobs = this.getFilteredJobs();
        const sortedJobs = this.getSortedJobs(jobs);
        const paginatedJobs = this.getPaginatedJobs(sortedJobs);
        const timeSessions = this.getData('timeSessions') || [];
        
        // Update job count
        const countEl = this.find('#active-job-count');
        if (countEl) {
            countEl.textContent = jobs.length;
        }

        // Render jobs list
        const jobsContainer = this.find('#jobs-container');
        if (!jobsContainer) return;

        if (paginatedJobs.length === 0) {
            jobsContainer.innerHTML = `
                <div class="loading">
                    ${jobs.length === 0 ? '✅ No jobs found' : '🔍 No jobs match current filters'}
                </div>
            `;
            return;
        }

        jobsContainer.innerHTML = paginatedJobs.map(job => {
            const jt = this.calculateJobTime(job.id, timeSessions);
            const deps = Array.isArray(job.dependencies) && job.dependencies.length
                ? `<p><strong>🔗 Dependencies:</strong> ${job.dependencies.join(', ')}</p>` : '';
            
            return `
                <div class="job ${job.priority || ''} ${this.getStatusClass(job.status)}" data-job-id="${job.id}">
                    <div class="job-header">
                        <input type="checkbox" data-action="select-job" data-job-id="${job.id}">
                        <div class="job-title">🏷️ ${job.title || job.id}</div>
                    </div>
                    <div class="job-meta">
                        <p><strong>ID:</strong> ${job.id}</p>
                        <p><strong>Status:</strong> ${(job.status || 'unknown').replace('_', ' ').toUpperCase()}</p>
                        <p><strong>Priority:</strong> ${(job.priority || 'unknown').toUpperCase()}</p>
                        <p><strong>Agent:</strong> ${job.assignee || 'UNASSIGNED'}</p>
                        <p><strong>Estimate:</strong> ${job.estimated_hours != null ? job.estimated_hours + 'h' : 'N/A'}</p>
                        <p><strong>Actual Time:</strong> ${jt.formattedTime}</p>
                        <p><strong>Created:</strong> ${this.formatTimestamp(job.created_at)}</p>
                        ${deps}
                    </div>
                </div>
            `;
        }).join('');

        this.renderPagination(jobs.length);
        this.updateSelectionUI();
    }

    renderPagination(totalJobs) {
        const totalPages = Math.ceil(totalJobs / this.pageSize);
        const paginationEl = this.find('#pagination');
        
        if (!paginationEl || totalPages <= 1) {
            if (paginationEl) paginationEl.innerHTML = '';
            return;
        }

        const startJob = (this.currentPage - 1) * this.pageSize + 1;
        const endJob = Math.min(this.currentPage * this.pageSize, totalJobs);

        paginationEl.innerHTML = `
            <div class="pagination">
                <button data-action="prev-page" ${this.currentPage === 1 ? 'disabled' : ''}>
                    ← Previous
                </button>
                <span class="pagination-info">
                    Showing ${startJob}-${endJob} of ${totalJobs} jobs
                </span>
                <button data-action="next-page" ${this.currentPage === totalPages ? 'disabled' : ''}>
                    Next →
                </button>
            </div>
        `;
    }

    calculateJobTime(jobId, timeSessions) {
        let totalSeconds = 0;
        const now = Date.now() / 1000;
        
        timeSessions
            .filter(s => s.job_id === jobId)
            .forEach(s => {
                if (s.status === 'completed' && s.duration_seconds) {
                    totalSeconds += s.duration_seconds;
                } else if (s.status === 'active') {
                    totalSeconds += (now - s.start_timestamp - (s.paused_time || 0));
                }
            });

        return {
            totalSeconds,
            totalHours: totalSeconds / 3600,
            formattedTime: this.formatDuration(totalSeconds)
        };
    }

    formatDuration(seconds) {
        if (!seconds || seconds < 0) return '0:00:00';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = Math.floor(seconds % 60);
        return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
    }

    formatTimestamp(timestamp) {
        if (!timestamp) return 'N/A';
        return new Date(timestamp).toLocaleString();
    }

    getStatusClass(status) {
        return status ? status.replace('_', '-') : 'unknown';
    }

    async onMount() {
        this.render(`
            <div class="card">
                <div class="card-header">
                    <h3>📋 ACTIVE JOBS (<span id="active-job-count">0</span>)</h3>
                    <div class="card-controls">
                        <input type="text" id="job-search" placeholder="Search jobs..." class="search-input">
                        <select data-action="filter" class="filter-select">
                            <option value="active">Active Jobs</option>
                            <option value="all">All Jobs</option>
                            <option value="completed">Completed Jobs</option>
                        </select>
                        <button data-action="sort" data-sort="priority" class="sort-btn">Priority</button>
                        <button data-action="sort" data-sort="created_at" class="sort-btn">Created</button>
                        <button data-action="refresh-jobs" class="refresh-btn">🔄</button>
                    </div>
                </div>
                <div id="selection-info" class="selection-info"></div>
                <div class="job-list" id="jobs-container">
                    <div class="loading">🔄 Loading jobs...</div>
                </div>
                <div id="pagination"></div>
            </div>
        `);

        // Setup event listeners after rendering
        this.setupEventListeners();
        
        // Initial render
        this.renderJobs();
    }

    onUnmount() {
        this.cleanupEventListeners();
        this.cleanupSubscriptions();
    }
}

// Export for use in modules
window.JobsListComponent = JobsListComponent;