/**
 * API Client - Centralized HTTP service for webui
 * Provides robust API communication with caching and error handling
 * 
 * Features:
 * - Request/response interceptors
 * - Automatic retry with exponential backoff
 * - Response caching with invalidation
 * - Request deduplication
 * - Real-time update integration
 */

class APIClient {
    constructor() {
        this.baseURL = '/api';
        this.defaultTimeout = 10000;
        this.maxRetries = 3;
        this.cache = new Map();
        this.requestQueue = new Map();
        this.interceptors = {
            request: [],
            response: [],
            error: []
        };
        
        // Performance metrics
        this.metrics = {
            totalRequests: 0,
            successfulRequests: 0,
            failedRequests: 0,
            cacheHits: 0,
            averageResponseTime: 0,
            retryCount: 0
        };
        
        // Setup default headers
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    /**
     * Make HTTP request with full feature support
     */
    async request(endpoint, options = {}) {
        const {
            method = 'GET',
            data = null,
            headers = {},
            timeout = this.defaultTimeout,
            retries = this.maxRetries,
            cache = method === 'GET', // Only cache GET requests by default
            cacheTTL = 30000,
            deduplicate = true,
            signal = null
        } = options;

        const requestId = this.generateRequestId(method, endpoint, data);
        
        // Request deduplication
        if (deduplicate && this.requestQueue.has(requestId)) {
            return this.requestQueue.get(requestId);
        }

        const startTime = performance.now();
        this.metrics.totalRequests++;

        const requestPromise = this.executeRequest(requestId, endpoint, {
            method,
            data,
            headers: { ...this.defaultHeaders, ...headers },
            timeout,
            retries,
            cache,
            cacheTTL,
            signal
        });

        // Store for deduplication
        if (deduplicate) {
            this.requestQueue.set(requestId, requestPromise);
            
            // Clean up after request completes
            requestPromise.finally(() => {
                this.requestQueue.delete(requestId);
            });
        }

        try {
            const response = await requestPromise;
            this.metrics.successfulRequests++;
            
            // Update performance metrics
            const responseTime = performance.now() - startTime;
            this.updateMetrics(responseTime);
            
            return response;
            
        } catch (error) {
            this.metrics.failedRequests++;
            throw error;
        }
    }

    /**
     * Execute the actual HTTP request with retry logic
     */
    async executeRequest(requestId, endpoint, options) {
        const { method, data, headers, timeout, retries, cache, cacheTTL, signal } = options;
        
        // Check cache first for GET requests
        if (cache && method === 'GET') {
            const cached = this.getCachedResponse(endpoint);
            if (cached) {
                this.metrics.cacheHits++;
                return cached;
            }
        }

        let lastError;
        
        for (let attempt = 0; attempt <= retries; attempt++) {
            try {
                // Apply request interceptors
                const processedOptions = await this.applyRequestInterceptors({
                    method,
                    headers,
                    data,
                    timeout,
                    attempt
                });

                // Make the request
                const response = await this.makeHTTPRequest(endpoint, processedOptions, signal);
                
                // Apply response interceptors
                const processedResponse = await this.applyResponseInterceptors(response);
                
                // Cache successful GET responses
                if (cache && method === 'GET' && response.ok) {
                    this.setCachedResponse(endpoint, processedResponse, cacheTTL);
                }
                
                return processedResponse;
                
            } catch (error) {
                lastError = error;
                
                // Don't retry on certain errors
                if (!this.shouldRetry(error, attempt, retries)) {
                    break;
                }
                
                // Exponential backoff
                const delay = Math.min(1000 * Math.pow(2, attempt), 10000);
                await this.delay(delay);
                
                this.metrics.retryCount++;
                console.warn(`Retrying request ${method} ${endpoint}, attempt ${attempt + 1}/${retries + 1}`);
            }
        }
        
        // Apply error interceptors
        return this.applyErrorInterceptors(lastError);
    }

    /**
     * Make the actual HTTP request using fetch
     */
    async makeHTTPRequest(endpoint, options, signal) {
        const { method, headers, data, timeout } = options;
        
        // Create AbortController for timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), timeout);
        
        // Chain abort signals if provided
        if (signal) {
            signal.addEventListener('abort', () => controller.abort());
        }
        
        try {
            const fetchOptions = {
                method,
                headers,
                signal: controller.signal
            };
            
            if (data && method !== 'GET') {
                fetchOptions.body = JSON.stringify(data);
            }
            
            const response = await fetch(this.baseURL + endpoint, fetchOptions);
            
            clearTimeout(timeoutId);
            
            // Parse response
            let responseData;
            const contentType = response.headers.get('content-type');
            
            if (contentType && contentType.includes('application/json')) {
                responseData = await response.json();
            } else {
                responseData = await response.text();
            }
            
            // Create enhanced response object
            const enhancedResponse = {
                ok: response.ok,
                status: response.status,
                statusText: response.statusText,
                headers: Object.fromEntries(response.headers.entries()),
                data: responseData,
                url: response.url,
                timestamp: Date.now()
            };
            
            if (!response.ok) {
                throw new APIError(response.status, responseData.message || response.statusText, enhancedResponse);
            }
            
            return enhancedResponse;
            
        } catch (error) {
            clearTimeout(timeoutId);
            
            if (error.name === 'AbortError') {
                throw new APIError(408, 'Request timeout');
            }
            
            throw error;
        }
    }

    /**
     * Specific API methods
     */
    
    // Data retrieval
    async getStatus() {
        return this.request('/status');
    }
    
    async getHealth() {
        return this.request('/health');
    }
    
    async getAgents() {
        return this.request('/agents');
    }
    
    // Agent management
    async spawnAgent(personality, taskId = null) {
        return this.request('/spawn-agent', {
            method: 'POST',
            data: { personality, task_id: taskId }
        });
    }
    
    async assignAgent(agentId, taskId) {
        return this.request('/assign-agent', {
            method: 'POST',
            data: { agent_id: agentId, task_id: taskId }
        });
    }
    
    async completeTask(agentId, success = true, notes = '') {
        return this.request('/complete-task', {
            method: 'POST',
            data: { agent_id: agentId, success, notes }
        });
    }
    
    async updateAgentStatus(agentId, status, thought = null) {
        return this.request('/agent-status', {
            method: 'POST',
            data: { agent_id: agentId, status, thought }
        });
    }
    
    async killAgent(agentId) {
        return this.request('/kill-agent', {
            method: 'POST',
            data: { agent_id: agentId }
        });
    }

    /**
     * Batch multiple requests
     */
    async batchRequests(requests) {
        const promises = requests.map(({ endpoint, options }) => 
            this.request(endpoint, options)
        );
        
        return Promise.allSettled(promises);
    }

    /**
     * Streaming updates (for future WebSocket integration)
     */
    async connectStream(endpoint, onData, onError) {
        try {
            const response = await fetch(this.baseURL + endpoint);
            
            if (!response.ok) {
                throw new Error(`Stream connection failed: ${response.status}`);
            }
            
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            
            while (true) {
                const { done, value } = await reader.read();
                
                if (done) break;
                
                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n').filter(line => line.trim());
                
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.substring(6);
                        try {
                            const parsed = JSON.parse(data);
                            onData(parsed);
                        } catch (error) {
                            console.warn('Invalid stream data:', data);
                        }
                    }
                }
            }
            
        } catch (error) {
            if (onError) onError(error);
            throw error;
        }
    }

    /**
     * Interceptor management
     */
    addRequestInterceptor(interceptor) {
        this.interceptors.request.push(interceptor);
    }
    
    addResponseInterceptor(interceptor) {
        this.interceptors.response.push(interceptor);
    }
    
    addErrorInterceptor(interceptor) {
        this.interceptors.error.push(interceptor);
    }

    /**
     * Apply interceptors
     */
    async applyRequestInterceptors(options) {
        let processedOptions = { ...options };
        
        for (const interceptor of this.interceptors.request) {
            processedOptions = await interceptor(processedOptions) || processedOptions;
        }
        
        return processedOptions;
    }
    
    async applyResponseInterceptors(response) {
        let processedResponse = response;
        
        for (const interceptor of this.interceptors.response) {
            processedResponse = await interceptor(processedResponse) || processedResponse;
        }
        
        return processedResponse;
    }
    
    async applyErrorInterceptors(error) {
        let processedError = error;
        
        for (const interceptor of this.interceptors.error) {
            try {
                processedError = await interceptor(processedError) || processedError;
            } catch (interceptorError) {
                console.error('Error in error interceptor:', interceptorError);
            }
        }
        
        throw processedError;
    }

    /**
     * Cache management
     */
    getCachedResponse(endpoint) {
        const cached = this.cache.get(endpoint);
        
        if (cached && Date.now() < cached.expires) {
            return cached.response;
        }
        
        if (cached) {
            this.cache.delete(endpoint);
        }
        
        return null;
    }
    
    setCachedResponse(endpoint, response, ttl) {
        this.cache.set(endpoint, {
            response,
            expires: Date.now() + ttl
        });
    }
    
    clearCache(endpoint = null) {
        if (endpoint) {
            this.cache.delete(endpoint);
        } else {
            this.cache.clear();
        }
    }

    /**
     * Utility methods
     */
    shouldRetry(error, attempt, maxRetries) {
        if (attempt >= maxRetries) return false;
        
        // Don't retry on client errors (4xx)
        if (error.status >= 400 && error.status < 500) return false;
        
        // Retry on network errors and server errors (5xx)
        return true;
    }
    
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    generateRequestId(method, endpoint, data) {
        const dataHash = data ? JSON.stringify(data) : '';
        return `${method}:${endpoint}:${btoa(dataHash)}`;
    }
    
    updateMetrics(responseTime) {
        const total = this.metrics.averageResponseTime * (this.metrics.successfulRequests - 1) + responseTime;
        this.metrics.averageResponseTime = total / this.metrics.successfulRequests;
    }

    /**
     * Get client statistics
     */
    getMetrics() {
        return {
            ...this.metrics,
            averageResponseTime: this.metrics.averageResponseTime.toFixed(2) + 'ms',
            cacheSize: this.cache.size,
            pendingRequests: this.requestQueue.size,
            successRate: this.metrics.totalRequests > 0 
                ? Math.round((this.metrics.successfulRequests / this.metrics.totalRequests) * 100) + '%'
                : '0%'
        };
    }

    /**
     * Reset metrics
     */
    resetMetrics() {
        this.metrics = {
            totalRequests: 0,
            successfulRequests: 0,
            failedRequests: 0,
            cacheHits: 0,
            averageResponseTime: 0,
            retryCount: 0
        };
    }
}

/**
 * Custom API Error class
 */
class APIError extends Error {
    constructor(status, message, response = null) {
        super(message);
        this.name = 'APIError';
        this.status = status;
        this.response = response;
    }
}

// Global API client instance
window.apiClient = new APIClient();

// Setup default interceptors
window.apiClient.addRequestInterceptor((options) => {
    // Add request timestamp for debugging
    options.requestTimestamp = Date.now();
    return options;
});

window.apiClient.addResponseInterceptor((response) => {
    // Emit successful response events
    window.eventBus?.emit('api:response', {
        url: response.url,
        status: response.status,
        timestamp: response.timestamp
    });
    
    return response;
});

window.apiClient.addErrorInterceptor((error) => {
    // Emit error events for global error handling
    window.eventBus?.emit('api:error', {
        error: error.message,
        status: error.status,
        timestamp: Date.now()
    });
    
    return error;
});

// Set up periodic cache cleanup
setInterval(() => {
    window.apiClient.clearCache();
}, 300000); // Clear cache every 5 minutes

export { APIClient, APIError };