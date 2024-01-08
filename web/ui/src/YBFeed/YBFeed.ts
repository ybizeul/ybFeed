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
    items: YBFeedItem[]|undefined;
    constructor(name: string) {
        this.name = name
        this.items = undefined
    }
    webSocketUrl(): string {
        const path = window.location.origin + "/ws/" + self.name
        return path
    }
}