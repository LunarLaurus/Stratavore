/**
 * Base Component class for all UI components
 * Provides lifecycle management, event handling, and state management
 */
class BaseComponent {
    constructor(elementId, options = {}) {
        this.elementId = elementId;
        this.element = document.getElementById(elementId);
        this.options = options;
        this.isMounted = false;
        this.eventListeners = [];
        this.subscriptions = [];
        this.children = [];
        
        if (!this.element) {
            console.warn(`Component element not found: ${elementId}`);
            return;
        }
        
        // Auto-mount if autoMount is not false
        if (options.autoMount !== false) {
            this.mount();
        }
    }

    /**
     * Mount component - initialize and render
     */
    async mount() {
        if (this.isMounted) return;
        
        try {
            await this.onMount();
            this.setupEventListeners();
            this.setupSubscriptions();
            this.isMounted = true;
            
            window.StratavoreCore.eventBus.emit('component:mounted', {
                component: this.constructor.name,
                elementId: this.elementId
            });
        } catch (error) {
            console.error(`Failed to mount component ${this.constructor.name}:`, error);
            this.handleError(error);
        }
    }

    /**
     * Unmount component - cleanup
     */
    unmount() {
        if (!this.isMounted) return;
        
        try {
            this.cleanupEventListeners();
            this.cleanupSubscriptions();
            this.children.forEach(child => child.unmount());
            this.children = [];
            
            this.onUnmount();
            this.isMounted = false;
            
            window.StratavoreCore.eventBus.emit('component:unmounted', {
                component: this.constructor.name,
                elementId: this.elementId
            });
        } catch (error) {
            console.error(`Failed to unmount component ${this.constructor.name}:`, error);
        }
    }

    /**
     * Setup component event listeners
     */
    setupEventListeners() {
        // Override in subclasses
    }

    /**
     * Cleanup event listeners
     */
    cleanupEventListeners() {
        this.eventListeners.forEach(({ element, event, handler }) => {
            element.removeEventListener(event, handler);
        });
        this.eventListeners = [];
    }

    /**
     * Setup subscriptions to data store
     */
    setupSubscriptions() {
        // Override in subclasses
    }

    /**
     * Cleanup subscriptions
     */
    cleanupSubscriptions() {
        this.subscriptions.forEach(({ key, callback }) => {
            window.StratavoreCore.dataStore.unsubscribe(key, callback);
        });
        this.subscriptions = [];
    }

    /**
     * Add event listener with automatic cleanup
     * @param {Element} element - Target element
     * @param {string} event - Event name
     * @param {function} handler - Event handler
     */
    addEventListener(element, event, handler) {
        if (element && event && handler) {
            element.addEventListener(event, handler);
            this.eventListeners.push({ element, event, handler });
        }
    }

    /**
     * Subscribe to data store changes
     * @param {string} key - Data key to watch
     * @param {function} callback - Change handler
     */
    subscribe(key, callback) {
        window.StratavoreCore.dataStore.subscribe(key, callback);
        this.subscriptions.push({ key, callback });
    }

    /**
     * Emit event through event bus
     * @param {string} event - Event name
     * @param {*} data - Event data
     */
    emit(event, data) {
        window.StratavoreCore.eventBus.emit(event, data);
    }

    /**
     * Get data from data store
     * @param {string} key - Data key
     * @returns {*} Data value
     */
    getData(key) {
        return window.StratavoreCore.dataStore.get(key);
    }

    /**
     * Set data in data store
     * @param {string} key - Data key
     * @param {*} value - Data value
     */
    setData(key, value) {
        window.StratavoreCore.dataStore.set(key, value);
    }

    /**
     * Show loading state
     */
    showLoading() {
        if (this.element) {
            this.element.innerHTML = '<div class="loading">🔄 Loading...</div>';
        }
    }

    /**
     * Show error state
     * @param {Error|string} error - Error to display
     */
    showError(error) {
        const message = error?.message || error || 'Unknown error occurred';
        if (this.element) {
            this.element.innerHTML = `<div class="error">❌ ${message}</div>`;
        }
        this.handleError(error);
    }

    /**
     * Handle component errors
     * @param {Error} error - Error object
     */
    handleError(error) {
        console.error(`Component error in ${this.constructor.name}:`, error);
        this.emit('component:error', {
            component: this.constructor.name,
            error,
            elementId: this.elementId
        });
    }

    /**
     * Render component HTML
     * @param {string} html - HTML content
     */
    render(html) {
        if (this.element) {
            this.element.innerHTML = html;
        }
    }

    /**
     * Add CSS class to element
     * @param {string} className - Class name to add
     */
    addClass(className) {
        if (this.element) {
            this.element.classList.add(className);
        }
    }

    /**
     * Remove CSS class from element
     * @param {string} className - Class name to remove
     */
    removeClass(className) {
        if (this.element) {
            this.element.classList.remove(className);
        }
    }

    /**
     * Toggle CSS class on element
     * @param {string} className - Class name to toggle
     */
    toggleClass(className) {
        if (this.element) {
            this.element.classList.toggle(className);
        }
    }

    /**
     * Check if element has CSS class
     * @param {string} className - Class name to check
     * @returns {boolean} Has class
     */
    hasClass(className) {
        return this.element ? this.element.classList.contains(className) : false;
    }

    /**
     * Hide element
     */
    hide() {
        if (this.element) {
            this.element.style.display = 'none';
        }
    }

    /**
     * Show element
     */
    show() {
        if (this.element) {
            this.element.style.display = '';
        }
    }

    /**
     * Set element display style
     * @param {string} display - Display value
     */
    setDisplay(display) {
        if (this.element) {
            this.element.style.display = display;
        }
    }

    /**
     * Find child element by selector
     * @param {string} selector - CSS selector
     * @returns {Element|null} Found element
     */
    find(selector) {
        return this.element ? this.element.querySelector(selector) : null;
    }

    /**
     * Find all child elements by selector
     * @param {string} selector - CSS selector
     * @returns {NodeList} Found elements
     */
    findAll(selector) {
        return this.element ? this.element.querySelectorAll(selector) : [];
    }

    /**
     * Create child component
     * @param {Function} ComponentClass - Component class
     * @param {string} elementId - Child element ID
     * @param {object} options - Component options
     * @returns {BaseComponent} Child component instance
     */
    createChild(ComponentClass, elementId, options = {}) {
        const child = new ComponentClass(elementId, options);
        this.children.push(child);
        return child;
    }

    /**
     * Lifecycle hook - called when component mounts
     */
    async onMount() {
        // Override in subclasses
    }

    /**
     * Lifecycle hook - called when component unmounts
     */
    onUnmount() {
        // Override in subclasses
    }

    /**
     * Lifecycle hook - called when data updates
     * @param {string} key - Data key that changed
     * @param {*} newValue - New value
     * @param {*} oldValue - Previous value
     */
    onDataUpdate(key, newValue, oldValue) {
        // Override in subclasses
    }

    /**
     * Debounce function calls
     * @param {function} func - Function to debounce
     * @param {number} delay - Delay in milliseconds
     * @returns {function} Debounced function
     */
    debounce(func, delay = 300) {
        let timeoutId;
        return (...args) => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => func.apply(this, args), delay);
        };
    }
}

// Export for use in modules
window.BaseComponent = BaseComponent;