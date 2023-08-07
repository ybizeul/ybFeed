import { useParams } from 'react-router';
import { useEffect, useState } from "react"
import React, { FC } from 'react';
import { Navigate } from 'react-router';
import queryString from 'query-string';
import { Button, Form, message } from 'antd';
import { Modal } from 'antd';

import { Input } from 'antd';
import { Image } from 'antd';
import YBBreadCrumb from '../YBBreadCrumb'

import { DownOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { Dropdown, Space } from 'antd';

import {
    FileImageOutlined,
    FileTextOutlined,
    LinkOutlined,
    DeleteOutlined,
    NumberOutlined
  } from '@ant-design/icons';

interface FeedItemProps {
    item: { [key: string]: any },
    feed: string,
    onDelete?: (item: string) => void
}

interface PasteCardProps {
    empty?: boolean
}
const FeedItem: FC<FeedItemProps> = (props: FeedItemProps) => {
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
        <Heading item={props.item} feed={props.feed} onDelete={props.onDelete}/>

        {(props.item.type === 0)?
        <pre>{textValue}</pre>
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

const Heading: FC<FeedItemProps> = (props: FeedItemProps) => {
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

export default function Feed() {
    const params=useParams()
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const [feedItems,setFeedItems] = useState([])
    const [secret,setSecret] = useState<string|null>(null)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState(false)

    const handleOnPaste = (event: React.ClipboardEvent) => {
        const items = event.clipboardData.items
        var data, type
        console.log(items)
        for (let i=0; i<items.length;i++) {
            console.log(items[i].type)
            if (items[i].type.indexOf("image") === 0) {
                type = items[i].type
                data = items[i].getAsFile()
                break
            }
            else if (items[i].type === "text/plain") {
                type = items[i].type
                data = event.clipboardData.getData('text')
                break
            }
        }
        if (type === undefined) {
            return
        }
        const requestHeaders: HeadersInit = new Headers();
        requestHeaders.set("Content-Type", type)
        fetch("/api/feed/" + params.feed,{
            method: "POST",
            body: data,
            headers: requestHeaders,
            credentials: "include"
          })
          .then(
            update
          )
    }
    const update = () => {
        fetch("/api/feed/"+params.feed,{
            credentials: "include"
          })
        .then(r => {
            if (r.status === 200) {
                r.json().then((j) => {
                    setFeedItems(j["items"])
                    setSecret(j["secret"])
                })
                setAuthenticated(true)
            }
            else if (r.status === 401) {
                setAuthenticated(false)
            }
            else if (r.status === 402) {
                setAuthenticated(false)
            }
        })
    }
    const copyLink = () => {
        const link = window.location.href + "?secret=" + secret
        navigator.clipboard.writeText(link)
        message.info('Link Copied!')
    }
    const deleteItem = (item: string) => {
        fetch("/api/feed/"+params.feed+"/"+item,{
            method: "DELETE",
            credentials: "include"
          })
        .then(r => {
            update()
        })
    }

    useEffect(
        () => {
            const interval = window.setInterval(update,2000)
            let query = queryString.parse(window.location.search)

            if ("secret" in query) {          
                fetch("/api/feed/"+params.feed+"?secret="+query.secret,{
                    credentials: "include"
                  })
                  .then(() => {
                    setGoTo("/" + params.feed)
                    update()
                  })
            }
            else {
                update()
            }
            return () => {
                window.clearInterval(interval)
            }
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    ,[])
    const handleMenuClick: MenuProps['onClick'] = (e) => {
        if (e.key === "1") {
            copyLink()
        } else if (e.key === "2") {
            setPinModalOpen(true)
        }
      };
    const handlePinModalCancel = () => {
        setPinModalOpen(false)
    }
    const items: MenuProps['items'] = [
        {
          label: 'Copy Link',
          key: '1',
          icon: <LinkOutlined />,
        },
        {
          label: 'Set Temporary PIN',
          key: '2',
          icon: <NumberOutlined />,
        },
      ];
      const menuProps = {
        items,
        onClick: handleMenuClick,
      };
      const setPIN = (e: any) => {
        fetch("/api/feed/"+params.feed,{
            method: "PATCH",
            credentials: "include",
            body: e.PIN
          })
          .then(() => {
            message.info("PIN set")
            setPinModalOpen(false)
          })
      }
      const sendPIN = (e: any) => {
        fetch("/api/feed/"+params.feed+"?secret=" + e.PIN,{
            credentials: "include"
          })
          .then(() => {
            update()
          })
      }
    return (
        <>
        {goTo?
        <Navigate to={goTo} />
        :""}
        {authenticated?
        <Dropdown menu={menuProps}>
            <Button style={{float: 'right'}} >
                <Space>
                <LinkOutlined onClick={copyLink}/>
                <DownOutlined />
                </Space>
            </Button>
        </Dropdown>
        :""}

        <YBBreadCrumb />
        
        {authenticated?
        <>
        <Modal title="Set Temporary PIN" className="PINModal" open={pinModalOpen} footer={null} onCancel={handlePinModalCancel} destroyOnClose={true}>
            <p>Please choose a PIN, it wille expire after 2 minutes:</p>
            <Form
                onFinish={setPIN}
            >
                <Form.Item
                    name="PIN"
                    rules={[{ required: true, type: 'string', len: 4, pattern: RegExp("[0-9]{4}"), validateTrigger:"onBlur" }]}
                    validateTrigger="onBlur"
                >
                <Input size="large" width={4} type="password" maxLength={4} placeholder="1234" prefix={<NumberOutlined />} />
                </Form.Item>
            </Form>
        </Modal>

        <div className="pasteCard" onPaste={handleOnPaste}>
            <PasteCard empty={feedItems.length === 0}/>
        </div>

        {feedItems.map((f) => 
            <FeedItem item={f} feed={params.feed!} onDelete={deleteItem}/>
        )}
        </>
        :
        <>
        <p>It doesn't look like you are authorized to view this feed.</p>
        <p>Would you like to authenticate with a PIN?</p>
            <Form
                onFinish={sendPIN}
            >
                <Form.Item
                    name="PIN"
                    rules={[{ required: true, type: 'string', len: 4, pattern: RegExp("[0-9]{4}"), validateTrigger:"onBlur" }]}
                    validateTrigger="onBlur"
                >
                <Input size="large" width={4} type="password" maxLength={4} placeholder="1234" prefix={<NumberOutlined />} />
                </Form.Item>
            </Form>
        </>
        }
        </>
    )
}

const PasteCard: FC<PasteCardProps> = (props:PasteCardProps) => {
    return (
            <div className="pasteDiv" tabIndex={0}>
                {(props.empty === true)?<p>Your feed is empty</p>:""}
                Paste content here
            </div>
    )
}
