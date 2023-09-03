export * from './YBBreadCrumbComponent'
export * from './YBPasteCardComponent'
export * from './YBFeedItemsComponent'
export * from './YBFeedItemComponent'
export * from './YBFeedComponent'
export * from './YBNotificationToggle'

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
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName),{
                    credentials: "include"
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text()
                    .then(text => {
                        reject(new YBFeedError(f.status, text))
                    })
                }
                f.json()
                .then(j => {
                    for (var i=0;i<j.items.length;i++) {
                        j.items[i].feed = j
                    }
                    resolve(j)
                })
                .catch((e) => {
                    reject(new YBFeedError(f.status, "Server Unavailable"))
                })
            })
            .catch((e) => {
                reject(new YBFeedError(e.status, "Server Unavailable"))
            })
        })
    }
    async AuthenticateFeed(feedName: string, secret: string): Promise<boolean> {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName)+"?secret="+encodeURIComponent(secret),{
                credentials: "include"
            })
            .then(f => {
                if (f.status !== 200) {
                    f.text()
                    .then(text => {
                        reject(new YBFeedError(f.status, text))
                    })
                    .catch((e) => {
                        reject(new YBFeedError(f.status, "Server Unavailable"))
                    })
                } else {
                    resolve(true)
                }
            })
        })
    }
    async GetItem(item: FeedItem): Promise<string> {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(item.feed.name)+"/"+item.name,{
                credentials: "include"
            })
            .then(r => {
                if (r.status !== 200) {
                    reject(new YBFeedError(r.status, "Error while getting item"))
                }
                r.text()
                .then(t => {
                    resolve(t)
                })
                .catch(e => {
                    reject(new YBFeedError(e.status, "Error while getting item"))
                })
            })
        })
    }
    async DeleteItem(item: FeedItem) {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(item.feed.name)+"/"+encodeURIComponent(item.name),{
                method: "DELETE",
                credentials: "include"
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text()
                    .then(text => {
                        reject(new YBFeedError(f.status, text))
                    })
                    .catch((e) => {
                        reject(new YBFeedError(f.status, "Server Unavailable"))
                    })
                } else {
                    resolve(true)
                }
            })
        })
    }
    async SetPIN(feedName: string, pin: string): Promise<boolean> {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName),{
                method: "PATCH",
                credentials: "include",
                body: pin
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text().then((b) => {
                        reject(new YBFeedError(f.status, b))
                    })
                    .catch((e) => {
                        reject(new YBFeedError(f.status, "Server Unavailable"))
                    })
                }
                resolve(true)
            })
            .catch((e) => {
                reject(new YBFeedError(e.status, "Server Unavailable"))
            })
        })
    }
    async AddSubscription(feedName: string, subscription: any): Promise<boolean> {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName)+"/subscription",{
                method: "POST",
                credentials: "include",
                body: subscription
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text().then((b) => {
                        reject(new YBFeedError(f.status, b))
                    })
                    .catch((e) => {
                        reject(new YBFeedError(f.status, "Server Unavailable"))
                    })
                }
                resolve(true)
            })
            .catch((e) => {
                reject(new YBFeedError(e.status, "Server Unavailable"))
            })
        })
    }
}