import { useEffect, useState, useMemo } from "react"

import { useParams, useSearchParams, useNavigate, useLocation } from 'react-router-dom';

import { notifications } from '@mantine/notifications';

import { Menu, ActionIcon, Center, Group, rem, Button, Box} from '@mantine/core';

import { YBFeedConnector } from '.'

import { YBBreadCrumbComponent, YBPasteCardComponent, YBFeedItemsComponent, YBNotificationToggleComponent } from './Components'
import { defaultNotificationProps } from './config';

import {
    IconLink, IconHash
  } from '@tabler/icons-react';
import { PinModal } from "./Components/PinModal";
import { PinRequest } from "./Components/PinRequest";

export function YBFeedFeed() {
    const { feedName } = useParams()
    const navigate = useNavigate()
    const location = useLocation()
    const [searchParams] = useSearchParams()
    // const navigate = useNavigate()
    const [secret,setSecret] = useState<string>("")

    // const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [vapid, setVapid] = useState<string|undefined>(undefined)


    const [fatal, setFatal] = useState(false)

    const connection = useMemo(() => new YBFeedConnector(),[])

    if (!feedName) {
        navigate("/")
        return
    }

    // If secret is sent as part of the URL params, set secret state and
    // redirect to the URL omitting the secret
    useEffect(() => {
        const s=searchParams.get("secret")
        if (s) {
            connection.AuthenticateFeed(feedName,s)
            .then((se) => {
                console.log(se)
                navigate("/" + feedName)
            })
            .catch((e) => console.log(e))
        }
    },[searchParams, feedName, connection, secret])

    // Get current feed over http without web-socket to fetch feed secret
    // As websocket doesn't send current cookie, we have to perform a regular
    // http request first to get the secret
    useEffect(() => {
        console.log(secret)
        if (!secret && !searchParams.get("secret")) {
            connection.GetFeed(feedName)
            .then((f) => {
                if (f && f.secret) {
                    setSecret(f.secret)
                    setVapid(f.vapidpublickey)
                    console.log("setting authenticated true")
                    setAuthenticated(true)
                }
            })
            .catch((e) => {
                console.log(e)
                if (e.status === 401) {
                    setAuthenticated(false)
                }
                else {
                    setFatal(e.message)
                }
            })
        }
    },[secret,connection,feedName,location])

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

    console.log("render YBFeedFeed")
    return (
        <Box>
        {authenticated===true&&
        <>
            <Group gap="xs" justify="flex-end" style={{float: 'right'}}>
                <Button size="xs" variant="outline" color="red" onClick={deleteAll}>Delete Content</Button>
                {vapid&&
                <YBNotificationToggleComponent vapid={vapid} feedName={feedName}/>
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

            <YBPasteCardComponent/>
            
            <YBFeedItemsComponent feedName={feedName} secret={secret}/>
            </>
            }
       </Box>
    )
}
