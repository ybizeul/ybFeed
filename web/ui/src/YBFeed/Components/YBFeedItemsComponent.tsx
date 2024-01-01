import { YBFeedItemComponent } from '.'
import { YBFeedItem } from '../'

export interface YBFeedItemsComponentProps {
    items: YBFeedItem[],
    onUpdate?: (item: YBFeedItem) => void
    onDelete?: () => void
}

export function YBFeedItemsComponent(props: YBFeedItemsComponentProps) {
    const { items, onUpdate } = props
    return(
        <>
        {items.map((f:YBFeedItem) => 
            <YBFeedItemComponent item={f} onUpdate={(f) => { if (onUpdate) { onUpdate(f)}}} onDelete={props.onDelete} key={f.name}/>
        )}
        </>
    )
}