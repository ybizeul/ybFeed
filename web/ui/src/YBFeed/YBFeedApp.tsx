import { AppShell, Container } from "@mantine/core"

import {
    createBrowserRouter,
    RouterProvider,
  } from "react-router-dom";

import { YBFeedHome } from "./YBFeedHome";
import { YBFeedFeed } from './YBFeedFeed'
import { YBFeedVersionComponent } from "./Components";

const router = createBrowserRouter([
    {
        path: "/",
        element:<YBFeedHome/>
    },
    {
        path: "/:feedName",
        element:<YBFeedFeed/>
    },
])

export function YBFeedApp() {
    return (
        <AppShell withBorder={false}>
            <AppShell.Main>
                <Container size="md" mx="auto">
                    <RouterProvider router={router} />
                </Container>
            </AppShell.Main>
            <AppShell.Footer style={{backgroundColor:"transparent"}} zIndex={100}>
                <YBFeedVersionComponent/>
            </AppShell.Footer>
        </AppShell>
    )
}