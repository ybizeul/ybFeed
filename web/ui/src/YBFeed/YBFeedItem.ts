import { YBFeed } from './YBFeed'

export interface YBFeedItem {
    name: string,
    date: string,
    type: number,
    feed: YBFeed
}