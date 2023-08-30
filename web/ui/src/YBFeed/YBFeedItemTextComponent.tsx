import { useEffect, useState } from 'react'
import { message } from 'antd'
import { Button } from 'antd'
import { MdContentCopy } from "react-icons/md";
import { FeedItemProps } from './YBFeedItemComponent'
import { FeedConnector } from '.';

export function FeedItemText(props:FeedItemProps) {
    const { item, showCopyButton } = props
    const [textValue,setTextValue] = useState("")

    const copyItem = () => {
        navigator.clipboard.writeText(textValue)
        message.info("Copied to clipboard!")
    }

    useEffect(() => {
        var connection = new FeedConnector()
        connection.GetItem(item)
        .then((text) => {
            setTextValue(text)
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