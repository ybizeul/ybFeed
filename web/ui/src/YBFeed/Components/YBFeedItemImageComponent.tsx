import { Image, Card } from "@mantine/core"
import { YBFeedItemComponentProps } from './YBFeedItemComponent'

export function YBFeedItemImageComponent(props:YBFeedItemComponentProps) {
    const { item } = props

    return(
        <Card.Section mt="sm">
            <Image src={"/api/feed/"+encodeURIComponent(item.feed.name)+"/"+item.name} />
        </Card.Section>
    )
}