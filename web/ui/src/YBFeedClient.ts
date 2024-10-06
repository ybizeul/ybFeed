import { APIClient } from "./APIClient"

class YBFeedClient extends APIClient {
    constructor() {
        super('/api')
    }
}

export const Y = new YBFeedClient()