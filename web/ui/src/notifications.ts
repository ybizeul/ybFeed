if ('serviceWorker' in navigator) {
    console.log("registering service worker")
    navigator.serviceWorker.register('/service-worker.js')
        .then((registration) => {
            console.log("got registration", registration)
            if (! registration) {
                return
            }
            // setCanPushNotification(registration.pushManager !== undefined)
            // if (registration.scope === window.location.href) {
            //     if (registration.pushManager) {
            //         return registration.pushManager.getSubscription();
            //     }
            // }
        })
        // .then((subscription) => {
        //     console.log("got subscription", subscription)
        //     if (subscription) {
        //         setNotificationsOn(true)
        //     }
        // })
        .catch(error => {
            console.error('Service Worker registration failed:', error);
        })
}

export function Subscribe(vapid: string) {
    return new Promise((resolve, reject) => {
        if (!vapid) {
            reject("VAPID not declared")
        }

        console.log("getting registration for", window.location.href)
        navigator.serviceWorker.getRegistration()
            .then((registration) => {  
                if (!registration) {
                    console.log("no registration")
                    return
                }
                console.log("subscribing",vapid)
                registration.pushManager.subscribe({
                    userVisibleOnly: true,
                    applicationServerKey: urlBase64ToUint8Array(vapid),
                }).then(
                    (subscription) => {
                        console.log("got subscription", subscription)
                        if (!subscription) {
                            reject("Unable to subscribe (empty subscription)")
                        }
                        resolve(subscription)
                        console.log("adding subscription to backend", subscription)
                    })
            })
            .catch((err) => {
                reject(err)
            });
    })
}

export function Unsubscribe(): Promise<boolean> {
    return new Promise((resolve, reject) => {
        navigator.serviceWorker.ready
            .then((registration) => {  
                return registration.pushManager.getSubscription()
            })
            .then(function(subscription) {
                if (!subscription) {
                    reject(new Error("Unable to unsubscribe (empty subscription)"))
                    return
                }
                subscription.unsubscribe()
                resolve(true)
            })
            .catch(err => {
                console.error(err)
                reject(err)
            });
    })
}

export function Subscribed(): Promise<boolean> {
    return new Promise((resolve, reject) => {
        navigator.serviceWorker.ready
            .then((registration) => {  
                return registration.pushManager.getSubscription()
            })
            .then(function(subscription) {
                if (!subscription) {
                    resolve(false)
                    return
                }
                resolve(true)
            })
            .catch(err => {
                console.error(err)
                reject(err)
            });
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