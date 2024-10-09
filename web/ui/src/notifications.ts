export function Subscribe(vapid: string) {
    return new Promise((resolve, reject) => {
        if (!vapid) {
            reject("VAPID not declared")
        }

        console.log("getting registration for", window.location.href)
        navigator.serviceWorker.getRegistration(window.location.href)
            .then((registration) => {  
                if (!registration) {
                    console.log("no registration")
                    return
                }
                console.log("subscribing",vapid)
                return registration.pushManager.subscribe({
                    userVisibleOnly: true,
                    applicationServerKey: urlBase64ToUint8Array(vapid),
                });
            })
            .then((subscription) => {
                console.log("got subscription", subscription)
                if (!subscription) {
                    reject("Unable to subscribe (empty subscription)")
                }
                resolve(subscription)
                console.log("adding subscription to backend", subscription)
            })
            .catch((err) => {
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