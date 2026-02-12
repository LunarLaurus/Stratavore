/**
 * Utility functions for the Stratavore Web UI
 */

/**
 * Format duration in seconds to human-readable format
 * @param {number} seconds - Duration in seconds
 * @returns {string} Formatted duration (HH:MM:SS)
 */
export function formatDuration(seconds) {
    if (!seconds || seconds < 0) return '0:00:00';
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

/**
 * Get status emoji for agent states
 * @param {string} status - Agent status
 * @returns {string} Status emoji
 */
export function getAgentStatusEmoji(status) {
    const statusEmojis = {
        idle: '😴',
        working: '🔨',
        spawning: '🚀',
        completed: '✅',
        error: '❌',
        paused: '⏸️'
    };
    return statusEmojis[status] || '❓';
}

/**
 * Get personality emoji for agent types
 * @param {string} personality - Agent personality
 * @returns {string} Personality emoji
 */
export function getAgentPersonalityEmoji(personality) {
    const personalityEmojis = {
        cadet: '👨‍🚀',
        senior: '👴',
        specialist: '🎯',
        researcher: '🔬',
        debugger: '🐛',
        optimizer: '⚡'
    };
    return personalityEmojis[personality] || '🤖';
}

/**
 * Get priority styling class
 * @param {string} priority - Priority level
 * @returns {string} CSS class name
 */
export function getPriorityClass(priority) {
    return priority || 'low';
}

/**
 * Get status styling class
 * @param {string} status - Status value
 * @returns {string} CSS class name
 */
export function getStatusClass(status) {
    return status ? status.replace('_', '-') : 'unknown';
}

/**
 * Calculate job time from time sessions
 * @param {string} jobId - Job ID
 * @param {array} timeSessions - Array of time session objects
 * @returns {object} Time calculation results
 */
export function calculateJobTime(jobId, timeSessions = []) {
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
        formattedTime: formatDuration(totalSeconds)
    };
}

/**
 * Generate summary statistics from data
 * @param {object} data - Data object with jobs, agents, timeSessions
 * @returns {object} Summary statistics
 */
export function generateSummaryStats(data) {
    const { jobs = [], agents = {}, timeSessions = [] } = data;
    
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

    return {
        totalJobs,
        pendingJobs,
        inProgressJobs,
        completedJobs,
        completionRate: totalJobs > 0 ? Math.round((completedJobs / totalJobs) * 100) : 0,
        agentCount,
        workingAgents,
        idleAgents,
        totalTrackedHours: totalTrackedHours.toFixed(2)
    };
}

/**
 * Validate agent personality
 * @param {string} personality - Personality string
 * @returns {boolean} Valid personality
 */
export function isValidPersonality(personality) {
    const validPersonalities = [
        'cadet', 'senior', 'specialist', 'researcher', 'debugger', 'optimizer'
    ];
    return validPersonalities.includes(personality?.toLowerCase());
}

/**
 * Validate agent status
 * @param {string} status - Status string
 * @returns {boolean} Valid status
 */
export function isValidAgentStatus(status) {
    const validStatuses = [
        'idle', 'working', 'paused', 'completed', 'error', 'spawning'
    ];
    return validStatuses.includes(status?.toLowerCase());
}

/**
 * Debounce function calls
 * @param {function} func - Function to debounce
 * @param {number} delay - Delay in milliseconds
 * @returns {function} Debounced function
 */
export function debounce(func, delay) {
    let timeoutId;
    return function (...args) {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
}

/**
 * Create a safe element ID from string
 * @param {string} str - Input string
 * @returns {string} Safe element ID
 */
export function createSafeId(str) {
    return str.replace(/[^a-zA-Z0-9]/g, '-').toLowerCase();
}

/**
 * Show temporary notification message
 * @param {string} message - Message to display
 * @param {string} type - Message type (success, error, warning, info)
 * @param {number} duration - Display duration in milliseconds
 */
export function showNotification(message, type = 'info', duration = 4000) {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    // Add styles if not already present
    if (!document.querySelector('#notification-styles')) {
        const style = document.createElement('style');
        style.id = 'notification-styles';
        style.textContent = `
            .notification {
                position: fixed;
                top: 20px;
                right: 20px;
                padding: 12px 20px;
                border-radius: 5px;
                font-weight: bold;
                z-index: 10000;
                opacity: 0;
                transform: translateX(100%);
                transition: all 0.3s ease;
            }
            .notification-success {
                background: #28a745;
                color: white;
                border: 1px solid #28a745;
            }
            .notification-error {
                background: #dc3545;
                color: white;
                border: 1px solid #dc3545;
            }
            .notification-warning {
                background: #ffc107;
                color: #212529;
                border: 1px solid #ffc107;
            }
            .notification-info {
                background: #17a2b8;
                color: white;
                border: 1px solid #17a2b8;
            }
            .notification.show {
                opacity: 1;
                transform: translateX(0);
            }
        `;
        document.head.appendChild(style);
    }
    
    document.body.appendChild(notification);
    
    // Trigger animation
    setTimeout(() => notification.classList.add('show'), 100);
    
    // Remove after duration
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, duration);
}

/**
 * Copy text to clipboard
 * @param {string} text - Text to copy
 * @returns {Promise} Copy result
 */
export async function copyToClipboard(text) {
    try {
        if (navigator.clipboard && window.isSecureContext) {
            await navigator.clipboard.writeText(text);
            return true;
        } else {
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = text;
            textArea.style.position = 'absolute';
            textArea.style.left = '-999999px';
            document.body.appendChild(textArea);
            textArea.select();
            const result = document.execCommand('copy');
            document.body.removeChild(textArea);
            return result;
        }
    } catch (error) {
        console.error('Failed to copy text:', error);
        return false;
    }
}

/**
 * Format timestamp to human-readable string
 * @param {number} timestamp - Unix timestamp
 * @returns {string} Formatted datetime
 */
export function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    return new Date(timestamp * 1000).toLocaleString();
}

/**
 * Generate unique ID
 * @returns {string} Unique identifier
 */
export function generateId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
}