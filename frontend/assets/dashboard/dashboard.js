// Component registry
window.dashboardComponents = {
    'home': null,
    'new-books': null,
    'book-edits': null,
    'page-edits': null,
    'log': null
};

window.formatDate = function (dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        hour12: false,
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
};

function dashboardApp() {
    return {
        loadingTemplate: false,
        templateError: null,
        currentRoute: 'home',
        currentTitle: 'Dashboard',
        templateContent: '',
        templateCache: {},
        componentCache: {},

        async init() {
            // Handle browser back/forward buttons
            window.addEventListener('popstate', (e) => {
                if (e.state && e.state.route) {
                    this.currentRoute = e.state.route;
                    this.currentTitle = e.state.title;
                    this.loadTemplate();
                }
            });

            // Load initial route from hash
            const hash = window.location.hash.slice(1);
            if (hash && window.dashboardComponents.hasOwnProperty(hash)) {
                await this.navigate(hash, false);
            } else {
                // Load home by default
                await this.navigate('home', false);
            }
        },

        async navigate(route, pushState = true) {
            this.currentRoute = route;
            this.currentTitle = this.getTitleForRoute(route);
            this.templateError = null;

            if (pushState) {
                const url = `/dashboard/#${route}`;
                window.history.pushState({ route, title: this.currentTitle }, '', url);
            }

            await this.loadTemplate();
        },

        async loadTemplate() {
            // Check cache first
            if (this.templateCache[this.currentRoute]) {
                this.templateContent = this.templateCache[this.currentRoute];
                return;
            }

            this.loadingTemplate = true;
            this.templateError = null;

            try {
                // Load component JS first
                await this.loadComponentScript(this.currentRoute);

                // Then load HTML template
                const response = await fetch(`/assets/dashboard/${this.currentRoute}.html`);

                if (!response.ok) {
                    throw new Error(`Failed to load template: ${response.status}`);
                }

                const html = await response.text();

                // Cache the template and component
                this.templateCache[this.currentRoute] = html;
                this.componentCache[this.currentRoute] = window.dashboardComponents[this.currentRoute];
                this.templateContent = html;

            } catch (error) {
                console.error('Error loading template:', error);
                this.templateError = 'Failed to load page template. Please try again.';
            } finally {
                this.loadingTemplate = false;
            }
        },

        async loadComponentScript(route) {
            // Check if component is already loaded
            if (window.dashboardComponents[route]) {
                return;
            }

            return new Promise((resolve, reject) => {
                const script = document.createElement('script');
                script.src = `/assets/dashboard/${route}.js`;
                script.onload = () => resolve();
                script.onerror = () => reject(new Error(`Failed to load component: ${route}`));
                document.head.appendChild(script);
            });
        },

        // Helper function
        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleDateString('en-US', {
                hour12: false,
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        },

        getTitleForRoute(route) {
            const titles = {
                'home': 'Dashboard',
                'new-books': 'New Books',
                'change-requests': 'Change requests',
            };
            return titles[route] || 'Dashboard';
        }
    }
}