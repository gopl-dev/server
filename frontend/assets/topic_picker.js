window.TopicPicker = {
    make(options) {
        const opts = { wrapRef: 'topicsWrap', ...options }

        return {
            sortedTopics() {
                const selected = new Set(this.form?.topics ?? [])
                return [...(this.topics ?? [])].sort((a, b) => {
                    const as = selected.has(a.id)
                    const bs = selected.has(b.id)
                    if (as !== bs) return as ? -1 : 1
                    return (a.name || '').localeCompare(b.name || '')
                })
            },

            toggleTopic(id) {
                const wrap = this.$refs?.[opts.wrapRef]
                const before = new Map()

                if (wrap) {
                    wrap.querySelectorAll('[data-topic-id]').forEach(el => {
                        before.set(el.dataset.topicId, el.getBoundingClientRect())
                    })
                }

                const cur = this.form.topics ?? []
                this.form.topics = cur.includes(id) ? cur.filter(v => v !== id) : [...cur, id]

                this.$nextTick(() => {
                    if (!wrap) return
                    wrap.querySelectorAll('[data-topic-id]').forEach(el => {
                        const first = before.get(el.dataset.topicId)
                        if (!first) return
                        const last = el.getBoundingClientRect()
                        const dx = first.left - last.left
                        const dy = first.top - last.top
                        if (dx === 0 && dy === 0) return

                        el.style.transform = `translate(${dx}px, ${dy}px)`
                        el.style.transition = 'transform 0s'
                        requestAnimationFrame(() => {
                            el.style.transition = 'transform 150ms ease-out'
                            el.style.transform = ''
                        })
                    })
                })
            },
        }
    },
}