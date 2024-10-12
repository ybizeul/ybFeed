import { useState, useEffect } from 'react'

import { ActionIcon } from '@mantine/core';
import { IconBell } from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';

import { YBFeedConnector } from '../'

import { defaultNotificationProps } from '../config';

import { Subscribe } from '../../notifications';

interface NotificationToggleProps {
    vapid: string
    feedName: string 
}

export function YBNotificationToggleComponent(props:NotificationToggleProps) {
    const {vapid, feedName} = props
    const [notificationsOn,setNotificationsOn] = useState(false)
    const [loading,setLoading] = useState(false)
    const [canPushNotifications, setCanPushNotification] = useState(false)

    // const subscribe = (): Promise<boolean> => {
    //         return new Promise((resolve, reject) => {
    //             if (!vapid) {
    //                 reject("VAPID not declared")
    //             }
    //             const connection = new YBFeedConnector()

    //             console.log("getting registration for", window.location.href)
    //             navigator.serviceWorker.getRegistration(window.location.href)
    //                 .then((registration) => {  
    //                     if (!registration) {
    //                         console.log("no registration")
    //                         return
    //                     }
    //                     console.log("subscribing",vapid)
    //                     return registration.pushManager.subscribe({
    //                         userVisibleOnly: true,
    //                         applicationServerKey: urlBase64ToUint8Array(vapid),
    //                     });
    //                 })
    //                 .then((subscription) => {
    //                     console.log("got subscription", subscription)
    //                     if (!subscription) {
    //                         reject("Unable to subscribe (empty subscription)")
    //                     }
    //                     console.log("adding subscription to backend", subscription)

    //                     connection.AddSubscription(feedName,JSON.stringify(subscription))
    //                         .then(() => {
    //                             console.log("subscription added")
    //                             resolve(true)
    //                         })
    //                 })
    //                 .catch((err) => {
    //                     setLoading(false)
    //                     notifications.show({title:"Error", message: err.message, color: "red", ...defaultNotificationProps})
    //                 });
    //         })
    // }
    
    async function unsubscribe(): Promise<boolean> {
        return new Promise((resolve, reject) => {
            if (!vapid) {
                reject("VAPID not declared")
            }
            const connection = new YBFeedConnector()

            navigator.serviceWorker.ready
                .then((registration) => {  
                    return registration.pushManager.getSubscription()
                })
                .then(function(subscription) {
                    if (!subscription) {
                        reject("Unable to unsubscribe (empty subscription)")
                        return
                    }
                    subscription.unsubscribe()
                    connection.RemoveSubscription(feedName,JSON.stringify(subscription))
                        .then(() => {
                            resolve(true)
                        })
                })
                .catch(err => console.error(err));
        })
    }

    const toggleNotifications = () => {
        console.log("toggle notification")
        if ('serviceWorker' in navigator) {
            console.log("get registration for", window.location.href)

            navigator.serviceWorker.getRegistration(window.location.href)
                .then((registration) => {
                    if (!registration) {
                        console.log("no registration")
                        return
                    }
                    return registration.pushManager.getSubscription()
                })
                .then((subscription) => {
                    console.log("got subscription", subscription)
                    if (subscription === null) {
                        setLoading(true)
                        console.log("subscribing")
                        Subscribe(vapid)
                            .then((subscription) => {
                                setLoading(false)
                                const s = subscription as PushSubscription
                                console.log("got subscription", subscription)
                                if (s.endpoint) {
                                    setNotificationsOn(true)
                                    const connection = new YBFeedConnector()
                                    connection.AddSubscription(feedName,JSON.stringify(subscription))
                                    .then(() => {
                                        console.log("subscription added")
                                    })
                                    return
                                }
                                throw new Error("empty endpoint")
                            })
                            .catch(e => {
                                setLoading(false)
                                console.log(e)
                                notifications.show({message:"Unable to subscribe", color:"red", ...defaultNotificationProps})
                            })
                        return
                    }

                    if (notificationsOn) {
                        console.log("unsubscribing")
                        unsubscribe()
                            .then(() => {
                                console.log("done")
                                setNotificationsOn(false)
                            })
                    }
                    else {
                        setLoading(true)
                        console.log("subscribing")
                        Subscribe(vapid)
                            .then((subscription) => {
                                setLoading(false)
                                const s = subscription as PushSubscription
                                console.log("got subscription", subscription)
                                if (s.endpoint) {
                                    setNotificationsOn(true)
                                    return
                                }
                                throw new Error("empty endpoint")
                            })
                            .catch(e => {
                                setLoading(false)
                                console.log(e)
                                notifications.show({message:"Unable to subscribe", color:"red", ...defaultNotificationProps})
                            })
                    }
                });
        }
    }

    useEffect(() => {
        if ('serviceWorker' in navigator && feedName) {
            console.log("registering service worker")
            navigator.serviceWorker.register('/service-worker.js',{scope: "/" + feedName})
                .then((registration) => {
                    console.log("got registration", registration)
                    if (! registration) {
                        return
                    }
                    setCanPushNotification(registration.pushManager !== undefined)
                    if (registration.scope === window.location.href) {
                        if (registration.pushManager) {
                            return registration.pushManager.getSubscription();
                        }
                    }
                })
                .then((subscription) => {
                    console.log("got subscription", subscription)
                    if (subscription) {
                        setNotificationsOn(true)
                    }
                })
                .catch(error => {
                    console.error('Service Worker registration failed:', error);
                })
        }
    })

    console.log("render YBNotificationToggleComponent")

    return(
        <>
        {canPushNotifications?
        <ActionIcon
            size="md" 
            variant={notificationsOn?"filled":"outline"}
            aria-label="Settings"
            onClick={toggleNotifications}
            loading={loading}
            >
            <IconBell style={{ width: '70%', height: '70%' }} stroke={1.5} />
        </ActionIcon>
        :""}
        </>
    )
}