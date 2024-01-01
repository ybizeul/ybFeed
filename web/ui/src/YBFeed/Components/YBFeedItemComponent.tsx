import { useState, useEffect } from 'react'
import { Group, Modal, Button, Text, Center, Card } from "@mantine/core"
import { YBFeedItemTextComponent } from './YBFeedItemTextComponent'
import { YBFeedItemImageComponent } from './YBFeedItemImageComponent'
import { YBFeedConnector, YBFeedItem } from '../'
import { IconPhoto, IconTrash, IconTxt } from "@tabler/icons-react"
import { IconClipboardCopy } from "@tabler/icons-react"
import { notifications } from '@mantine/notifications';
import { defaultNotificationProps } from '../config';

import './YBFeedItemComponent.css'

import { copyImageItem } from './clipboard'

export interface FeedItemHeadingComponentProps {
    item: YBFeedItem,
    onDelete?: () => void,
    clipboardContent?: string,
}
function YBHeadingComponent(props: FeedItemHeadingComponentProps) {
    const { item, clipboardContent } = props
    const { name, type } = item
    const [deleteModalOpen,setDeleteModalOpen] = useState(false)

    function deleteItem() {
        setDeleteModalOpen(true)
    }
    function doDeleteItem() {
        const connection = new YBFeedConnector()
        connection.DeleteItem(item)
        .then(() => {
            setDeleteModalOpen(false)
            if (props.onDelete) {
                props.onDelete()
            }
        })
    }
    function doCopyItem() {

        if (clipboardContent) {

            navigator.clipboard.writeText(clipboardContent)
            notifications.show({message:"Copied to clipboard!", ...defaultNotificationProps})
        }
        else {
            if (item.type === 1) {
                copyImageItem(item)
                .then(() => {
                    notifications.show({message:"Copied to clipboard!", ...defaultNotificationProps})
                })
            }
    }
    }
    return (
        <Card.Section >
        <Modal title="Delete" className="DeleteModal" 
            opened={deleteModalOpen} 
            onClose={() => setDeleteModalOpen(false)}>
            <Text>Do you really want to delete file "{name}"?</Text>
            <Center mt="1em">
                <Group align='right'>
                    <Button size="xs" onClick={doDeleteItem}>OK</Button>
                    <Button size="xs" variant="outline" onClick={() => setDeleteModalOpen(false)}>Cancel</Button>
                </Group>
            </Center>
        </Modal>
        <Group ml="1em" mr="1em" justify="space-between">
        <Group>
            {(type === 0)?
            <IconTxt />
            :""}
            {(type === 1)?
            <IconPhoto />
            :""}
            &nbsp;{name}
        </Group>
        <Group>
            <Button onClick={doCopyItem} size="xs" leftSection={<IconClipboardCopy size={14} />} variant="default">
                Copy
            </Button>
            <Button onClick={deleteItem} size="xs" leftSection={<IconTrash size={14} />} variant="outline" color="red">
                Delete
            </Button>
        </Group>
        </Group>
        </Card.Section>
    )
}

export interface YBFeedItemComponentProps {
    item: YBFeedItem,
    showCopyButton?: boolean
    onUpdate?: (item: YBFeedItem) => void
    onDelete?: () => void
}

export function YBFeedItemComponent(props: YBFeedItemComponentProps) {
    const { item } = props
    const { type } = props.item

    const [textContent,setTextContent] = useState<string|undefined>(undefined)

    useEffect(() => {
        if (item.type === 0) {
            const connection = new YBFeedConnector()
            connection.GetItem(item)
            .then((text) => {
                setTextContent(text)
            })     
        }
    })

    return(
        <Card withBorder shadow="sm" radius="md" mb="2em">
            <YBHeadingComponent item={item} onDelete={props.onDelete} clipboardContent={textContent}/>
            {(type===0)?
            <YBFeedItemTextComponent>
                {textContent}
            </YBFeedItemTextComponent>
            :
            <YBFeedItemImageComponent item={item}/>
            }
        </Card>
    )
}
