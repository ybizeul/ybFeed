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
    feedUrl(feedName: string): string {
        return "/api/feed/"+encodeURIComponent(feedName)
    }
    async GetFeed(feedName: string): Promise<Feed|null> { 
        var f = await fetch(this.feedUrl(feedName),{
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
    async AuthenticateFeed(feedName: string, secret: string): Promise<boolean> {
        var f = await fetch(this.feedUrl(feedName)+"?secret="+encodeURIComponent(secret),{
            credentials: "include"
        })
        if (f.status !== 200) {
            throw new YBFeedError(f.status, f.statusText)
        }
        return true
    }
    async SetPIN(feedName: string, pin: string): Promise<boolean> {
        return fetch(this.feedUrl(feedName),{
            method: "PATCH",
            credentials: "include",
            body: pin
          })
          .then((f) => {
            if (f.status !== 200) {
                throw new YBFeedError(f.status, f.statusText)
            }
            return true
          })
          .catch((e) => {
            throw e
        })
    }
}