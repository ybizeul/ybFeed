import { Navigate } from 'react-router';
import { Form, Input } from 'antd';
import { useState } from "react"
import YBBreadCrumb from '../YBBreadCrumb'


export default function Root() {
    const [feed,setFeed] = useState("")
    const [goToPath,setGoToPath] = useState("")

    if (window.location.pathname !== "/") {
        setGoToPath(window.location.pathname)
    }

    return (
        <>
        {goToPath && (
            <Navigate to={goToPath} />
        )}
        <YBBreadCrumb />
        <p><b>Welcome to ybFeed</b></p>
        <p>
            Choose a unique name for your feed :
        </p>
        <Form name="basic" layout="inline" onFinish={handleFinish}>
            <Form.Item
                name="feed"
            >
                <Input
                    placeholder="Feed name"
                    onChange={(e) => setFeed(e.currentTarget.value)}
                />
            </Form.Item>
        </Form>
        </>
    )
}