import { Image, Card } from "@mantine/core"

import { FeedItemContext } from "."
import { useContext } from "react"

export function YBFeedItemImageComponent() {
    const item = useContext(FeedItemContext)

    return(
        <Card.Section mt="sm">
            <Image src={"/api/feed/"+encodeURIComponent(item!.feed.name)+"/"+item!.name} />
        </Card.Section>
    )
}