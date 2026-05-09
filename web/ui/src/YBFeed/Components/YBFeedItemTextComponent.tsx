
import { useEffect, useMemo } from 'react'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

export interface YBFeedItemTextComponentProps {
    children?: string
    highlight?: boolean
    language?: string | null
    onLanguageDetected?: (lang: string | null) => void
}

export function YBFeedItemTextComponent(props: YBFeedItemTextComponentProps) {
    const { children: text, highlight = true, language, onLanguageDetected } = props

    const highlighted = useMemo(() => {
        if (!text) return null
        if (language) {
            try {
                const result = hljs.highlight(text, { language })
                return { html: result.value, language }
            } catch {
                return null
            }
        }
        const result = hljs.highlightAuto(text)
        if (result.relevance > 5 && result.language) {
            return { html: result.value, language: result.language }
        }
        return null
    }, [text, language])

    useEffect(() => {
        if (!language) {
            onLanguageDetected?.(highlighted?.language ?? null)
        }
    }, [highlighted?.language, language, onLanguageDetected])

    const showHighlight = highlight && highlighted !== null

    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll", fontSize: "0.8em"}}>
                    {showHighlight
                        ? <code className={`hljs language-${highlighted!.language}`} dangerouslySetInnerHTML={{ __html: highlighted!.html }} />
                        : <code>{text}</code>
                    }
                </pre>
            </div>
        </div>
    )
}