import { useEffect, useState } from "react"

import { useParams } from 'react-router';
import { Navigate } from 'react-router';

import queryString from 'query-string';

import { Button, Form, Input, Modal, Row, Col } from 'antd';
import { message } from 'antd';
import type { MenuProps } from 'antd';
import { Dropdown, Space } from 'antd';

import {
    LinkOutlined,
    NumberOutlined,
    DownOutlined
  } from '@ant-design/icons';

import YBBreadCrumb from '../YBBreadCrumb'
import YBPasteCard from '../YBPasteCard'
import YBFeedItem from '../YBFeedItem'

export default function Feed() {
    const params=useParams()
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const [feedItems,setFeedItems] = useState([])
    const [secret,setSecret] = useState<string|null>(null)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)

    //
    // Pasting Data
    //
    const handleOnPaste = (event: React.ClipboardEvent) => {
        const items = event.clipboardData.items
        var data, type

        for (let i=0; i<items.length;i++) {
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
          .then(() => {
             update()
          })
    }

    //
    // Creating links to feed
    //
    const copyLink = () => {
        const link = window.location.href + "?secret=" + secret
        navigator.clipboard.writeText(link)
        message.info('Link Copied!')
    }

    //
    // Update feed is run every 2s or o, some events
    //
    function update() {
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
            // eslint-disable-next-line react-hooks/exhaustive-deps
        },[]
    )

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
    

    //
    // Delete an item in the feed
    //
    
    const [deleteModalOpen,setDeleteModalOpen] = useState(false)
    const [deleteFileName, setDeleteFileName] = useState<string|undefined>(undefined)
    const deleteItem = (item: string) => {
        setDeleteFileName(item)
        setDeleteModalOpen(true)
    }
    const doDelete = () => {
        fetch("/api/feed/"+params.feed+"/"+deleteFileName,{
            method: "DELETE",
            credentials: "include"
          })
        .then(r => {
            update()
        })
        setDeleteModalOpen(false)
    }

    const handleDeleteModalOK = () => {
        doDelete()
    } 
    
    const handleDeleteModalCancel = () => {
        setDeleteModalOpen(false)
    } 
    return (
        <>
        {goTo?
        <Navigate to={goTo} />
        :""}
        {authenticated===true?
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
        
        {authenticated===true?
        <>
        <Modal title="Set Temporary PIN" className="PINModal" open={pinModalOpen} footer={null} onCancel={handlePinModalCancel} destroyOnClose={true}>
            <div className="text-center">
                Please choose a PIN, it wille expire after 2 minutes:
            </div>
            <Form
                onFinish={setPIN}
                >
                <Row justify='center'>
                    <Col>
                        <Form
                            action="/"
                            >
                            <Form.Item
                                name="PIN"
                                rules={[{ required: true, type: 'string', len: 4, pattern: RegExp("[0-9]{4}"), validateTrigger:"onBlur" }]}
                                validateTrigger="onBlur"
                                className='pin-field'
                            >
                            <Input size="large" width={4} type="password" maxLength={4} placeholder="1234" prefix={<NumberOutlined />} />
                            </Form.Item>
                        </Form>
                    </Col>
                </Row>
            </Form>
        </Modal>
        <Modal title="Delete" className="DeleteModal" open={deleteModalOpen} onOk={handleDeleteModalOK} onCancel={handleDeleteModalCancel} destroyOnClose={true}>
            <p>Do you really want to delete file "{deleteFileName}"?</p>
        </Modal>
        <div className="pasteCard" onPaste={handleOnPaste}>
            <YBPasteCard empty={feedItems.length === 0}/>
        </div>

        {feedItems.map((f) => 
            <YBFeedItem item={f} feed={params.feed!} onDelete={deleteItem}/>
        )}
        </>
        :""}

        {authenticated===false?
        <>
        <Row justify='center'>
            <Col>
                <div className="text-center">
                    <p>It doesn't look like you are authorized to view this feed.</p>
                    <p>Would you like to authenticate with a PIN?</p>
                </div>
            </Col>
        </Row>
        <Row justify='center'>
            <Col>
                <Form
                        onFinish={sendPIN}
                        className='form-container'
                        >
                        <Form.Item
                            name="PIN"
                            rules={[{ required: true, type: 'string', len: 4, pattern: RegExp("[0-9]{4}"), validateTrigger:"onBlur" }]}
                            validateTrigger="onBlur"
                            >
                        <Input 
                            className="pin-field"
                            size="large" 
                            width={3} 
                            type="password" 
                            maxLength={4} 
                            placeholder="1234" 
                            prefix={<NumberOutlined />} />
                        </Form.Item>
                </Form>
            </Col>
        </Row>
        </>
        :""
        }
        </>
    )
}
