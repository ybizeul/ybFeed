import { AppShell, Container, Text } from "@mantine/core"

import { useEffect, useState } from "react";
import {
    createBrowserRouter,
    RouterProvider,
  } from "react-router-dom";

import { YBFeedHome } from "./YBFeedHome";
import { YBFeedFeed } from './YBFeedFeed'

const router = createBrowserRouter([
    {
        path: "/",
        element:<YBFeedHome/>
    },
    {
        path: "/:feed",
        element:<YBFeedFeed/>
    },
])


export function YBFeedApp() {
    const [version, setVersion] = useState("unknown")
    useEffect(() => {
        fetch("/api")
        .then(r => {
            const v = r.headers.get("Ybfeed-Version")
            if (v !== null) {
                setVersion(v)
            }
        })
    })
    //const pinned = useHeadroom({ fixedAt: 120 });
    return (
        <AppShell withBorder={false} >
            <AppShell.Main>
                <Container size="md" mx="auto">
                    <RouterProvider router={router} />
                </Container>
            </AppShell.Main>
            <AppShell.Footer style={{backgroundColor: "rgba(0,0,0,0)"}}>
                <Text mb="1em" size="xs" ta="center">ybFeed {version}</Text>
            </AppShell.Footer>
        </AppShell>
    )
}