import { createContext, useEffect, useRef, useState } from 'react'
import { Space } from "@mantine/core"
import { YBFeedItemComponent } from '.'
import { Connector, YBFeed, YBFeedItem } from '../'
import { useNavigate } from 'react-router-dom';

export const FeedItemContext = createContext<undefined|YBFeedItem>(undefined);

export interface YBFeedItemsComponentProps {
    feedName: string
    secret: string
    onDelete?: (item: YBFeedItem) => void
    setEmpty?: (arg0: boolean) => void
}

export function YBFeedItemsComponent(props: YBFeedItemsComponentProps) {
    const { feedName, secret } = props

    const navigate = useNavigate()
    const [feedItems, setFeedItems] = useState<YBFeedItem[]>([])
    // Setup websocket to receive feed events
    const ws = useRef<WebSocket|null>(null)

    // Do the actual item deletion callback
    const deleteItem = (item: YBFeedItem) => {
        Connector.DeleteItem(item)
    }

    const removeItem = (item: YBFeedItem) => {
        const newI = feedItems.filter((i) => i.name !== item.name)
        setFeedItems(newI)
        props.setEmpty && props.setEmpty(newI.length === 0)
        props.setEmpty && props.setEmpty(newI.length === 0)
    }

    const addItem = (item: YBFeedItem) => {
        setFeedItems((items) => [item].concat(items))
        props.setEmpty && props.setEmpty(false)
    }

    useEffect(() => {
        const webSocketURL = window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedName + "?secret=" + secret

        function disconnect() {
            if (ws.current === null) {
                return
            }
            ws.current.close()
            ws.current = null
        }

        function connect() {
            disconnect()
            ws.current = new WebSocket(webSocketURL)
            if (ws.current === null) {
                return
            }
            ws.current.onopen = () => {
                console.log("websocket connected")
                ws.current?.send("feed")
            }

            ws.current.onclose = (e) => {
                console.log("websocket closed : ",e)

                if (e.code > 4000) {
                    navigate("/")
                    return
                }
                // Try to reconnect
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

    useEffect(() => {
        if (!ws.current) {
            return
        }

        ws.current.onmessage = (m:WebSocketEventMap["message"]) => {
            const message_data = JSON.parse(m.data)
            if (message_data) {
                if (Object.prototype.hasOwnProperty.call(message_data, "items")) {
                    const f = (message_data as YBFeed)
                    setFeedItems(f.items)
                    props.setEmpty && props.setEmpty(f.items.length === 0)
                }
                if (Object.prototype.hasOwnProperty.call(message_data, "action")) {
                    interface ActionMessage {
                        action: string,
                        item: YBFeedItem
                    }
                    const am = (message_data as ActionMessage)
                    if (am.action === "remove") {
                        removeItem(am.item)
                    } else if (am.action === "add") {
                        addItem(am.item)
                    } else if (am.action === "empty") {
                        setFeedItems([])
                        props.setEmpty && props.setEmpty(true)
                    }
                }
            }
        }
    },[feedItems])

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