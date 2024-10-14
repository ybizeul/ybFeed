import { useState, useEffect } from 'react'

import { ActionIcon } from '@mantine/core';
import { IconBell } from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';

import { defaultNotificationProps } from '../config';

import { Subscribe, Subscribed, Unsubscribe } from '../../notifications';
import { Connector } from '../YBFeedConnector';

interface NotificationToggleProps {
    vapid: string
    feedName: string 
}

export function YBNotificationToggleComponent(props:NotificationToggleProps) {
    const {vapid, feedName} = props
    const [notificationsOn,setNotificationsOn] = useState(false)
    const [loading,setLoading] = useState(false)
    const [canPushNotifications, setCanPushNotification] = useState(false)

    useEffect(() => {
        if ("serviceWorker" in navigator) {
            navigator.serviceWorker.getRegistration()
            .then((registration) => {
                if (! registration) {
                    return
                }
                setCanPushNotification(registration.pushManager !== undefined)
            })
        }
    },[])

    useEffect(() => {
        Subscribed()
        .then((registration) => {
            setNotificationsOn(registration)
        })
    },[])

    const toggleNotifications = () => {
        navigator.serviceWorker.getRegistration()
        .then((registration) => {
            if (!registration) {
                return
            }
            return registration.pushManager.getSubscription()
        })
        .then((subscription) => {

            setLoading(true)

            if (subscription === null) {
                Subscribe(vapid)
                .then((subscription) => {
                    setNotificationsOn(true)
                    Connector.AddSubscription(feedName,JSON.stringify(subscription))
                    .then(() => {
                        setLoading(false)
                    })
                    return
                })
                .catch((e: Error) => {
                    setLoading(false)
                    console.log(e)
                    notifications.show({message:e.message, color:"red", ...defaultNotificationProps})
                })
                return
            } else {
                Unsubscribe()
                .then(() => {
                    Connector.RemoveSubscription(feedName,JSON.stringify(subscription))
                    .then(() => {
                        setLoading(false)
                        setNotificationsOn(false)
                    })
                })
                .catch(e => {
                    setLoading(false)
                    console.log(e)
                    notifications.show({message:"Unable to unsubscribe", color:"red", ...defaultNotificationProps})
                })
            }
        });
    }

    return(
        <>
        {canPushNotifications&&
        <ActionIcon
            size="md" 
            variant={notificationsOn?"filled":"outline"}
            aria-label="Settings"
            onClick={toggleNotifications}
            loading={loading}
            >
            <IconBell style={{ width: '70%', height: '70%' }} stroke={1.5} />
        </ActionIcon>}
        </>
    )
}