import { useEffect, useState, useRef } from "react"

import { useParams } from 'react-router';
import { Navigate } from 'react-router';

import queryString from 'query-string';

import { Button, Form, Input, Dropdown, Space, Modal, Row, Col } from 'antd';
import { message } from 'antd';
import type { MenuProps } from 'antd';

import {
    LinkOutlined,
    NumberOutlined,
    DownOutlined
  } from '@ant-design/icons';

import { YBBreadCrumb, YBPasteCard, FeedItem, FeedItems } from '../YBFeed'

interface Feed {
    items: FeedItem[],
    secret: string
}

export default function FeedComponent() {
    const { feed } = useParams();
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const feedItems = useRef<FeedItem[]>([])
    const [secret,setSecret] = useState<string|null>(null)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [updateGeneration,setUpdateGeneration] = useState(0)

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
        fetch("/api/feed/"+encodeURIComponent(feed!),{
            credentials: "include"
          })
        .then(r => {
            if (r.status === 200) {
                let update=false
                r.json().then((f: Feed) => {
                    // Remove any deleted items
                    const keep: FeedItem[] = []
                    feedItems.current.map((i:FeedItem) => {
                        let changed = false
                        for (let j=0;j<f["items"].length;j++) {
                            const current_new_item = f["items"][j]
                            if (current_new_item.name === i.name) {
                                changed = true
                                keep.push(i)
                            }
                        }
                        if (changed === false) {
                            update=true
                        }
                        return null
                    })
                    feedItems.current.length = 0
                    feedItems.current = [...keep]

                    // Add new items
                    const newItems = f["items"].map((i: FeedItem) => {
                        for (let j=0;j<feedItems.current.length;j++) {
                                const current_existing_item = feedItems.current[j]
                                if (current_existing_item.name === i.name) {
                                    return feedItems.current[j]
                                }
                        }
                        update=true
                        i.feed = feed!
                        return i
                    })

                    feedItems.current.length = 0
                    feedItems.current = [...newItems]
                    if (update === true) {
                        setUpdateGeneration(updateGeneration+1)
                    }
                    setSecret(f["secret"])
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
                fetch("/api/feed/"+feed+"?secret="+query.secret,{
                    credentials: "include"
                  })
                  .then(() => {
                    setGoTo("/" + feed)
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
        },[updateGeneration]
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
          label: 'Copy Permalink',
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
        fetch("/api/feed/"+feed,{
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
        fetch("/api/feed/"+feed+"?secret=" + e.PIN,{
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
            <Row justify='center'>
                    <Col>
                                <Form
                action="/"
                onFinish={setPIN}
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

        </Modal>

        <div className="pasteCard">
            <YBPasteCard empty={feedItems.current.length === 0}/>
        </div>

        <FeedItems items={feedItems.current} onUpdate={update} />
        
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
