import { createContext } from 'react'
import { Space } from "@mantine/core"
import { YBFeedItemComponent } from '.'
import { YBFeedItem } from '../'

export const FeedItemContext = createContext<undefined|YBFeedItem>(undefined);

export interface YBFeedItemsComponentProps {
    items: YBFeedItem[],
    onUpdate?: (item: YBFeedItem) => void
    onDelete?: (item: YBFeedItem) => void
}

export function YBFeedItemsComponent(props: YBFeedItemsComponentProps) {
    const { items, onUpdate } = props
    return(
        <>
        {items.map((f:YBFeedItem) =>
        <FeedItemContext.Provider value={f} key={f.name}>
            <YBFeedItemComponent onUpdate={(f) => { if (onUpdate) { onUpdate(f)}}} onDelete={props.onDelete} />
        </FeedItemContext.Provider>
        )}
        <Space h="md" />
        </>
    )
}