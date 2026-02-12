window.dashboardComponents['change-requests'] = function changeRequestsComponent() {
    return {
        loading: false,
        error: null,
        data: {
            data: [],
            count: 0,
            total_pages: 0,
        },
        filters: {
            status: 'pending',
            search: '',
        },
        pagination: {
            page: 1,
            limit: 20,
        },

        // review modal state
        selectedReq: null,
        diffLoading: false,
        diffError: null,
        diff: { diff: [] },

        rejectMode: false,
        reviewNote: '',

        actionLoading: false,
        actionType: null, // 'apply' | 'reject'

        get diffRows() {
            const fields = this.diff?.diff && Array.isArray(this.diff.diff) ? this.diff.diff : [];

            return fields.map((field) => ({
                key: field.key,
                type: field.type,
                hasDiff: !!(field.diff && field.diff !== ''),
                current: this.renderValue(field, 'current'),
                proposed: this.renderValue(field, 'proposed'),
            }));
        },

        renderValue(field, side) {
            // If field has diff property and it's not empty, show diff
            if (field.diff && field.diff !== '') {
                const html = (field.diff || '').replace(/\n/g, '<br/>');
                return { kind: 'diff', html: html };
            }

            const value = field[side];
            if (value === null || value === undefined || value === '') {
                return { kind: 'text', text: '' };
            }

            switch (field.type) {
                case 'image':
                    return { kind: 'image', src: `/files/${value}/?preview`, alt: 'preview' };
                case 'list':
                    const items = Array.isArray(value) ? value : [];
                    // Check if items are objects
                    const hasObjects = items.length > 0 && typeof items[0] === 'object' && items[0] !== null;

                    if (hasObjects) {
                        // Convert objects to key-value pairs
                        return {
                            kind: 'list-objects',
                            items: items.map(item => {
                                if (typeof item === 'object' && item !== null) {
                                    return Object.entries(item)
                                        .filter(([key, val]) => val !== null && val !== undefined && val !== '')
                                        .map(([key, val]) => `${key}: ${val}`)
                                        .join(', ');
                                }
                                return String(item);
                            })
                        };
                    }

                    return { kind: 'list', items: items };
                case 'text':
                default:
                    return { kind: 'text', text: String(value) };
            }
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
                    ...this.filters,
                });

                const response = await fetch(`/api/change-requests/?${params}`);

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                // reset container
                this.data = { data: [], count: 0, total_pages: 0 };

                const payload = await response.json();

                this.data.data = Array.isArray(payload.data) ? payload.data : [];
                this.data.count = payload.count ?? 0;
                this.data.total_pages = payload.total_pages ?? 0;
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

        resetReviewState() {
            this.selectedReq = null;
            this.diffLoading = false;
            this.diffError = null;
            this.diff = { diff: [] };
            this.rejectMode = false;
            this.reviewNote = '';
            this.actionLoading = false;
            this.actionType = null;
        },

        async openReview(req) {
            this.resetReviewState();
            this.selectedReq = req;

            // open modal
            this.$refs.reviewModal.showModal();

            // load diff
            await this.loadDiff(req.id);
        },

        async loadDiff(id) {
            this.diffLoading = true;
            this.diffError = null;
            this.diff = { diff: [] };

            try {
                const resp = await fetch(`/api/change-requests/${id}/diff/`);
                if (!resp.ok) throw new Error(`HTTP error! status: ${resp.status}`);

                const payload = await resp.json();
                this.diff.diff = Array.isArray(payload?.diff) ? payload.diff : [];
            } catch (e) {
                console.error('Error loading diff:', e);
                this.diffError = 'Failed to load diff. Please try again.';
            } finally {
                this.diffLoading = false;
            }
        },

        closeModal() {
            this.$refs.reviewModal.close();
        },

        removeSelectedRow() {
            const id = this.selectedReq?.id;
            if (!id) return;
            this.data.data = this.data.data.filter((x) => x.id !== id);
            this.data.count = Math.max(0, (this.data.count ?? 0) - 1);
        },

        async applySelected() {
            const id = this.selectedReq?.id;
            if (!id || this.actionLoading) return;

            this.actionLoading = true;
            this.actionType = 'apply';

            try {
                const resp = await fetch(`/api/change-requests/${id}/apply/`, { method: 'PUT' });
                if (resp.status !== 200) throw new Error(`HTTP error! status: ${resp.status}`);

                this.removeSelectedRow();
                this.closeModal();
            } catch (e) {
                console.error('Error applying changes:', e);
                this.diffError = 'Failed to apply changes. Please try again.';
            } finally {
                this.actionLoading = false;
                this.actionType = null;
            }
        },

        startReject() {
            this.rejectMode = true;
            this.reviewNote = '';
        },

        cancelReject() {
            this.rejectMode = false;
            this.reviewNote = '';
        },

        async submitReject() {
            const id = this.selectedReq?.id;
            if (!id || this.actionLoading) return;

            this.actionLoading = true;
            this.actionType = 'reject';

            try {
                const body = {};
                const note = (this.reviewNote ?? '').trim();
                if (note) body.note = note;

                const resp = await fetch(`/api/change-requests/${id}/reject/`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(body),
                });

                if (resp.status !== 200) throw new Error(`HTTP error! status: ${resp.status}`);

                this.removeSelectedRow();
                this.closeModal();
            } catch (e) {
                console.error('Error rejecting changes:', e);
                this.diffError = 'Failed to reject changes. Please try again.';
            } finally {
                this.actionLoading = false;
                this.actionType = null;
            }
        },
    };
};