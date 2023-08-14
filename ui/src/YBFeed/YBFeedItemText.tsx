import { useEffect, useState } from 'react'
import { Button } from 'antd'
import { MdContentCopy } from "react-icons/md";
import { FeedItemProps } from './YBFeedItem'

export function FeedItemText(props:FeedItemProps) {
    const { item, showCopyButton } = props
    const { name, feed } = item
    const [textValue,setTextValue] = useState("")

    const copyItem = () => {
        navigator.clipboard.writeText(textValue)
    }

    useEffect(() => {
        fetch("/api/feed/"+feed+"/"+name,{
            credentials: "include"
            })
        .then(r => {
            r.text()
            .then(t => {
                setTextValue(t)
            })
        })
     
    })

    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll"}}>{textValue}</pre>
                {showCopyButton===undefined || showCopyButton === true?
                <Button icon={<MdContentCopy />} onClick={copyItem} />
                :""}
            </div>
        </div>
    )
}