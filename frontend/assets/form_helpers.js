(function (global) {
    function clone(obj) {
        if (typeof structuredClone === 'function') return structuredClone(obj)
        return JSON.parse(JSON.stringify(obj))
    }

    function emptyErrors(fields) {
        return Object.fromEntries(fields.map(k => [k, '']))
    }

    function applyInputErrors(targetErrors, inputErrors) {
        if (!inputErrors) return
        for (const [k, v] of Object.entries(inputErrors)) {
            if (k in targetErrors) targetErrors[k] = String(v ?? '')
        }
    }

    function makeForm({ defaults, submit }) {
        const fields = Object.keys(defaults)

        return {
            form: clone(defaults),
            errors: emptyErrors(fields),
            error: '',
            success: false,
            submitting: false,

            resetErrors() {
                this.error = ''
                this.errors = emptyErrors(fields)
            },

            async submitForm() {
                if (this.submitting) return

                this.submitting = true
                this.resetErrors()

                try {
                    await submit.call(this)
                } finally {
                    this.submitting = false
                }
            },
        }
    }

    global.FormHelpers = { clone, emptyErrors, applyInputErrors, makeForm }
})(window)