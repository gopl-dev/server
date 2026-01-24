(function (global) {
    function makeFileUpload({ purpose, onUploaded, onRemoved }) {
        return {
            uploading: false,
            removing: false,
            error: '',
            fileName: '',

            async upload(e) {
                this.error = ''
                const file = e?.target?.files?.[0]
                if (!file) return

                this.uploading = true
                this.fileName = file.name

                try {
                    const fd = new FormData()
                    fd.append('file', file)
                    fd.append('purpose', purpose)

                    const resp = await fetch('/api/files/', { method: 'POST', body: fd })
                    const data = await resp.json().catch(() => ({}))

                    if (resp.status !== 201) {
                        this.error = data?.error || 'Upload failed'
                        return
                    }

                    onUploaded?.(data.id, data)
                } catch (err) {
                    console.error('upload error:', err)
                    this.error = 'Upload failed'
                } finally {
                    this.uploading = false
                }
            },

            async remove(fileID) {
                this.error = ''
                if (!fileID) return

                this.removing = true
                try {
                    const resp = await fetch('/api/files/' + fileID + '/', {
                        method: 'DELETE',
                        headers: { 'Content-Type': 'application/json' },
                    })

                    if (resp.status !== 200 && resp.status !== 204) {
                        let msg = 'Remove failed'
                        try {
                            const data = await resp.json()
                            if (data?.error) msg = data.error
                        } catch (_) {}
                        this.error = msg
                        return
                    }

                    this.fileName = ''
                    onRemoved?.()
                } catch (_) {
                    this.error = 'Remove failed'
                } finally {
                    this.removing = false
                }
            },
        }
    }

    global.FileUpload = { makeFileUpload }
})(window)
