import { YBFeed, YBFeedItem, YBFeedError } from '.'

export class YBFeedConnector {
    feedUrl(feedName: string): string {
        return "/api/feed/"+encodeURIComponent(feedName)
    }
    async Ping(): Promise<Headers> {
        return new Promise((resolve, reject) => {
            fetch("/api",{
                credentials: "include"
            })
            .then((f) => {
                if (f) {
                    resolve(f.headers)
                }
            })
            .catch((e)=>reject(e))
        })
    }
    async GetFeed(feedName: string): Promise<YBFeed|null> { 
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName),{
                    credentials: "include"
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text()
                    .then(text => {
                        if (!text) {
                            text="Empty Response"
                        }
                        reject(new YBFeedError(f.status, text))
                    })
                }
                else {
                    f.json()
                    .then(j => {
                        j.vapidpublickey = f.headers.get("Ybfeed-Vapidpublickey")
                        for (let i=0;i<j.items.length;i++) {
                            j.items[i].feed = j
                        }
                        resolve(j)
                    })
                    .catch((e) => {
                        reject(new YBFeedError(e.status, "Server Error: " + e.message))
                    })
                }
            })
            .catch((e) => {
                reject(new YBFeedError(e.status, "Server Unavailable"))
            })
        })
    }
    async AuthenticateFeed(feedName: string, secret: string): Promise<string|YBFeedError> {
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
                    .catch(() => {
                        reject(new YBFeedError(f.status, "Server Unavailable"))
                    })
                } else {
                    f.json()
                    .then((j) => {
                        resolve(j.secret)
                    })
                }
            })
        })
    }
    async GetItem(item: YBFeedItem): Promise<string> {
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
    async DeleteItem(item: YBFeedItem) {
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
                    .catch(() => {
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
                    .catch(() => {
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
                    .catch(() => {
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    async RemoveSubscription(feedName: string, subscription: any): Promise<boolean> {
        return new Promise((resolve, reject) => {
            fetch(this.feedUrl(feedName)+"/subscription",{
                method: "DELETE",
                credentials: "include",
                body: subscription
            })
            .then((f) => {
                if (f.status !== 200) {
                    f.text().then((b) => {
                        reject(new YBFeedError(f.status, b))
                    })
                    .catch(() => {
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
