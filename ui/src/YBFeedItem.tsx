import { useEffect, useState } from 'react'
import { Image } from 'antd'

import {
    FileImageOutlined,
    FileTextOutlined,
    DeleteOutlined,
  } from '@ant-design/icons';


interface FeedItemProps {
    item: { [key: string]: any },
    feed: string,
    onDelete?: (item: string) => void
}

export default function YBFeedItem(props:FeedItemProps) {
    const [textValue,setTextValue] = useState("")

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
            <YBHeading item={props.item} feed={props.feed} onDelete={props.onDelete}/>

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
                    width={600}
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
        fetch("/api/feed/"+props.feed+"/"+props.item.name,{
            method: "DELETE",
            credentials: "include"
            })
        .then(r => {
            r.text()
            .then(t => {
                    if (props.onDelete !== null) {
                        props.onDelete!(props.item.name)
                    }
                }
            )
        })
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
        <DeleteOutlined style={{float: 'right', fontSize: '14px'}} onClick={deleteItem} />
        </div>
    )
}

