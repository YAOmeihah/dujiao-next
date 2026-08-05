import { describe, expect, it } from 'vitest'
import { resolveHomePopupNoticeHtml } from './homePopupNotice'

const getLocalizedText = (value: unknown) => {
    const localized = value as Record<string, string> | undefined
    return localized?.['zh-CN'] || ''
}

describe('resolveHomePopupNoticeHtml', () => {
    it('keeps rich notice content for the home popup', () => {
        const notice = {
            summary: { 'zh-CN': '纯文本摘要' },
            content: {
                'zh-CN': '<p>正文 <a href="https://example.com">链接</a></p><img src="/uploads/notice.png">',
            },
        }

        const html = resolveHomePopupNoticeHtml(notice, getLocalizedText)

        expect(html).toContain('<a href="https://example.com">链接</a>')
        expect(html).toContain('src="/uploads/notice.png"')
        expect(html).not.toContain('纯文本摘要')
    })

    it('falls back to summary when content is empty', () => {
        const notice = {
            summary: { 'zh-CN': '<p>摘要 <strong>重点</strong></p>' },
            content: { 'zh-CN': '' },
        }

        expect(resolveHomePopupNoticeHtml(notice, getLocalizedText)).toBe('<p>摘要 <strong>重点</strong></p>')
    })
})
