import { useState, useEffect } from 'react'

import { Button, message } from 'antd'

import {
    BellOutlined
  } from '@ant-design/icons';

import { FeedConnector } from '.'

interface NotificationToggleProps {
    vapid: string
    feedName: string 
}

export function NotificationToggle(props:NotificationToggleProps) {
    const {vapid, feedName} = props
    const [notificationsOn,setNotificationsOn] = useState(false)
    const [loading,setLoading] = useState(false)
    const [canPushNotifications, setCanPushNotification] = useState(false)

    async function subscribe(): Promise<Boolean> {
        return new Promise((resolve, reject) => {
            if (!vapid) {
                reject("VAPID not declared")
            }
            const connection = new FeedConnector()

            navigator.serviceWorker.getRegistration(window.location.href)
                .then(function(registration) {  
                    if (!registration) {
                        return
                    }
                    return registration.pushManager.subscribe({
                        userVisibleOnly: true,
                        applicationServerKey: urlBase64ToUint8Array(vapid),
                    });
                })
                .then(function(subscription) {
                    if (subscription) {
                        connection.AddSubscription(feedName,JSON.stringify(subscription))
                            .then((r) => {
                                resolve(true)
                            })
                    }
                    else {
                        reject("Unable to subscribe (empty subscription)")
                    }
                })
                .catch(err => console.error(err));
        })
    }
    // async function unsubscribe(): Promise<Boolean> {
    //     return new Promise((resolve, reject) => {
    //         if (!vapid) {
    //             reject("VAPID not declared")
    //         }
    //         const connection = new FeedConnector()

    //         navigator.serviceWorker.ready
    //             .then(function(registration) {
    //                 registration.pushManager.getSubscription()
    //                     .then(function(subscription) {
    //                         if (subscription) {
    //                             subscription.unsubscribe()
    //                                 .then(b => {
    //                                     connection.RemoveSubscription(feedName,JSON.stringify(subscription))
    //                                         .then((r) => {
    //                                             registration.unregister()
    //                                                 .then(b => {
    //                                                     resolve(true)
    //                                                 })
    //                                         })
    //                                 })
    //                         }
    //                         else {
    //                             reject("Unable to unsubscribe (empty subscription)")
    //                         }
    //                     })
    //             })
    //             .catch(err => console.error(err));
    //     })
    // }
    async function unsubscribe(): Promise<Boolean> {
        return new Promise((resolve, reject) => {
            if (!vapid) {
                reject("VAPID not declared")
            }
            const connection = new FeedConnector()

            navigator.serviceWorker.ready
                .then(function(registration) {  
                    return registration.pushManager.getSubscription()
                })
                .then(function(subscription) {
                    if (subscription) {
                        subscription.unsubscribe()
                        connection.RemoveSubscription(feedName,JSON.stringify(subscription))
                            .then((r) => {
                                resolve(true)
                            })
                    }
                    else {
                        reject("Unable to unsubscribe (empty subscription)")
                    }
                })
                .catch(err => console.error(err));
        })
    }
    function urlBase64ToUint8Array(base64String: string) {
        const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
        const base64 = (base64String + padding)
            .replace(/-/g, '+')
            .replace(/_/g, '/');
        const rawData = window.atob(base64);
        return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
    }

    const toggleNotifications = (e: any) => {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.getRegistration(window.location.href)
                .then(function(registration) {
                    if (!registration) {
                        return
                    }
                    return registration.pushManager.getSubscription()
                })
                .then(function(subscription) {
                    if (subscription) {
                        if (notificationsOn) {
                            unsubscribe()
                                .then(() => {
                                    setNotificationsOn(false)
                                })
                        }
                        else {
                            setLoading(true)
                            subscribe()
                                .then((b) => {
                                    setLoading(false)
                                    if (b) {
                                        setNotificationsOn(true)
                                    }
                                })
                                .catch(e => {
                                    setLoading(false)
                                    console.log(e)
                                    message.error("Error")
                                })
                        }
                    }
                    else {
                        setLoading(true)
                        subscribe()
                            .then((b) => {
                                setLoading(false)
                                if (b) {
                                    setNotificationsOn(true)
                                }
                            })
                            .catch(e => {
                                setLoading(false)
                                console.log(e)
                                message.error("Error")
                            })
                    }
                });
        }
    }

    useEffect(() => {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.register('service-worker.js',{scope: "/" + feedName});
            navigator.serviceWorker.getRegistration(window.location.href)
                .then(function(registration) {
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
                .then(function(subscription) {
                    if (subscription) {
                        setNotificationsOn(true)
                    }
                })
        }
    })

    return(
        <>
        {canPushNotifications?
        <Button 
            type={notificationsOn?"primary":"default"}
            loading={loading}
            icon={<BellOutlined/>}
            onClick={toggleNotifications}
        />
        :""}
        </>
    )
}