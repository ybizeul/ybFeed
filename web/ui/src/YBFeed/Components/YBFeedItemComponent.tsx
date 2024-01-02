import { useState, useEffect, useContext } from 'react'

import { Group, Modal, Button, Text, Center, Card, Skeleton } from "@mantine/core"
import { notifications } from '@mantine/notifications';
import { IconPhoto, IconTrash, IconTxt, IconClipboardCopy } from "@tabler/icons-react"

import { YBFeedItemTextComponent, YBFeedItemImageComponent, copyImageItem, FeedItemContext } from '.'
import { YBFeedConnector, YBFeedItem } from '../'

import { defaultNotificationProps } from '../config';

const connection = new YBFeedConnector()

// This is the heading component of a single item.
// Its how the item type, name and the Copy and Delete buttons
export interface FeedItemHeadingComponentProps {
    onDelete?: () => void,
    clipboardContent?: string,
}

function YBHeadingComponent(props: FeedItemHeadingComponentProps) {
    const item = useContext(FeedItemContext)
    
    const { clipboardContent } = props

    let name, type = undefined

    if (item) {
        ({name,type} = item)
    }

    const [deleteModalOpen,setDeleteModalOpen] = useState(false)

    // Display delete item confirmation dialog
    function deleteItem() {
        setDeleteModalOpen(true)
    }

    // Do the actual item deletion callback
    function doDeleteItem() {
        connection.DeleteItem(item!)
        .then(() => {
            setDeleteModalOpen(false)
            if (props.onDelete) {
                props.onDelete()
            }
        })
    }

    // Copy item to pasteboard
    // if `clipboardContent` is set as an attribute, this is what will be put
    // in the clipboard, otherwise, we are assuming that's an image.
    function doCopyItem() {
        if (clipboardContent) {
            navigator.clipboard.writeText(clipboardContent)
            notifications.show({message:"Copied to clipboard!", ...defaultNotificationProps})
        }
        else {
            if (item!.type === 1) {
                copyImageItem(item!)
                .then(() => {
                    notifications.show({message:"Copied to clipboard!", ...defaultNotificationProps})
                })
            }
        }
    }

    return (
        <FeedItemContext.Provider value={item}>
            <Card.Section >
                <Modal title="Delete" className="DeleteModal" 
                    opened={deleteModalOpen} 
                    onClose={() => setDeleteModalOpen(false)}>
                    <Text>Do you really want to delete item "{name}"?</Text>
                    <Center mt="1em">
                        <Group align='right'>
                            <Button size="xs" color="red" onClick={doDeleteItem}>Delete</Button>
                            <Button size="xs" onClick={() => setDeleteModalOpen(false)}>Cancel</Button>
                        </Group>
                    </Center>
                </Modal>
                <Group ml="1em" mr="1em" justify="space-between">
                    <Group>
                        {(type === undefined)?
                        <Skeleton width={20} height={20} />
                        :""}
                        {(type === 0)?
                        <IconTxt />
                        :""}
                        {(type === 1)?
                        <IconPhoto />
                        :""}
                        &nbsp;{name}
                    </Group>
                    <Group>
                        {item===undefined?
                        <>
                        <Skeleton width={80} height={20} mr="1em"/>
                        <Skeleton width={80} height={20} />
                        </>
                        :
                        <>
                        <Button onClick={doCopyItem} size="xs" leftSection={<IconClipboardCopy size={14} />} variant="default" >
                            Copy
                        </Button>
                        <Button onClick={deleteItem} size="xs" leftSection={<IconTrash size={14} />} variant="light" color="red">
                            Delete
                        </Button>
                        </>}
                    </Group>
                </Group>
            </Card.Section>
        </FeedItemContext.Provider>
    )
}

export interface YBFeedItemComponentProps {
    showCopyButton?: boolean
    onUpdate?: (item: YBFeedItem) => void
    onDelete?: () => void
}

export function YBFeedItemComponent(props: YBFeedItemComponentProps) {
    const item = useContext(FeedItemContext)

    const [textContent,setTextContent] = useState<string|undefined>(undefined)

    useEffect(() => {
        if (item && item!.type === 0) {
            const connection = new YBFeedConnector()
            connection.GetItem(item!)
            .then((text) => {
                setTextContent(text)
            })     
        }
    })

    return(
        <>
        {!item?
            <Card mt="2em" withBorder shadow="sm" radius="md" mb="2em">
                <YBHeadingComponent onDelete={props.onDelete} clipboardContent={textContent}/>
                <Skeleton mt="2em" height={50}/>
            </Card>
        :
        <Card withBorder shadow="sm" radius="md" mb="2em">
            <YBHeadingComponent onDelete={props.onDelete} clipboardContent={textContent}/>
            {(item.type===0)?
            <YBFeedItemTextComponent>
                {textContent}
            </YBFeedItemTextComponent>
            :
            <YBFeedItemImageComponent/>
            }
        </Card>}
        </>
    )
}
