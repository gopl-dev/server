window.dashboardComponents['new-books'] = function newBooksComponent() {
    return {
        loading: false,
        error: null,
        data: {
            items: [],
            total: 0,
            total_pages: 0
        },
        filters: {
            s: 'review', // status
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

                const response = await fetch(`/api/books/?${params}`);

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                this.data = await response.json();

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

        async approveBook(id) {
            if (!confirm('Are you sure you want to approve this book?')) return;

            try {
                const response = await fetch(`/api/dashboard/books/${id}/approve`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });

                if (!response.ok) throw new Error('Failed to approve book');

                await this.loadData();
            } catch (error) {
                console.error('Error approving book:', error);
                alert('Failed to approve book');
            }
        },

        async rejectBook(id) {
            if (!confirm('Are you sure you want to reject this book?')) return;

            try {
                const response = await fetch(`/api/dashboard/books/${id}/reject`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });

                if (!response.ok) throw new Error('Failed to reject book');

                await this.loadData();
            } catch (error) {
                console.error('Error rejecting book:', error);
                alert('Failed to reject book');
            }
        }
    }
};