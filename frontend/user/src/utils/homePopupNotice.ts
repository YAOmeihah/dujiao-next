import { processHtmlForDisplay } from './content'

type LocalizedTextGetter = (value: unknown) => string

interface HomePopupNoticeLike {
    summary?: unknown
    content?: unknown
}

export function resolveHomePopupNoticeHtml(
    notice: HomePopupNoticeLike | null | undefined,
    getLocalizedText: LocalizedTextGetter,
): string {
    if (!notice) return ''

    const content = getLocalizedText(notice.content).trim()
    const summary = getLocalizedText(notice.summary).trim()
    return processHtmlForDisplay(content || summary)
}
