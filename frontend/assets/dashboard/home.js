window.dashboardComponents['home'] = function homeComponent() {
    return {
        loading: false,
        stats: {
            newBooks: 0,
            bookEdits: 0,
            pageEdits: 0,
            totalActions: 0
        },

        async init() {
            await this.loadStats();
        },

        async loadStats() {
            this.loading = true;

            try {
                const response = await fetch('/api/dashboard/stats');
                if (!response.ok) throw new Error('Failed to load stats');
                this.stats = await response.json();
            } catch (error) {
                console.error('Error loading stats:', error);
            } finally {
                this.loading = false;
            }
        }
    }
};