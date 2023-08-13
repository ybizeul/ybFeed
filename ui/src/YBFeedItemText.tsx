import { MouseEventHandler, useEffect, useState } from 'react'
import { Button, message } from 'antd'
import { MdContentCopy } from "react-icons/md";
import { FeedItemProps } from './YBFeedItem'

export function FeedItemText(props:FeedItemProps) {
    const { item } = props
    const { name, feed } = item
    const [textValue,setTextValue] = useState("")

    const copyItem = () => {
        console.log("copy item")
        navigator.clipboard.writeText(textValue)
        message.info("Copied to clipboard!")
    }

    useEffect(() => {
        fetch("/api/feed/"+feed+"/"+name,{
            credentials: "include"
            })
        .then(r => {
            r.text()
            .then(t =>
                setTextValue(t)
            )
        })
    })

    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll"}}>{textValue}</pre>
                <Button icon={<MdContentCopy />} onClick={copyItem} />
            </div>
        </div>
    )
}