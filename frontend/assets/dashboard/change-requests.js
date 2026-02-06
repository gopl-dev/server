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
        diff: { current: {}, proposed: {} },

        rejectMode: false,
        reviewNote: '',

        actionLoading: false,
        actionType: null, // 'apply' | 'reject'

        get diffRows() {
            const cur = this.diff?.current && typeof this.diff.current === 'object' ? this.diff.current : {};
            const prop = this.diff?.proposed && typeof this.diff.proposed === 'object' ? this.diff.proposed : {};

            const keys = new Set([...Object.keys(cur), ...Object.keys(prop)]);
            const sorted = Array.from(keys).sort((a, b) => a.localeCompare(b));

            return sorted.map((key) => ({
                key,
                current: this.renderValueForKey(key, cur[key]),
                proposed: this.renderValueForKey(key, prop[key]),
            }));
        },

        formatCellValue(v) {
            if (v === null || v === undefined) return '';
            if (typeof v === 'string') return v;
            try {
                return JSON.stringify(v, null, 2);
            } catch (_) {
                return String(v);
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
            this.diff = { current: {}, proposed: {} };
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
            this.diff = { current: {}, proposed: {} };

            try {
                const resp = await fetch(`/api/change-requests/${id}/diff/`);
                if (!resp.ok) throw new Error(`HTTP error! status: ${resp.status}`);

                const payload = await resp.json();
                this.diff.current = payload?.current && typeof payload.current === 'object' ? payload.current : {};
                this.diff.proposed = payload?.proposed && typeof payload.proposed === 'object' ? payload.proposed : {};
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

        renderValueForKey(key, v) {
            if (v === null || v === undefined || v === '') return { kind: 'text', text: '' };

            if (key === 'cover_file_id') {
                return { kind: 'image', src: `/files/${v}/?preview`, alt: 'cover preview' };
            }

            // default
            return { kind: 'text', text: this.formatCellValue(v) };
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
