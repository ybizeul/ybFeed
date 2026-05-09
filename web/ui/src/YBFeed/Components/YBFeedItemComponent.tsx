import { useState, useEffect, useContext, useCallback } from 'react'

import { Group, Button, ActionIcon, Card, Skeleton, Space, Switch, Menu, ScrollArea } from "@mantine/core"
import { useMediaQuery } from '@mantine/hooks'
import { notifications } from '@mantine/notifications';
import { IconPhoto, IconTrash, IconTxt, IconClipboardCopy, IconFile, IconDownload, IconChevronDown } from "@tabler/icons-react"
import hljs from 'highlight.js'

import { YBFeedItemTextComponent, YBFeedItemImageComponent, copyImageItem, FeedItemContext } from '.'
import { Connector, YBFeedItem } from '../'

import { defaultNotificationProps } from '../config';
import { ConfirmPopoverButton } from './ConfirmPopoverButton';

const allLanguages = hljs.listLanguages().sort()

//const connection = new YBFeedConnector()

// This is the heading component of a single item.
// Its how the item type, name and the Copy and Delete buttons
export interface FeedItemHeadingComponentProps {
    onDelete?: (item: YBFeedItem) => void,
    clipboardContent?: string,
    detectedLanguage?: string | null,
    manualLanguage?: string | null,
    highlightEnabled?: boolean,
    onHighlightToggle?: (enabled: boolean) => void,
    onLanguageChange?: (lang: string | null) => void,
}

function YBHeadingComponent(props: FeedItemHeadingComponentProps) {
    const item = useContext(FeedItemContext)
    const isMobile = useMediaQuery('(max-width: 600px)')
    
    const { clipboardContent, detectedLanguage, manualLanguage, highlightEnabled, onHighlightToggle, onLanguageChange } = props

    const activeLanguage = manualLanguage ?? detectedLanguage

    let name, type = undefined

    if (item) {
        ({name,type} = item)
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
                <Group ml="1em" mr="1em"  mt="sm" justify="center" wrap="wrap">
                    <Group style={{ flex: 1 }} wrap='nowrap' gap="xs">
                        {(type === undefined)?
                        <Skeleton width={20} height={20} />
                        :""}
                        {(type === 0)&&
                        <IconTxt />
                        }
                        {(type === 1)&&
                        <IconPhoto />
                        }
                        {(type === 2)&&
                        <IconFile />
                        }
                        <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{name}</span>
                    </Group>
                    <Group justify="center">
                        {(type === 0 && activeLanguage) &&
                        <Group gap="xs">
                            <Menu shadow="md" width={200}>
                                <Menu.Target>
                                    <Button size="xs" variant="subtle" rightSection={<IconChevronDown size={12} />}>
                                        {activeLanguage}
                                    </Button>
                                </Menu.Target>
                                <Menu.Dropdown>
                                    <ScrollArea h={200}>
                                        <Menu.Item
                                            onClick={() => onLanguageChange?.(null)}
                                            fw={!manualLanguage ? 700 : undefined}
                                        >
                                            Auto ({detectedLanguage ?? 'none'})
                                        </Menu.Item>
                                        <Menu.Divider />
                                        {allLanguages.map(lang => (
                                            <Menu.Item
                                                key={lang}
                                                onClick={() => onLanguageChange?.(lang)}
                                                fw={lang === manualLanguage ? 700 : undefined}
                                            >
                                                {lang}
                                            </Menu.Item>
                                        ))}
                                    </ScrollArea>
                                </Menu.Dropdown>
                            </Menu>
                            <Switch
                                size="xs"
                                checked={highlightEnabled ?? true}
                                onChange={(e) => onHighlightToggle?.(e.currentTarget.checked)}
                            />
                        </Group>
                        }
                        {item===undefined?
                        <>
                        <Skeleton width={80} height={20} mr="1em"/>
                        <Skeleton width={80} height={20} />
                        </>
                        :
                        <>
                        {(type === 2)?
                        isMobile ?
                        <ActionIcon component="a" href={"/api/feeds/"+encodeURIComponent(item.feed.name)+"/items/"+item.name} size="sm" variant="default" aria-label="Download">
                            <IconDownload size={14} />
                        </ActionIcon>
                        :
                        <Button component="a" href={"/api/feeds/"+encodeURIComponent(item.feed.name)+"/items/"+item.name} size="xs" leftSection={<IconDownload size={14} />} variant="default" >
                        Download
                        </Button>
                        :
                        isMobile ?
                        <ActionIcon onClick={doCopyItem} size="sm" variant="default" aria-label="Copy">
                            <IconClipboardCopy size={14} />
                        </ActionIcon>
                        :
                        <Button onClick={doCopyItem} size="xs" leftSection={<IconClipboardCopy size={14} />} variant="default" >
                            Copy
                        </Button>
        }
                        <ConfirmPopoverButton buttonTitle='Delete' message='Do you really want to delete item ?' onConfirm={() => props.onDelete&&props.onDelete(item!)}>
                            {isMobile ?
                            <ActionIcon size="sm" variant="light" color="red" aria-label="Delete">
                                <IconTrash size={14} />
                            </ActionIcon>
                            :
                            <Button size="xs" leftSection={<IconTrash size={14} />} variant="light" color="red">
                                Delete
                            </Button>
                            }
                        </ConfirmPopoverButton>
                        
                        </>}
                    </Group>
                </Group>
            </Card.Section>
        </FeedItemContext.Provider>
    )
}

export interface YBFeedItemComponentProps {
    showCopyButton?: boolean
    onDelete?: (item: YBFeedItem) => void
}

export function YBFeedItemComponent(props: YBFeedItemComponentProps) {
    const item = useContext(FeedItemContext)

    const [textContent,setTextContent] = useState<string|undefined>(undefined)
    const [detectedLanguage, setDetectedLanguage] = useState<string | null>(null)
    const [manualLanguage, setManualLanguage] = useState<string | null>(null)
    const [highlightEnabled, setHighlightEnabled] = useState(true)
    // const [timedOut, setTimedOut] = useState(false)

    // useEffect(()=> {
    //     window.setTimeout(() => setTimedOut(true),1000)
    // })
    
    useEffect(() => {
        if (item && item!.type === 0) {
            Connector.GetItem(item!)
            .then((text) => {
                setTextContent(text)
            })     
        }
    })

    const handleLanguageDetected = useCallback((lang: string | null) => {
        setDetectedLanguage(lang)
    }, [])

    if (! item) {
        return(
        <Card mt="2em" withBorder shadow="sm" radius="md" mb="2em">
            <YBHeadingComponent onDelete={props.onDelete} clipboardContent={textContent}/>
            <Skeleton mt="2em" height={50}/>
        </Card>
        )
    }
    
    return(
        <Card withBorder shadow="sm" radius="md" mb="2em">
            <YBHeadingComponent
                onDelete={props.onDelete}
                clipboardContent={textContent}
                detectedLanguage={detectedLanguage}
                manualLanguage={manualLanguage}
                highlightEnabled={highlightEnabled}
                onHighlightToggle={setHighlightEnabled}
                onLanguageChange={setManualLanguage}
            />
            {(item.type===0)&&
            <YBFeedItemTextComponent highlight={highlightEnabled} language={manualLanguage} onLanguageDetected={handleLanguageDetected}>
                {textContent}
            </YBFeedItemTextComponent>
            }
            {(item.type===1)&&
            <YBFeedItemImageComponent/>
            }
            {(item.type===2)&&
            <Space/>
            }
        </Card>
    )
}
