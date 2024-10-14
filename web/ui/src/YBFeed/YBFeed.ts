import { YBFeedItem } from './YBFeedItem'

export class YBFeedError extends Error {
    status: number;
    constructor(status: number, message?:string) {
        super(message);
        this.status = status
    }
}

export class YBFeed {
    name: string;
    secret: string|undefined;
    items: YBFeedItem[];
    vapidpublickey: string|undefined;
    constructor(name: string) {
        this.name = name
        this.items = []
    }
    webSocketUrl(): string {
        const path = window.location.origin + "/ws/" + self.name
        return path
    }
}