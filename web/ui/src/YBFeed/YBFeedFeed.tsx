import { useEffect, useState, useMemo } from "react"

import { useParams, useSearchParams, Navigate } from 'react-router-dom';

import useWebSocket, {ReadyState} from 'react-use-websocket';

import { notifications } from '@mantine/notifications';

import { Menu, ActionIcon, PinInput, Text, Modal, Center, Group, rem} from '@mantine/core';

import { YBFeed, YBFeedConnector, YBFeedItem } from '.'

import { YBBreadCrumbComponent, YBPasteCardComponent, YBFeedItemsComponent, YBNotificationToggleComponent, YBFeedItemComponent } from './Components'
import { defaultNotificationProps } from './config';

import {
    IconLink, IconHash
  } from '@tabler/icons-react';

export function YBFeedFeed() {
    const feedParam: string = useParams().feed!

    const [searchParams] = useSearchParams()
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const [secret,setSecret] = useState<string>("")
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [vapid, setVapid] = useState<string|undefined>(undefined)

    const [feedItems, setFeedItems] = useState<YBFeedItem[]>([])

    const [fatal, setFatal] = useState(false)

    const connection = useMemo(() => new YBFeedConnector(),[])


    const { sendMessage, lastJsonMessage, readyState } = useWebSocket(window.location.protocol.replace("http","ws") + "//" + window.location.host + "/ws/" + feedParam,{queryParams:{"secret":secret},retryOnError: true}, secret != "");

    // If secret is sent as part of the URL params, set secret state and
    // redirect to the URL omitting the secret
    useEffect(() => {
        const s=searchParams.get("secret")
        if (s) {
            setSecret(s)
            connection.AuthenticateFeed(feedParam,secret)
            .then(() => {
                setGoTo("/" + feedParam)
            })
            .catch((e) => console.log(e))
        }
    },[searchParams, feedParam, connection, secret])

    // Get current feed over http without web-socket to fetch feed secret
    // As websocket doesn't send current cookie, we have to perform a regular
    // http request first to get the secret
    useEffect(() => {
        if (!secret) {
            connection.GetFeed(feedParam)
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
    },[secret,connection,feedParam])

    useEffect(() => {
        if (readyState === ReadyState.OPEN) {
            setFatal(false)
            sendMessage("feed")
        }
        else if (readyState === ReadyState.CLOSED){
            setFatal(true)
        }
    },[readyState,sendMessage])

    // Handles messages received by websocket
    useEffect(() => {
        if (lastJsonMessage) {
            if (Object.prototype.hasOwnProperty.call(lastJsonMessage, "items")) {
                const f = (lastJsonMessage as YBFeed)
                if (f.items) {
                    setFeedItems(f.items)
                    setAuthenticated(true)
                }
                if (f.secret) {
                    setSecret(f.secret)
                }
            }
            if (Object.prototype.hasOwnProperty.call(lastJsonMessage, "action")) {
                sendMessage("feed")
            }
        }

    },[lastJsonMessage,sendMessage])

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
        connection.SetPIN(feedParam,pin)
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
        connection.AuthenticateFeed(feedParam,e)
        .then((s) => {
            setSecret(s.toString())
        })
        .catch((e) => {
            notifications.show({message:e.message, color:"red", ...defaultNotificationProps})
        })
    }

    return (
        <>
        {goTo?
        <Navigate to={goTo} />
        :""}
        {authenticated===true?
            <Group gap="xs" justify="flex-end" style={{float: 'right'}}>
            {vapid?
            <YBNotificationToggleComponent vapid={vapid} feedName={feedParam}/>
            :""}
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
        :""}

        <YBBreadCrumbComponent />
        {(authenticated==undefined && !fatal)?<YBFeedItemComponent/>:""}
        {!fatal?
            <>
            {authenticated===true?
            <>
            <Modal title="Set Temporary PIN" className="PINModal" opened={pinModalOpen} onClose={() => setPinModalOpen(false)}>
                <div className="text-center">
                    Please choose a PIN, it will expire after 2 minutes:
                </div>
                <Center>
                <PinInput data-autofocus mt="1em" mb="1em" type="number" mask onComplete={(v) => { setPIN(v)}}/>
                </Center>
            </Modal>

            <YBPasteCardComponent empty={feedItems.length === 0}/>

            <YBFeedItemsComponent items={feedItems}/>
            </>
            :""}

            {authenticated===false?
            <>
            <Text mt="2em" ta="center">This feed is protected by a PIN.</Text>
            <Center>
                <PinInput mt="2em" type="number" mask onComplete={(v) => { sendPIN(v)}}/>
            </Center>
            </>
            :""}
            </>
        :
            <Center>{fatal}</Center>
        
        }
        </>
    )
}
