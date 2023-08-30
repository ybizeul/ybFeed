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

interface YBFeedError {
    status: number
}
class YBFeedError extends Error {
    constructor(status: number, message?:string) {
        super(message);
        this.status = status
    }
}
export class FeedConnector {
    async GetFeed(feedName: string): Promise<Feed|null> {
        var f = await fetch("/api/feed/"+encodeURIComponent(feedName),{
                credentials: "include"
            })
        if (f.status !== 200) {
            throw new YBFeedError(f.status, f.statusText)
        }
        const j = await f.json()
        for (var i=0;i<j.items.length;i++) {
            j.items[i].feed = j
        }
        return j
    }
}