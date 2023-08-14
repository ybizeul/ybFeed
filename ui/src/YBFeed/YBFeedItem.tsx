import { useState,useEffect } from 'react'
import { Space, Modal } from 'antd'

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
    showCopyButton?: boolean
    onUpdate?: (item: FeedItem) => void
}

export interface FeedItemHeadingProps {
    item: FeedItem,
}
function YBHeading(props: FeedItemHeadingProps) {
    const { item } = props
    const { name, type } = props.item
    const [deleteModalOpen,setDeleteModalOpen] = useState(false)

    function deleteItem() {
        setDeleteModalOpen(true)
    }
    function doDeleteItem() {
        fetch("/api/feed/"+item.feed+"/"+item.name,{
            method: "DELETE",
            credentials: "include"
            })
            .then(() => setDeleteModalOpen(false))
    }

    return (
        <div className='heading'>
        <Modal title="Delete" className="DeleteModal" open={deleteModalOpen} onOk={doDeleteItem} onCancel={() => setDeleteModalOpen(false)} >
            <p>Do you really want to delete file "{item.name}"?</p>
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

export function YBFeedItem(props: FeedItemProps) {
    const { item } = props
    const { type } = props.item

    const [isMobile, setIsMobile] = useState(false)

    console.log("Render YBFeedItem")

    let component
    if (type === 0){
        component = FeedItemText({item: item, showCopyButton:!isMobile})
    } else {
        component = FeedItemImage({item: item, showCopyButton:!isMobile})
    }


    useEffect(() => {
        const handleResize = () => {
          setIsMobile(window.innerWidth <= 576); // Adjust the breakpoint as needed
        };
    
        handleResize(); // Initial call to set the initial state
    
        window.addEventListener('resize', handleResize);
    
        return () => {
          window.removeEventListener('resize', handleResize);
        };
    }, []);

    return(
        <div className='item'>
            <YBHeading item={item} />
            {component}
        </div>
    )
}
