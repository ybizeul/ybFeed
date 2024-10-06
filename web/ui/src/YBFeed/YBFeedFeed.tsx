import { useEffect, useState, useMemo, useRef } from "react"

import { useParams, useSearchParams, Navigate, redirect } from 'react-router-dom';

import { notifications } from '@mantine/notifications';

import { Menu, ActionIcon, PinInput, Text, Modal, Center, Group, rem, Button, Box} from '@mantine/core';

import { YBFeed, YBFeedConnector, YBFeedItem } from '.'

import { YBBreadCrumbComponent, YBPasteCardComponent, YBFeedItemsComponent, YBNotificationToggleComponent, YBFeedItemComponent } from './Components'
import { defaultNotificationProps } from './config';

import {
    IconLink, IconHash
  } from '@tabler/icons-react';
import { PinModal } from "./Components/PinModal";
import { PinRequest } from "./Components/PinRequest";

export function YBFeedFeed() {
    const { feedName } = useParams()

    const [searchParams] = useSearchParams()
    // const navigate = useNavigate()

    // const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const [secret,setSecret] = useState<string>("")
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [vapid, setVapid] = useState<string|undefined>(undefined)

    const [feedItems, setFeedItems] = useState<YBFeedItem[]>([])

    const [fatal, setFatal] = useState(false)

    const connection = useMemo(() => new YBFeedConnector(),[])

    if (!feedName) {
        redirect("/")
        return
    }

    // If secret is sent as part of the URL params, set secret state and
    // redirect to the URL omitting the secret
    useEffect(() => {
        const s=searchParams.get("secret")
        if (s) {
            setSecret(s)
            connection.AuthenticateFeed(feedName,secret)
            .then(() => {
                redirect("/" + feedName)
            })
            .catch((e) => console.log(e))
        }
    },[searchParams, feedName, connection, secret])

    const ws = useRef<WebSocket|null>(null)

    useEffect(() => {
        if (!secret) {
            return
        }
        ws.current = new WebSocket(window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedName + "?secret=" + secret)
        if (ws.current === null) {
            return
        }
        ws.current.onopen = () => {
            ws.current?.send("feed")
        }

        ws.current.onmessage = (m:WebSocketEventMap["message"]) => {
            const message_data = JSON.parse(m.data)
            if (message_data) {
                if (Object.prototype.hasOwnProperty.call(message_data, "items")) {
                    const f = (message_data as YBFeed)
                    if (f.items) {
                        setFeedItems(f.items)
                        setAuthenticated(true)
                    }
                    if (f.secret) {
                        setSecret(f.secret)
                    }
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
        return () => {
            ws.current?.close()
        }
    },[secret])

    // const { sendMessage, readyState } = useMemo(() => useWebSocket(
    //     window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedParam,
    //     {queryParams:{"secret":secret},retryOnError: true, onMessage: handleMessage}, secret != ""),[]);

    // Do the actual item deletion callback
    const deleteItem = (item: YBFeedItem) => {
        setFeedItems((items) => items.filter((i) => i.name !== item.name))
        connection.DeleteItem(item)
    }

    

    // Get current feed over http without web-socket to fetch feed secret
    // As websocket doesn't send current cookie, we have to perform a regular
    // http request first to get the secret
    useEffect(() => {
        if (!secret) {
            connection.GetFeed(feedName)
            .then((f) => {
                if (f && f.secret) {
                    setSecret(f.secret)
                    setVapid(f.vapidpublickey)
                }
            })
            .catch((e) => {
                if (e.status === 401) {
                    setAuthenticated(false)
                }
                else {
                    setFatal(e.message)
                }
            })
        }
    },[secret,connection,feedName])

    // useEffect(() => {
    //     if (readyState === ReadyState.OPEN) {
    //         setFatal(false)
    //         sendMessage("feed")
    //     }
    //     else if (readyState === ReadyState.CLOSED){
    //         setFatal(true)
    //     }
    // },[readyState,sendMessage])

    //
    // Creating links to feed
    //
    const copyLink = () => {
        const link = window.location.href + "?secret=" + secret
        navigator.clipboard.writeText(link)
        notifications.show({
            message:'Link Copied!', ...defaultNotificationProps
        })
    }

    const setPIN = (pin: string) => {
        connection.SetPIN(feedName,pin)
        .then(() => {
            notifications.show({message:"PIN set", ...defaultNotificationProps})
            setPinModalOpen(false)
        })
        .catch((e) => {
            notifications.show({message:e.message, color:"red", ...defaultNotificationProps})
            setPinModalOpen(false)
        })
    }

    const sendPIN = (e: string) => {
        connection.AuthenticateFeed(feedName,e)
        .then((s) => {
            setSecret(s.toString())
        })
        .catch((e) => {
            notifications.show({message:e.message, color:"red", ...defaultNotificationProps})
        })
    }

    const deleteAll = () => {
        connection.EmptyFeed(feedName)
    }

    if (authenticated===false)  {
        return (
            <PinRequest sendPIN={sendPIN}/>
        )
    }

    if (fatal)  {
        return (
            <Center>{fatal}</Center>
        )
    }

    return (
        <Box>
        {authenticated===true&&
        <>
            <Group gap="xs" justify="flex-end" style={{float: 'right'}}>
                <Button size="xs" variant="outline" color="red" onClick={deleteAll}>Delete Content</Button>
                {vapid&&
                <YBNotificationToggleComponent vapid={vapid} feedName={feedName }/>
                }
                <Menu trigger="hover" position="bottom-end" withArrow arrowPosition="center">
                    <Menu.Target>
                        <ActionIcon size="md" variant="outline" aria-label="Menu" onClick={copyLink}>
                            <IconLink style={{ width: '70%', height: '70%' }} stroke={1.5} />
                        </ActionIcon>
                    </Menu.Target>
                    <Menu.Dropdown>
                    <Menu.Item leftSection={<IconLink style={{ width: rem(14), height: rem(14) }} />} onClick={copyLink}>
                        Copy Permalink
                    </Menu.Item>
                    <Menu.Item leftSection={<IconHash style={{ width: rem(14), height: rem(14) }} />} onClick={() => setPinModalOpen(true)}>
                        Set Temporary PIN
                    </Menu.Item>
                    </Menu.Dropdown>
                </Menu>
            </Group>

            <YBBreadCrumbComponent />
            <PinModal opened={pinModalOpen} setOpened={() => setPinModalOpen(false)} setPIN={setPIN}/>

            <YBPasteCardComponent empty={feedItems.length === 0}/>
            
            <YBFeedItemsComponent items={feedItems} onDelete={deleteItem}/>
            </>
            }
       </Box>
    )
}
