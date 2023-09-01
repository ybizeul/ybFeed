import { useState } from 'react'
import { Space, Modal } from 'antd'

import {
    FileImageOutlined,
    FileTextOutlined,
    DeleteOutlined,
  } from '@ant-design/icons';

import { FeedItemText } from './YBFeedItemTextComponent'
import { FeedItemImage } from './YBFeedItemImageComponent'
import { FeedConnector, FeedItem } from '.'

export interface FeedItemHeadingProps {
    item: FeedItem,
    onDelete?: () => void
}
function YBHeading(props: FeedItemHeadingProps) {
    const { item } = props
    const { name, type } = item
    const [deleteModalOpen,setDeleteModalOpen] = useState(false)

    function deleteItem() {
        setDeleteModalOpen(true)
    }
    function doDeleteItem() {
        var connection = new FeedConnector()
        connection.DeleteItem(item)
        .then(() => {
            setDeleteModalOpen(false)
            if (props.onDelete) {
                props.onDelete()
            }
        })
    }

    return (
        <div className='heading'>
        <Modal title="Delete" className="DeleteModal" open={deleteModalOpen} onOk={doDeleteItem} onCancel={() => setDeleteModalOpen(false)} >
            <p>Do you really want to delete file "{name}"?</p>
        </Modal>
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

export interface FeedItemProps {
    item: FeedItem,
    showCopyButton?: boolean
    onUpdate?: (item: FeedItemProps) => void
    onDelete?: () => void
}

export function YBFeedItem(props: FeedItemProps) {
    const { item } = props
    const { type } = props.item

    let component
    if (type === 0){
        component = FeedItemText({item: item, showCopyButton:true})
    } else {
        component = FeedItemImage({item: item, showCopyButton:true})
    }

    return(
        <div className='item'>
            <YBHeading item={item} onDelete={props.onDelete}/>
            {component}
        </div>
    )
}
