import { createContext, useEffect, useRef, useState } from 'react'
import { Space } from "@mantine/core"
import { YBFeedItemComponent } from '.'
import { Connector, YBFeed, YBFeedItem } from '../'

export const FeedItemContext = createContext<undefined|YBFeedItem>(undefined);

export interface YBFeedItemsComponentProps {
    feedName: string
    secret: string
    onDelete?: (item: YBFeedItem) => void
}

export function YBFeedItemsComponent(props: YBFeedItemsComponentProps) {
    const { feedName, secret } = props

    const [feedItems, setFeedItems] = useState<YBFeedItem[]>([])

    // Setup websocket to receive feed events
    const ws = useRef<WebSocket|null>(null)

    // Do the actual item deletion callback
    const deleteItem = (item: YBFeedItem) => {
        setFeedItems((items) => items.filter((i) => i.name !== item.name))
        Connector.DeleteItem(item)
    }

    useEffect(() => {
        const webSocketURL = window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedName + "?secret=" + secret

        function connect() {
            ws.current = new WebSocket(webSocketURL)
            if (ws.current === null) {
                return
            }
            ws.current.onopen = () => {
                console.log("websocket connected")
                ws.current?.send("feed")
            }

            ws.current.onmessage = (m:WebSocketEventMap["message"]) => {
                const message_data = JSON.parse(m.data)
                if (message_data) {
                    if (Object.prototype.hasOwnProperty.call(message_data, "items")) {
                        const f = (message_data as YBFeed)
                        setFeedItems(f.items)
    
                    }
                    if (Object.prototype.hasOwnProperty.call(message_data, "action")) {
                        interface ActionMessage {
                            action: string,
                            item: YBFeedItem
                        }
                        const am = (message_data as ActionMessage)
                        if (am.action === "remove") {
                            setFeedItems((items) => items.filter((i) => i.name !== am.item.name))
                        } else if (am.action === "add") {
                            setFeedItems((items) => [am.item].concat(items))
                        } else if (am.action === "empty") {
                            setFeedItems([])
                        }
                    }
                }
            }

            ws.current.onclose = () => {
                console.log("websocket closed")

                // Try to reconnect
                ws.current = null
                setTimeout(() => {
                    console.log("reconnecting")
                    connect()
                },1000)
            }
        }

        connect()

        return () => {
            const w=ws.current
            if (!w) {
                console.log("no websocket to close")
                return
            }
            console.log("closing websocket")
            w.onclose = null
            w.close()
        }
    },[])

    return(
        <>
        {feedItems.map((f:YBFeedItem) =>
        <FeedItemContext.Provider value={f} key={f.name}>
            <YBFeedItemComponent onDelete={deleteItem} />
        </FeedItemContext.Provider>
        )}
        <Space h="md" />
        </>
    )
}