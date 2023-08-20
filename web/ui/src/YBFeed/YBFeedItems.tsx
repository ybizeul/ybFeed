import { YBFeedItem, FeedItem } from './YBFeedItem'

export interface FeedItemsProps {
    items: FeedItem[],
    onUpdate?: (item: FeedItem) => void
}

export function FeedItems(props: FeedItemsProps) {
    const { items, onUpdate } = props
    
    return(
        <>
        {items.map((f:FeedItem) => 
            <YBFeedItem item={f} onUpdate={onUpdate} key={f.name}/>
        )}
        </>
    )
}