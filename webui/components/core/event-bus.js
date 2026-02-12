/**
 * Event Bus - Central communication hub for webui components
 * Enables loose coupling between modular components
 * 
 * Provides:
 * - Event emission and subscription
 * - Component lifecycle management
 * - Error boundaries and recovery
 * - Performance monitoring
 */

class EventBus {
    constructor() {
        this.listeners = new Map();
        this.onceListeners = new Map();
        this.componentRegistry = new Map();
        this.errorHandlers = [];
        this.performanceMetrics = {
            eventsEmitted: 0,
            eventsProcessed: 0,
            averageProcessingTime: 0,
            errorCount: 0
        };
    }

    /**
     * Subscribe to events with optional component context
     * @param {string} eventType - Event type to listen for
     * @param {Function} callback - Event handler function
     * @param {Object} options - Options including context, priority, once
     */
    on(eventType, callback, options = {}) {
        const {
            context = null,
            priority = 0,
            once = false,
            timeout = null
        } = options;

        const listener = {
            callback,
            context,
            priority,
            timeout,
            createdAt: Date.now()
        };

        const targetMap = once ? this.onceListeners : this.listeners;
        
        if (!targetMap.has(eventType)) {
            targetMap.set(eventType, []);
        }
        
        targetMap.get(eventType).push(listener);
        
        // Sort by priority (higher priority first)
        targetMap.get(eventType).sort((a, b) => b.priority - a.priority);

        return () => this.off(eventType, callback);
    }

    /**
     * Subscribe to event only once
     */
    once(eventType, callback, options = {}) {
        return this.on(eventType, callback, { ...options, once: true });
    }

    /**
     * Unsubscribe from events
     */
    off(eventType, callback) {
        for (const map of [this.listeners, this.onceListeners]) {
            if (map.has(eventType)) {
                const listeners = map.get(eventType);
                const index = listeners.findIndex(l => l.callback === callback);
                if (index !== -1) {
                    listeners.splice(index, 1);
                }
            }
        }
    }

    /**
     * Emit event to all subscribers
     */
    async emit(eventType, data = {}, options = {}) {
        const {
            source = 'system',
            timestamp = Date.now(),
            batchId = null,
            timeout = 5000
        } = options;

        const startTime = performance.now();
        this.performanceMetrics.eventsEmitted++;

        const event = {
            type: eventType,
            data,
            source,
            timestamp,
            batchId,
            id: this.generateEventId()
        };

        try {
            // Process regular listeners
            await this.processListeners(this.listeners, eventType, event, timeout);
            
            // Process once listeners
            await this.processListeners(this.onceListeners, eventType, event, timeout);
            
            // Clear once listeners after processing
            this.onceListeners.delete(eventType);
            
            this.performanceMetrics.eventsProcessed++;
            
        } catch (error) {
            this.performanceMetrics.errorCount++;
            this.handleError(error, event);
        }

        // Update performance metrics
        const processingTime = performance.now() - startTime;
        this.updatePerformanceMetrics(processingTime);

        return event;
    }

    /**
     * Process listeners with timeout and error handling
     */
    async processListeners(listenersMap, eventType, event, timeout) {
        if (!listenersMap.has(eventType)) return;
        
        const listeners = listenersMap.get(eventType);
        const promises = listeners.map(async (listener) => {
            try {
                const startTime = performance.now();
                
                if (listener.timeout) {
                    // Add timeout wrapper
                    await Promise.race([
                        this.executeListener(listener, event),
                        new Promise((_, reject) => 
                            setTimeout(() => reject(new Error('Listener timeout')), listener.timeout)
                        )
                    ]);
                } else {
                    await this.executeListener(listener, event);
                }
                
                const executionTime = performance.now() - startTime;
                if (executionTime > 100) { // Log slow listeners
                    console.warn(`Slow listener detected: ${eventType} took ${executionTime.toFixed(2)}ms`);
                }
                
            } catch (error) {
                console.error(`Error in listener for ${eventType}:`, error);
                this.handleError(error, event, listener);
            }
        });

        await Promise.allSettled(promises);
    }

    /**
     * Execute listener with context binding
     */
    async executeListener(listener, event) {
        if (listener.context) {
            await listener.callback.call(listener.context, event);
        } else {
            await listener.callback(event);
        }
    }

    /**
     * Register component for lifecycle management
     */
    registerComponent(name, component) {
        this.componentRegistry.set(name, {
            component,
            registeredAt: Date.now(),
            eventCount: 0,
            errorCount: 0
        });

        // Set up component error boundary
        component.addEventListener = component.addEventListener || function(type, handler) {
            const wrappedHandler = (event) => {
                try {
                    handler(event);
                } catch (error) {
                    this.handleError(error, { type, source: name }, handler);
                }
            };
            
            return HTMLElement.prototype.addEventListener.call(this, type, wrappedHandler);
        };

        console.log(`Component registered: ${name}`);
    }

    /**
     * Unregister component
     */
    unregisterComponent(name) {
        if (this.componentRegistry.has(name)) {
            this.componentRegistry.delete(name);
            console.log(`Component unregistered: ${name}`);
        }
    }

    /**
     * Add global error handler
     */
    onError(handler) {
        this.errorHandlers.push(handler);
    }

    /**
     * Handle errors with registered error handlers
     */
    handleError(error, event, context = null) {
        const errorInfo = {
            error,
            event,
            context,
            timestamp: Date.now()
        };

        this.errorHandlers.forEach(handler => {
            try {
                handler(errorInfo);
            } catch (handlerError) {
                console.error('Error in error handler:', handlerError);
            }
        });

        // Default error handling
        console.error('EventBus Error:', errorInfo);
    }

    /**
     * Batch multiple events for efficiency
     */
    async emitBatch(events, options = {}) {
        const batchId = this.generateBatchId();
        const promises = events.map(({ type, data }) => 
            this.emit(type, data, { ...options, batchId })
        );
        
        return await Promise.all(promises);
    }

    /**
     * Get performance metrics
     */
    getMetrics() {
        return {
            ...this.performanceMetrics,
            activeListeners: Array.from(this.listeners.values()).flat().length,
            registeredComponents: this.componentRegistry.size,
            averageProcessingTime: this.performanceMetrics.averageProcessingTime.toFixed(2) + 'ms'
        };
    }

    /**
     * Clear all listeners (useful for cleanup)
     */
    clear() {
        this.listeners.clear();
        this.onceListeners.clear();
        this.componentRegistry.clear();
        this.performanceMetrics = {
            eventsEmitted: 0,
            eventsProcessed: 0,
            averageProcessingTime: 0,
            errorCount: 0
        };
    }

    // Private helper methods
    generateEventId() {
        return `evt_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    generateBatchId() {
        return `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    updatePerformanceMetrics(processingTime) {
        const total = this.performanceMetrics.averageProcessingTime * (this.performanceMetrics.eventsProcessed - 1) + processingTime;
        this.performanceMetrics.averageProcessingTime = total / this.performanceMetrics.eventsProcessed;
    }
}

// Global event bus instance
window.eventBus = new EventBus();

// Set up global error handling
window.eventBus.onError((errorInfo) => {
    // Log to debug panel if available
    if (window.debugPanel) {
        window.debugPanel.addError(errorInfo);
    }
    
    // Show user notification for critical errors
    if (errorInfo.error.name === 'CriticalError') {
        console.error('Critical system error:', errorInfo);
    }
});

// Performance monitoring
window.setInterval(() => {
    const metrics = window.eventBus.getMetrics();
    if (metrics.errorCount > 0) {
        console.warn('EventBus performance metrics:', metrics);
    }
}, 30000);

export default EventBus;