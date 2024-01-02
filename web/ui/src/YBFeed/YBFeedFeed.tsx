import { useEffect, useState, useRef } from "react"

import { useParams, useSearchParams, Navigate } from 'react-router-dom';

import { notifications } from '@mantine/notifications';

import { Menu, ActionIcon, PinInput, Text, Modal, Center, Group, rem} from '@mantine/core';

import { YBFeedConnector, YBFeedItem } from '.'

import { YBBreadCrumbComponent, YBPasteCardComponent, YBFeedItemsComponent, YBNotificationToggleComponent } from './Components'
import { defaultNotificationProps } from './config';

import {
    IconLink, IconHash
  } from '@tabler/icons-react';

export function YBFeedFeed() {
    const feedParam: string = useParams().feed!
    const [searchParams] = useSearchParams()
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const feedItems = useRef<YBFeedItem[]>([])
    const [secret,setSecret] = useState<string|null>(null)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [updateGeneration,setUpdateGeneration] = useState(0)
    const [fatal, setFatal] = useState(null)
    const [vapid, setVapid] = useState<string|undefined>(undefined)

    const connection = new YBFeedConnector()
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

    //
    // Update feed is run every 2s or o, some events
    //
    function update() {
        if (feedParam === undefined) {
            return
        }
        connection.GetFeed(feedParam)
        .then((f) => {
            if (f === null) {
                return
            }
            setFatal(null)
            let do_update = false

            let found

            // Loop over current items and keep what is already here

            const oldItems = []
            for (let i=0;i<feedItems.current.length;i++) {
                const current_old_item = feedItems.current[i]
                found = false
                for (let j=0;j<f.items.length;j++) {
                    const current_new_item = f.items[j]
                    if (current_new_item.name === current_old_item.name) {
                        found = true
                        oldItems.push(feedItems.current[i])
                    }
                }
                if (found === false) {
                    do_update = true
                }
            }
            feedItems.current.length = 0
            feedItems.current = [...oldItems]

            // Loop over new items and add what is new
            for (let i=0;i<f.items.length;i++) {
                const current_new_item = f.items[i]
                found = false
                for (let j=0;j<feedItems.current.length;j++) {
                    const current_existing_item = feedItems.current[j]
                    if (current_existing_item.name === current_new_item.name) {
                        found = true
                    }
                }
                if (found === false) {
                    feedItems.current.push(f.items[i])
                    do_update=true
                }
            }

            feedItems.current.sort((a,b) =>{
                return (a.date < b.date)?1:-1
            })
            if (do_update === true) {
                setUpdateGeneration(updateGeneration+1)
            }

            setSecret(f.secret)
            setAuthenticated(true)
        })
        .catch(e => {
            if (e.status === 401) {
                setAuthenticated(false)
            } else if (e.status === 500) {
                setFatal(e.message)
            }
        })
    }

    useEffect(
        () => {
            // Update feed every 2s
            const interval = window.setInterval(update,2000)

            // Authenticate feed if a secret is found in URL
            const secret = searchParams.get("secret")
            if (secret) {
                connection.AuthenticateFeed(feedParam,secret)
                    .then(() => {
                        setGoTo("/" + feedParam)
                        update()
                    })
                    .catch((e) => {
                        notifications.show({
                            message:e.message,
                            color: "red",
                            ...defaultNotificationProps
                        })
                        setAuthenticated(false)
                    })
            }
            else {
                update()
            }

            // Set web notification public key
            fetch("/api/feed/"+encodeURIComponent(feedParam),{cache: "no-cache"})
                .then(r => {
                    if (r.status === 200) {
                        const v = r.headers.get("Ybfeed-Vapidpublickey")
                        if (v) {
                            setVapid(v)
                        }
                    }
                })
            return () => {
                window.clearInterval(interval)
            }
            // eslint-disable-next-line react-hooks/exhaustive-deps
        },[updateGeneration]
    )

    const handlePinModalCancel = () => {
        setPinModalOpen(false)
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
        .then(() => {
            update()
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
        {!fatal?
            <>
            {authenticated===true?
            <>
            <Modal title="Set Temporary PIN" className="PINModal" opened={pinModalOpen} onClose={handlePinModalCancel}>
                <div className="text-center">
                    Please choose a PIN, it will expire after 2 minutes:
                </div>
                <Center>
                <PinInput data-autofocus mt="1em" mb="1em" type="number" mask onComplete={(v) => { setPIN(v)}}/>
                </Center>
            </Modal>

            <div className="pasteCard">
                <YBPasteCardComponent empty={feedItems.current.length === 0} onPaste={update}/>
            </div>

            <YBFeedItemsComponent items={feedItems.current} onUpdate={update} onDelete={update}/>
            
            </>
            :""}

            {authenticated===false?
            <>
            <Text mt="2em" ta="center">This feed is protected by a PIN.</Text>
            <Center>
                <PinInput mt="2em" type="number" mask onComplete={(v) => { sendPIN(v)}}/>
            </Center>
            </>
            :""
            }
            </>
        :
            <p>{fatal}</p>
        
        }   
        </>
    )
}
