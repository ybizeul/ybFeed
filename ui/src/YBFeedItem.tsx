import { useEffect, useState } from 'react'
import { Image,Space } from 'antd'
import { message } from 'antd'

import {
    FileImageOutlined,
    FileTextOutlined,
    DeleteOutlined,
    CopyOutlined
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
    const copyItem = (item: string) => {
        if (props.item.type === 0) {
            navigator.clipboard.writeText(textValue)
            message.info("Copied to clipboard!")
        }
        else if (props.item.type === 1) {
            const fetchPromise = async () => {
                const r = await fetch("/api/feed/"+props.feed+"/"+props.item.name,{
                                    credentials: "include"
                                })
                return await r.blob()
            }
            navigator.clipboard.write([new ClipboardItem({'image/png': fetchPromise()})])
            .then(() => {
                message.info("Copied to clipboard!")
            })
            .catch(() => {
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
            <div className="itemText">
                <pre style={{overflowY:"scroll"}}>{textValue}</pre>
            </div>
            :""
            }

            {(props.item.type === 1)?
                <div className='center'>
                <Image
                    className="itemImg"
                    src={"/api/feed/"+props.feed+"/"+props.item.name}
                    preview={false}
                />
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
    const copyItem = () => {
        if (props.onCopy !== undefined) {
            props.onCopy(props.item.name)
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
            <CopyOutlined style={{fontSize: '14px'}} onClick={copyItem} />
            <DeleteOutlined style={{fontSize: '14px'}} onClick={deleteItem} />
        </Space>
        </div>
    )
}

