import { useState, useEffect } from 'react'
import { Text, Skeleton, Center } from "@mantine/core"

export function YBFeedVersionComponent() {
        const [version, setVersion] = useState<undefined|string>(undefined)
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
            <>
            {!version?
                <Center>
                    <Skeleton mb="1em" width="10em" height="5"/>
                </Center>
            :
                <Text pb="1em" size="0.7em" c={"gray"} ta="center">ybFeed {version}</Text>
            }
            </>
        )
}