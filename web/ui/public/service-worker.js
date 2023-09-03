/*eslint no-restricted-globals: ["error"]*/
self.addEventListener('push', event => {
    console.log('[Service Worker] Push Received.');
    console.log(`[Service Worker] Push had this data: "${event.data.text()}"`);

    const title = 'YBFeed Notification';
    const options = {
      body: event.data.text(),
    };

    event.waitUntil(self.registration.showNotification(title, options));
});
