import { useEffect, useState, useRef } from "react"

import { useParams, useSearchParams } from 'react-router-dom';
import { Navigate } from 'react-router-dom';

import { Button, Form, Input, Dropdown, Space, Modal, Row, Col } from 'antd';
import { message } from 'antd';
import type { MenuProps } from 'antd';

import { FeedConnector, YBBreadCrumb, YBPasteCard, FeedItems, FeedItem } from '.'

import {
    LinkOutlined,
    NumberOutlined,
    DownOutlined
  } from '@ant-design/icons';



export function FeedComponent() {
    const feedParam: string = useParams().feed!
    const [searchParams] = useSearchParams()
    const [goTo,setGoTo] = useState<string|undefined>(undefined)
    const feedItems = useRef<FeedItem[]>([])
    const [secret,setSecret] = useState<string|null>(null)
    const [pinModalOpen,setPinModalOpen] = useState(false)
    const [authenticated,setAuthenticated] = useState<boolean|undefined>(undefined)
    const [updateGeneration,setUpdateGeneration] = useState(0)
    const [fatal, setFatal] = useState(null)
    const connection = new FeedConnector()
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
        if (feedParam === undefined) {
            return
        }
        connection.GetFeed(feedParam)
        .then((f) => {
            if (f === null) {
                return
            }
            setFatal(null)
            var do_update = false

            var found

            // Loop over current items and keep what is already here

            const oldItems = []
            for (let i=0;i<feedItems.current.length;i++) {
                var current_old_item = feedItems.current[i]
                found = false
                for (let j=0;j<f.items.length;j++) {
                    const current_new_item = f.items[j]
                    if (current_new_item.name === current_old_item.name) {
                        found = true
                        oldItems.push(feedItems.current[i])
                    }
                }
                if (found === false) {
                    do_update = true
                }
            }
            feedItems.current.length = 0
            feedItems.current = [...oldItems]

            // Loop over new items and add what is new
            for (let i=0;i<f.items.length;i++) {
                var current_new_item = f.items[i]
                found = false
                for (let j=0;j<feedItems.current.length;j++) {
                    const current_existing_item = feedItems.current[j]
                    if (current_existing_item.name === current_new_item.name) {
                        found = true
                    }
                }
                if (found === false) {
                    feedItems.current.push(f.items[i])
                    do_update=true
                }
            }

            feedItems.current.sort((a,b) =>{
                return (a.date < b.date)?1:-1
            })
            if (do_update === true) {
                setUpdateGeneration(updateGeneration+1)
            }

            setSecret(f.secret)
            setAuthenticated(true)
        })
        .catch(e => {
            if (e.status === 401) {
                setAuthenticated(false)
            } else if (e.status === 500) {
                setFatal(e.message)
            }
        })
    }

    useEffect(
        () => {
            const interval = window.setInterval(update,2000)
            const secret = searchParams.get("secret")
            if (secret) {
                console.log("test")
                connection.AuthenticateFeed(feedParam,secret)
                    .then(() => {
                        console.log("then")
                        setGoTo("/" + feedParam)
                        update()
                    })
                    .catch((e) => {
                        console.log(e.message)
                        message.error(e.message)
                        setAuthenticated(false)
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
        connection.SetPIN(feedParam,e.PIN)
        .then((r) => {
            message.info("PIN set")
            setPinModalOpen(false)
        })
        .catch((e) => {
            message.error(e.message)
            setPinModalOpen(false)
        })
      }
      const sendPIN = (e: any) => {
        connection.AuthenticateFeed(feedParam,e.PIN)
        .then((r) => {
            update()
        })
        .catch((e) => {
            message.error(e.message)
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

        {!fatal?
            <>
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
                <YBPasteCard empty={feedItems.current.length === 0} onPaste={update}/>
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
        :
            <p>{fatal}</p>
        }   
        </>
    )
}
