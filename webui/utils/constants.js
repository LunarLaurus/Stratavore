/**
 * Application constants and configuration
 */

// API Configuration
export const API_CONFIG = {
    BASE_URL: '',
    TIMEOUT: 30000,
    POLLING_INTERVAL: 30000,
    RETRY_ATTEMPTS: 3,
    RETRY_DELAY: 1000
};

// Agent Configuration
export const AGENT_CONFIG = {
    PERSONALITIES: {
        CADET: { value: 'cadet', label: '👨‍🚀 Cadet', description: 'Learning agent with basic capabilities' },
        SENIOR: { value: 'senior', label: '👴 Senior', description: 'Experienced agent with advanced skills' },
        SPECIALIST: { value: 'specialist', label: '🎯 Specialist', description: 'Focused on specific task types' },
        RESEARCHER: { value: 'researcher', label: '🔬 Researcher', description: 'Exploratory and analytical agent' },
        DEBUGGER: { value: 'debugger', label: '🐛 Debugger', description: 'Problem-solving and troubleshooting' },
        OPTIMIZER: { value: 'optimizer', label: '⚡ Optimizer', description: 'Performance and efficiency focused' }
    },
    STATUSES: {
        IDLE: { value: 'idle', label: '😴 Idle', color: '#0066ff' },
        WORKING: { value: 'working', label: '🔨 Working', color: '#00ff00' },
        SPAWNING: { value: 'spawning', label: '🚀 Spawning', color: '#ffff00' },
        COMPLETED: { value: 'completed', label: '✅ Completed', color: '#28a745' },
        ERROR: { value: 'error', label: '❌ Error', color: '#ff0000' },
        PAUSED: { value: 'paused', label: '⏸️ Paused', color: '#ffc107' }
    }
};

// Job Configuration
export const JOB_CONFIG = {
    STATUSES: {
        PENDING: 'pending',
        IN_PROGRESS: 'in_progress',
        COMPLETED: 'completed',
        CANCELLED: 'cancelled'
    },
    PRIORITIES: {
        HIGH: { value: 'high', label: '🔴 High', color: '#ff0000' },
        MEDIUM: { value: 'medium', label: '🟡 Medium', color: '#ffff00' },
        LOW: { value: 'low', label: '🔵 Low', color: '#0066ff' }
    }
};

// UI Configuration
export const UI_CONFIG = {
    THEMES: {
        DARK: 'dark',
        LIGHT: 'light',
        TERMINAL: 'terminal'
    },
    NOTIFICATION_DURATION: 4000,
    ANIMATION_DURATION: 300,
    DEBOUNCE_DELAY: 300,
    VIRTUAL_SCROLL_THRESHOLD: 100,
    PAGE_SIZE: 50,
    MAX_VISIBLE_AGENTS: 1000
};

// Event Names
export const EVENTS = {
    // Data events
    DATA_LOADED: 'data:loaded',
    DATA_ERROR: 'data:error',
    DATA_UPDATED: 'data:updated',
    
    // Agent events
    AGENT_SPAWNED: 'agent:spawned',
    AGENT_UPDATED: 'agent:updated',
    AGENT_KILLED: 'agent:killed',
    AGENT_ASSIGNED: 'agent:assigned',
    AGENT_STATUS_CHANGED: 'agent:status-changed',
    
    // Job events
    JOB_CREATED: 'job:created',
    JOB_UPDATED: 'job:updated',
    JOB_COMPLETED: 'job:completed',
    
    // UI events
    CONNECTION_STATUS_CHANGED: 'connection:status-changed',
    HEALTH_UPDATED: 'health:updated',
    NOTIFICATION_SHOW: 'notification:show',
    MODAL_OPEN: 'modal:open',
    MODAL_CLOSE: 'modal:close',
    REFRESH_REQUESTED: 'refresh:requested',
    
    // Component events
    COMPONENT_MOUNTED: 'component:mounted',
    COMPONENT_UNMOUNTED: 'component:unmounted',
    COMPONENT_ERROR: 'component:error'
};

// API Endpoints
export const API_ENDPOINTS = {
    STATUS: '/api/status',
    HEALTH: '/api/health',
    AGENTS: '/api/agents',
    SPAWN_AGENT: '/api/spawn-agent',
    ASSIGN_AGENT: '/api/assign-agent',
    COMPLETE_TASK: '/api/complete-task',
    AGENT_STATUS: '/api/agent-status',
    KILL_AGENT: '/api/kill-agent',
    BATCH_OPERATION: '/api/batch-operation'
};

// Error Messages
export const ERROR_MESSAGES = {
    NETWORK_ERROR: 'Network connection error. Please check your connection.',
    TIMEOUT_ERROR: 'Request timeout. Please try again.',
    API_ERROR: 'API error occurred. Please try again later.',
    VALIDATION_ERROR: 'Invalid input provided.',
    PERMISSION_ERROR: 'Permission denied.',
    NOT_FOUND: 'Resource not found.',
    UNKNOWN_ERROR: 'An unknown error occurred.'
};

// Success Messages
export const SUCCESS_MESSAGES = {
    AGENT_SPAWNED: 'Agent spawned successfully!',
    AGENT_ASSIGNED: 'Task assigned to agent successfully!',
    AGENT_STATUS_UPDATED: 'Agent status updated successfully!',
    AGENT_KILLED: 'Agent terminated successfully!',
    TASK_COMPLETED: 'Task marked as completed!',
    BATCH_OPERATION_COMPLETED: 'Batch operation completed successfully!'
};

// Validation Rules
export const VALIDATION_RULES = {
    AGENT_ID: {
        required: true,
        minLength: 1,
        maxLength: 100,
        pattern: /^[a-zA-Z0-9_-]+$/
    },
    TASK_ID: {
        required: true,
        minLength: 1,
        maxLength: 100,
        pattern: /^[a-zA-Z0-9_-]+$/
    },
    PERSONALITY: {
        required: true,
        allowed: Object.keys(AGENT_CONFIG.PERSONALITIES)
    },
    STATUS: {
        required: true,
        allowed: Object.keys(AGENT_CONFIG.STATUSES)
    }
};

// Default Values
export const DEFAULTS = {
    POLLING_INTERVAL: 30000,
    NOTIFICATION_TYPE: 'info',
    AGENT_PERSONALITY: 'cadet',
    JOB_PRIORITY: 'medium',
    UI_THEME: 'dark',
    PAGE_SIZE: 50
};

// Performance Thresholds
export const PERFORMANCE_THRESHOLDS = {
    SLOW_API_RESPONSE: 5000, // 5 seconds
    HIGH_AGENT_COUNT: 500,
    LARGE_JOB_LIST: 1000,
    MEMORY_WARNING: 0.8, // 80%
    CPU_WARNING: 0.9 // 90%
};

// Keyboard Shortcuts
export const KEYBOARD_SHORTCUTS = {
    REFRESH: 'F5',
    SPAWN_AGENT: 'Ctrl+N',
    TOGGLE_DEBUG: 'Ctrl+D',
    TOGGLE_CONTROLS: 'Ctrl+C',
    FOCUS_SEARCH: 'Ctrl+F',
    ESCAPE: 'Escape'
};

// Local Storage Keys
export const STORAGE_KEYS = {
    THEME: 'stratavore-theme',
    POLLING_INTERVAL: 'stratavore-polling-interval',
    DEBUG_MODE: 'stratavore-debug-mode',
    PREFERENCES: 'stratavore-preferences',
    LAST_SESSION: 'stratavore-last-session'
};