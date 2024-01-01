import { YBFeedItem } from './YBFeedItem'

export class YBFeedError extends Error {
    status: number;
    constructor(status: number, message?:string) {
        super(message);
        this.status = status
    }
}

export interface YBFeed {
    name: string,
    secret: string,
    items: YBFeedItem[],
}