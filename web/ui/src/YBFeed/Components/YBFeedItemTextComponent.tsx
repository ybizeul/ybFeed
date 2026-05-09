
import { useMemo } from 'react'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function YBFeedItemTextComponent(props: any) {
    const text = props.children as string | undefined

    const highlighted = useMemo(() => {
        if (!text) return null
        const result = hljs.highlightAuto(text)
        if (result.relevance > 5 && result.language) {
            return { html: result.value, language: result.language }
        }
        return null
    }, [text])

    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll", fontSize: "0.8em"}}>
                    {highlighted
                        ? <code className={`hljs language-${highlighted.language}`} dangerouslySetInnerHTML={{ __html: highlighted.html }} />
                        : <code>{text}</code>
                    }
                </pre>
            </div>
        </div>
    )
}