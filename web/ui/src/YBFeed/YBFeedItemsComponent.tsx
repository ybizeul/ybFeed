import { FeedItem, FeedItemProps, YBFeedItem } from './'

export interface FeedItemsProps {
    items: FeedItem[],
    onUpdate?: (item: FeedItemProps) => void
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