export * from './YBBreadCrumbComponent'
export * from './YBPasteCardComponent'
export * from './YBFeedItemsComponent'
export * from './YBFeedItemComponent'
export * from './YBFeedComponent'

export interface FeedItem {
    name: string,
    date: string,
    type: number,
    feed: Feed
}

export interface Feed {
    name: string,
    secret: string,
    items: FeedItem[],
}

export class FeedConnector {
    async GetFeed(feedName: string): Promise<Feed|null> {
        const f = await fetch("/api/feed/"+encodeURIComponent(feedName),{
            credentials: "include"
        })
        if (f.status !== 200) {
            return null
        }
        const j = await f.json()
        for (var i=0;i<j.items.length;i++) {
            j.items[i].feed = j
        }
        return j
    }
}