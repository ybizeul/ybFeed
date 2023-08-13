import { useEffect, useState } from 'react'
import { Space } from 'antd'

import {
    FileImageOutlined,
    FileTextOutlined,
    DeleteOutlined,
  } from '@ant-design/icons';

  import { FeedItemText } from './YBFeedItemText'
  import { FeedItemImage } from './YBFeedItemImage'

export interface FeedItem {
    name: string,
    date: string,
    type: number,
    feed: string
}

export interface FeedItemProps {
    item: FeedItem,
    onDelete?: (item: FeedItem) => void
}

function YBHeading(props: FeedItemProps) {
    const { item, onDelete} = props
    const { name, type } = props.item
    const deleteItem = () => {
        if (onDelete !== undefined) {
            onDelete(item)
        }
    }
    return (
        <div className='heading'>
        {(type === 0)?
        <FileTextOutlined />
        :""}
        {(type === 1)?
        <FileImageOutlined />
        :""}
        &nbsp;{name}
        <Space style={{float:'right'}}>
            <DeleteOutlined style={{fontSize: '14px', color: 'red'}} onClick={deleteItem} />
        </Space>
        </div>
    )
}

export function YBFeedItem(props: FeedItemProps) {
    const { item, onDelete} = props
    const { type } = props.item

    return(
        <div className='item'>
            <YBHeading item={item} onDelete={onDelete}/>

            {(type === 0)?<FeedItemText item={item}/>:""}
            {(type === 1)?<FeedItemImage item={item}/>:""}
        </div>
    )
}