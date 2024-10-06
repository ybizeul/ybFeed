import { AxiosResponseHeaders } from 'axios'
import { YBFeed, YBFeedItem, YBFeedError } from '.'
import { Y } from '../YBFeedClient'

export class YBFeedConnector {
    feedUrl(feedName: string): string {
        return "/api/feeds/"+encodeURIComponent(feedName)
    }
    async Ping(): Promise<AxiosResponseHeaders> {
        return new Promise((resolve, reject) => {
            Y.request({
                url: "/api",
                method: 'GET'
            })
            // fetch("/api",{
            //     credentials: "include"
            // })
            .then((f) => {
                if (f) {
                    f.headers.then((h:AxiosResponseHeaders) => {
                        resolve(h)
                    })
                }
            })
            .catch((e)=>reject(e))
        })
    }
    async GetFeed(feedName: string): Promise<YBFeed|null> { 
        return new Promise((resolve, reject) => {
            Y.get('/feeds/' + encodeURIComponent(feedName))
            .then((f) => {
                resolve(f as YBFeed)
            })
            // .then((f) => {
            //     if (f) {
            //         const result = f as YBFeed
            //         // f.vapidpublickey = f.vapidpublickey
            //         // for (let i=0;i<f.items.length;i++) {
            //         //     result.items[i].feed = f
            //         // }
            //         resolve(result)
            //     }
            // })
            .catch((e) => {
                reject(new YBFeedError(e.status, "Server Unavailable"))
            })

            // fetch(this.feedUrl(feedName),{
            //         credentials: "include"
            // })
            // .then((f) => {
            //     if (f.ok) {
            //         return f.json()
            //     }
            //     else {
            //         reject(new YBFeedError(f.status, "Server Error: " + f.statusText))
            //     }
            // })
            // .then((f) => {
            //     const result = f
            //     result.vapidpublickey = f.vapidpublickey
            //     for (let i=0;i<f.items.length;i++) {
            //         result.items[i].feed = f
            //     }
            //     resolve(result)
            // })
            // .catch((e) => {
            //     reject(new YBFeedError(e.status, "Server Unavailable"))
            // })
        })
    }
    async AuthenticateFeed(feedName: string, secret: string): Promise<string|YBFeedError> {
        return new Promise((resolve, reject) => {
            Y.get('/feeds/' + encodeURIComponent(feedName) + "?secret=" + encodeURIComponent(secret))
            .then((f) => {
                const fe = f as YBFeed
                if (fe.secret) {
                    resolve(fe.secret)
                }
            })
            .catch((error) => {
                if (error.status === 401) {
                    reject(new YBFeedError(401, "Unauthorized"))
                } else {
                    reject(new YBFeedError(error.status, "Server Unavailable"))
                }
            })

            // fetch(this.feedUrl(feedName)+"?secret="+encodeURIComponent(secret),{
            //     credentials: "include"
            // })
            // .then(f => {
            //     if (f.status !== 200) {
            //         f.text()
            //         .then(text => {
            //             reject(new YBFeedError(f.status, text))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     } else {
            //         f.json()
            //         .then((j) => {
            //             resolve(j.secret)
            //         })
            //     }
            // })
        })
    }
    async GetItem(item: YBFeedItem): Promise<string> {
        return new Promise((resolve, reject) => {
            Y.get('/feeds/' + encodeURIComponent(item.feed.name) + "/items/" + item.name)
            .then((i) => {
                resolve(i as string)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while getting item"))
            })

            // fetch(this.feedUrl(item.feed.name)+"/items/"+item.name,{
            //     credentials: "include"
            // })
            // .then(r => {
            //     if (r.status !== 200) {
            //         reject(new YBFeedError(r.status, "Error while getting item"))
            //     }
            //     r.text()
            //     .then(t => {
            //         resolve(t)
            //     })
            //     .catch(e => {
            //         reject(new YBFeedError(e.status, "Error while getting item"))
            //     })
            // })
        })
    }
    async DeleteItem(item: YBFeedItem) {
        return new Promise((resolve, reject) => {
            Y.delete('/feeds/' + encodeURIComponent(item.feed.name) + "/items/" + encodeURIComponent(item.name))
            .then(() => {
                resolve(true)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while getting item"))
            })

            // fetch(this.feedUrl(item.feed.name)+"/items/"+encodeURIComponent(item.name),{
            //     method: "DELETE",
            //     credentials: "include"
            // })
            // .then((f) => {
            //     if (f.status !== 200) {
            //         f.text()
            //         .then(text => {
            //             reject(new YBFeedError(f.status, text))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     } else {
            //         resolve(true)
            //     }
            // })
        })
    }
    async EmptyFeed(feedName: string): Promise<boolean> {
        return new Promise((resolve, reject) => {
            Y.delete('/feeds/' + encodeURIComponent(feedName) + "/items")
            .then(() => {
                resolve(true)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while deleting item"))
            })

            // fetch(this.feedUrl(feedName)+
            //     "/items",{
            //     method: "DELETE",
            //     credentials: "include"
            // })
            // .then((f) => {
            //     if (f.status !== 200) {
            //         f.text()
            //         .then(text => {
            //             reject(new YBFeedError(f.status, text))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     } else {
            //         resolve(true)
            //     }
            // })
        })
    }

    async SetPIN(feedName: string, pin: string): Promise<boolean> {
        return new Promise((resolve, reject) => {
            Y.patch('/feeds/' + encodeURIComponent(feedName), pin)
            .then(() => {
                resolve(true)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while setting PIN"))
            })

            // fetch(this.feedUrl(feedName),{
            //     method: "PATCH",
            //     credentials: "include",
            //     body: pin
            // })
            // .then((f) => {
            //     if (f.status !== 200) {
            //         f.text().then((b) => {
            //             reject(new YBFeedError(f.status, b))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     }
            //     resolve(true)
            // })
            // .catch((e) => {
            //     reject(new YBFeedError(e.status, "Server Unavailable"))
            // })
        })
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    async AddSubscription(feedName: string, subscription: any): Promise<boolean> {
        return new Promise((resolve, reject) => {
            Y.post('/feeds/' + encodeURIComponent(feedName) + "/subscription", subscription)
            .then(() => {
                resolve(true)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while adding subscription"))
            })

            // fetch(this.feedUrl(feedName)+"/subscription",{
            //     method: "POST",
            //     credentials: "include",
            //     body: subscription
            // })
            // .then((f) => {
            //     if (f.status !== 200) {
            //         f.text().then((b) => {
            //             reject(new YBFeedError(f.status, b))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     }
            //     resolve(true)
            // })
            // .catch((e) => {
            //     reject(new YBFeedError(e.status, "Server Unavailable"))
            // })
        })
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    async RemoveSubscription(feedName: string, subscription: any): Promise<boolean> {
        return new Promise((resolve, reject) => {
            Y.delete('/feeds/' + encodeURIComponent(feedName) + "/subscription", subscription)
            .then(() => {
                resolve(true)
            })
            .catch((error) => {
                reject(new YBFeedError(error.status, "Error while adding subscription"))
            })
            
            // fetch(this.feedUrl(feedName)+"/subscription",{
            //     method: "DELETE",
            //     credentials: "include",
            //     body: subscription
            // })
            // .then((f) => {
            //     if (f.status !== 200) {
            //         f.text().then((b) => {
            //             reject(new YBFeedError(f.status, b))
            //         })
            //         .catch(() => {
            //             reject(new YBFeedError(f.status, "Server Unavailable"))
            //         })
            //     }
            //     resolve(true)
            // })
            // .catch((e) => {
            //     reject(new YBFeedError(e.status, "Server Unavailable"))
            // })
        })
    }

}
