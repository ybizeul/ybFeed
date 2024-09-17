import { YBFeedItem } from '../'

export const copyImageItem = (item:YBFeedItem) => {
    return new Promise((resolve,reject) => {
        const img = document.createElement('img')
        const c = document.createElement('canvas')
        const ctx = c.getContext('2d')

        const imageDataPromise = new Promise<Blob>(resolve => {
            const b = (blob: Blob) => {
                resolve(blob)
            }
            const imageLoaded = () => {
                c.width = img.naturalWidth
                c.height = img.naturalHeight
                ctx?.drawImage(img,0,0)
                c.toBlob(blob=>{
                    b(blob!)
                },'image/png')
            }
            img.onload = imageLoaded

        })
        img.src = "/api/feeds/"+encodeURIComponent(item.feed.name)+"/items/"+item.name

        const mime = 'image/png'
        navigator.clipboard.write([new ClipboardItem({[mime]:imageDataPromise})])
        .then(() => {
            resolve(true)
        })
        .catch((e) => {
            reject(e)
        })
    })
}