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
    const { feedName, secret, setEmpty } = props

    const navigate = useNavigate()
    const [feedItems, setFeedItems] = useState<YBFeedItem[]>([])
    const ws = useRef<WebSocket|null>(null)
    const pollingInterval = useRef<number|undefined>(undefined)
    const wsRecoveryInterval = useRef<number|undefined>(undefined)
    const unmounted = useRef(false)

    // Do the actual item deletion callback
    const deleteItem = (item: YBFeedItem) => {
        Connector.DeleteItem(item)
    }

    const removeItem = (item: YBFeedItem) => {
        setFeedItems((items) => {
            const newItems = items.filter((i) => i.name !== item.name)
            setEmpty && setEmpty(newItems.length === 0)
            return newItems
        })
    }

    const addItem = (item: YBFeedItem) => {
        setFeedItems((items) => {
            const withoutItem = items.filter((i) => i.name !== item.name)
            return [item].concat(withoutItem)
        })
        setEmpty && setEmpty(false)
    }

    useEffect(() => {
        unmounted.current = false

        if (!secret) {
            return
        }

        const webSocketURL = window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedName + "?secret=" + secret
        const pollingDelayMs = 5000
        const wsRecoveryDelayMs = 30000

        const clearPolling = () => {
            if (pollingInterval.current !== undefined) {
                window.clearInterval(pollingInterval.current)
                pollingInterval.current = undefined
            }
        }

        const clearWsRecovery = () => {
            if (wsRecoveryInterval.current !== undefined) {
                window.clearInterval(wsRecoveryInterval.current)
                wsRecoveryInterval.current = undefined
            }
        }

        const applyFeedSnapshot = (feed: YBFeed) => {
            setFeedItems(feed.items)
            setEmpty && setEmpty(feed.items.length === 0)
        }

        const startPolling = () => {
            if (pollingInterval.current !== undefined || unmounted.current) {
                return
            }

            const poll = () => {
                Connector.GetFeed(feedName)
                .then((f) => {
                    if (!f || unmounted.current) {
                        return
                    }
                    applyFeedSnapshot(f)
                })
                .catch(() => {
                    // Keep polling until websocket recovers.
                })
            }

            poll()
            pollingInterval.current = window.setInterval(poll, pollingDelayMs)
        }

        const startWsRecovery = (connect: () => void) => {
            if (wsRecoveryInterval.current !== undefined || unmounted.current) {
                return
            }

            wsRecoveryInterval.current = window.setInterval(() => {
                connect()
            }, wsRecoveryDelayMs)
        }

        function disconnect() {
            if (ws.current === null) {
                return
            }
            ws.current.onclose = null
            ws.current.onmessage = null
            ws.current.onerror = null
            ws.current.close()
            ws.current = null
        }

        function connect() {
            if (unmounted.current) {
                return
            }

            if (ws.current && (ws.current.readyState === WebSocket.OPEN || ws.current.readyState === WebSocket.CONNECTING)) {
                return
            }

            disconnect()

            const socket = new WebSocket(webSocketURL)
            ws.current = socket

            socket.onopen = () => {
                if (ws.current !== socket || unmounted.current) {
                    return
                }
                clearPolling()
                clearWsRecovery()
                socket.send("feed")
            }

            socket.onmessage = (m:WebSocketEventMap["message"]) => {
                if (ws.current !== socket || unmounted.current) {
                    return
                }

                let messageData: unknown
                try {
                    messageData = JSON.parse(m.data)
                } catch {
                    return
                }

                if (!messageData) {
                    return
                }

                if (Object.prototype.hasOwnProperty.call(messageData, "items")) {
                    applyFeedSnapshot(messageData as YBFeed)
                    return
                }

                if (Object.prototype.hasOwnProperty.call(messageData, "action")) {
                    interface ActionMessage {
                        action: string,
                        item?: YBFeedItem
                    }
                    const actionMessage = (messageData as ActionMessage)
                    if (actionMessage.action === "remove" && actionMessage.item) {
                        removeItem(actionMessage.item)
                    } else if (actionMessage.action === "add" && actionMessage.item) {
                        addItem(actionMessage.item)
                    } else if (actionMessage.action === "empty") {
                        setFeedItems([])
                        setEmpty && setEmpty(true)
                    }
                }
            }

            socket.onerror = () => {
                if (ws.current !== socket || unmounted.current) {
                    return
                }

                startPolling()
                startWsRecovery(connect)
            }

            socket.onclose = (e) => {
                if (ws.current === socket) {
                    ws.current = null
                }

                if (unmounted.current) {
                    return
                }

                if (e.code > 4000) {
                    navigate("/")
                    return
                }

                startPolling()
                startWsRecovery(connect)
            }

            if (ws.current === null) {
                return
            }
        }

        connect()

        return () => {
            unmounted.current = true
            clearPolling()
            clearWsRecovery()
            disconnect()
        }
    },[feedName, navigate, secret, setEmpty])

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