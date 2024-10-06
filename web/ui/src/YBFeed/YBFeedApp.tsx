import { AppShell, Container } from "@mantine/core"

import {
    BrowserRouter,
    createBrowserRouter,
    Route,
    RouterProvider,
    Routes,
  } from "react-router-dom";

import { YBFeedHome } from "./YBFeedHome";
import { YBFeedFeed } from './YBFeedFeed'
import { YBFeedVersionComponent } from "./Components";

// const router = createBrowserRouter([
//     {
//         path: "/",
//         element:<YBFeedHome/>
//     },
//     {
//         path: "/:feed",
//         element:<YBFeedFeed/>
//     },
// ])

export function YBFeedApp() {
    return (
        <AppShell withBorder={false}>
            <AppShell.Main>
                <Container size="md" mx="auto">
                    <BrowserRouter>
                        <Routes>
                            <Route path="/" element={<YBFeedHome/>}/>
                            <Route path="/:feedName" element={<YBFeedFeed/>}/>
                        </Routes>
                    </BrowserRouter>
                </Container>
            </AppShell.Main>
            <AppShell.Footer style={{backgroundColor:"transparent"}} zIndex={100}>
                <YBFeedVersionComponent/>
            </AppShell.Footer>
        </AppShell>
    )
}