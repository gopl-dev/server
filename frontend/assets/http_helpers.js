(function (global) {
    async function requestJSON(url, { method = 'GET', body } = {}) {
        const opts = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        }

        if (body !== undefined) {
            opts.body = JSON.stringify(body)
        }

        const resp = await fetch(url, opts)

        let data = null
        try {
            data = await resp.json()
        } catch (_) {}

        return { resp, data }
    }

    function postJSON(url, body) {
        return requestJSON(url, { method: 'POST', body })
    }

    function putJSON(url, body) {
        return requestJSON(url, { method: 'PUT', body })
    }

    function deleteJSON(url) {
        return requestJSON(url, { method: 'DELETE' })
    }

    global.HTTP = {
        requestJSON,
        postJSON,
        putJSON,
        deleteJSON,
    }
})(window)