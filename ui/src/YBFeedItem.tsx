import { useEffect, useState } from 'react'
import { Image,Space, Button } from 'antd'
import { message } from 'antd'
import { MdContentCopy } from "react-icons/md";

import {
    FileImageOutlined,
    FileTextOutlined,
    DeleteOutlined,
  } from '@ant-design/icons';


interface FeedItemProps {
    item: { [key: string]: any },
    feed: string,
    onDelete?: (item: string) => void
    onCopy?: (item: string) => void
}


export default function YBFeedItem(props:FeedItemProps) {
    const [textValue,setTextValue] = useState("")
    //
    // Copy item to pasteboard
    //
    const copyItem = async (item: string) => {
        if (props.item.type === 0) {
            navigator.clipboard.writeText(textValue)
            message.info("Copied to clipboard!")
        }
        else if (props.item.type === 1) {
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
                    console.log("Image is " + c.width + "x" + c.height)
                    c.toBlob(blob=>{
                        b(blob!)
                    },'image/png')
                }
                img.onload = imageLoaded

            })
            img.src = "/api/feed/"+props.feed+"/"+props.item.name

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
    }

    useEffect(() => {
        if (props.item.type === 0) {
            fetch("/api/feed/"+props.feed+"/"+props.item.name,{
                credentials: "include"
                })
            .then(r => {
                r.text()
                .then(t =>
                    setTextValue(t)
                )
            })
        }
    })

    return(
        <div className='item'>
            <YBHeading item={props.item} feed={props.feed} onDelete={props.onDelete} onCopy={copyItem}/>

            {(props.item.type === 0)?
            <div className="itemContainer">
                <div className="itemText">
                    <pre style={{overflowY:"scroll"}}>{textValue}</pre>
                    <Button icon={<MdContentCopy />} onClick={(e) => {copyItem(props.item.name)}} />
                </div>
            </div>
            :""
            }

            {(props.item.type === 1)?
             <div className="itemContainer">
                <div className="itemImg">
                <Image
                className="itemImage"
                src={"/api/feed/"+props.feed+"/"+props.item.name}
                preview={false}
                />
                <Button icon={<MdContentCopy />} onClick={(e) => {copyItem(props.item.name)}} />
                </div>
            </div>
                :""
                }
        </div>
    )
}
function YBHeading(props: FeedItemProps) {
    const deleteItem = () => {
        if (props.onDelete !== undefined) {
            props.onDelete(props.item.name)
        }
    }
    return (
        <div className='heading'>
        {(props.item.type === 0)?
        <FileTextOutlined />
        :""}
        {(props.item.type === 1)?
        <FileImageOutlined />
        :""}
        &nbsp;{props.item.name}
        <Space style={{float:'right'}}>
            <DeleteOutlined style={{fontSize: '14px', color: 'red'}} onClick={deleteItem} />
        </Space>
        </div>
    )
}

