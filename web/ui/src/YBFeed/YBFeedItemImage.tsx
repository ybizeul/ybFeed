import { Button, Image, message } from 'antd'
import { MdContentCopy } from "react-icons/md";
import { FeedItemProps } from './YBFeedItem'

export function FeedItemImage(props:FeedItemProps) {
    const { item, showCopyButton } = props
    const {name, feed} = item

    const copyItem = () => {
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
        img.src = "/api/feed/"+encodeURIComponent(feed)+"/"+name

        let mime = 'image/png'
        navigator.clipboard.write([new ClipboardItem({[mime]:imageDataPromise})])
        .then(() => {
            message.info("Copied to clipboard!")
        })
        .catch((e) => {
            console.log(e)
            message.error("Unable to copy")
        }) 
    }

    return(
        <div className="itemContainer">
            <div className="itemImg">
                <Image
                src={"/api/feed/"+encodeURIComponent(feed)+"/"+name}
                preview={false}
                />
                {showCopyButton?
                <Button icon={<MdContentCopy />} onClick={copyItem} />
                :""}
            </div>
        </div>
    )
}