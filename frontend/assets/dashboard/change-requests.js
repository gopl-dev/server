window.dashboardComponents['change-requests'] = function changeRequestsComponent() {
    return {
        loading: false,
        error: null,
        data: {
            data: [],
            total: 0,
            total_pages: 0
        },
        filters: {
            status: 'pending',
            search: ''
        },
        pagination: {
            page: 1,
            limit: 20
        },

        async init() {
            await this.loadData();
        },

        async loadData() {
            this.loading = true;
            this.error = null;

            try {
                const params = new URLSearchParams({
                    page: this.pagination.page,
                    limit: this.pagination.limit,
                    ...this.filters
                });

                const response = await fetch(`/api/change-requests/?${params}`);

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                this.data = { data: [], count: 0 };

                const payload = await response.json();

                this.data.data = Array.isArray(payload.data) ? payload.data : [];
                this.data.count = payload.count ?? 0;

            } catch (error) {
                console.error('Error loading data:', error);
                this.error = 'Failed to load data. Please try again.';
            } finally {
                this.loading = false;
            }
        },

        async changePage(page) {
            if (page < 1 || page > this.data.total_pages) return;
            this.pagination.page = page;
            await this.loadData();
        },
    }
};