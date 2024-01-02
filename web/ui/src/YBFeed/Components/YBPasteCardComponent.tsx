import { useState,useEffect } from 'react'

import { useParams } from 'react-router-dom'

import { Textarea, Paper, Center } from '@mantine/core';
import { useForm } from '@mantine/form';

import './YBPasteCardComponent.css'

interface YBPasteCardComponentProps {
    empty?: boolean
    onPaste: () => void
}

export function YBPasteCardComponent(props:YBPasteCardComponentProps) {
    const [isActive,setActive] = useState(false)
    const [isMobile, setIsMobile] = useState(false)
    const {feed} = useParams()
    const form = useForm({
        initialValues: {
          text: '',
        },
    })

    //
    // Pasting Data
    //
    
    const handleOnPaste = (event: React.ClipboardEvent) => {
        const items = event.clipboardData.items
        let data, type
        form.setFieldValue("text","")
        for (let i=0; i<items.length;i++) {
            if (items[i].type.indexOf("image") === 0 && items[i].kind === "file") {
                type = items[i].type
                data = items[i].getAsFile()
                break
            }
            else if (items[i].type === "text/plain") {
                type = items[i].type
                data = event.clipboardData.getData('text')
                break
            }
        }

        if (type === undefined) {
            return
        }

        const requestHeaders: HeadersInit = new Headers();
        requestHeaders.set("Content-Type", type)
        fetch("/api/feed/" + encodeURIComponent(feed!),{
            method: "POST",
            body: data,
            headers: requestHeaders,
            credentials: "include"
          })
          .then(() => {
            form.setFieldValue("text","")
            props.onPaste()
          })
    }

    const handleFinish = (text:string) => {
        const requestHeaders: HeadersInit = new Headers();
        requestHeaders.set("Content-Type", "text/plain")
        fetch("/api/feed/" + encodeURIComponent(feed!),{
            method: "POST",
            body: text,
            headers: requestHeaders,
            credentials: "include"
          })
          .then(() => {
            form.setFieldValue("text","")
        })
    }

    useEffect(() => {
        const handleResize = () => {
            setIsMobile(window.innerWidth <= 734); // Adjust the breakpoint as needed
        };
    
        handleResize(); // Initial call to set the initial state
    
        window.addEventListener('resize', handleResize);
    
        return () => {
          window.removeEventListener('resize', handleResize);
        };
    }, []);

    return (
        <Paper shadow="xs" p="sm" mb="1em" mt="2em" withBorder tabIndex={0} onPaste={handleOnPaste}
            style={{backgroundColor:(isActive)?"var(--mantine-color-gray-9)":"var(--mantine-color-gray-10)"}}
                onFocus={() => setActive(true)} onBlur={() => setActive(false)}
        >
            <Center>
            {(props.empty === true)?<p>Your feed is empty</p>:""}
            {isMobile?
            <form style={{width:"100%"}} onSubmit={form.onSubmit((values) => handleFinish(values.text))} >
            <Textarea ta="center" pt="1em" variant="unstyled" {...form.getInputProps('text')} placeholder='Paste Here'
            style={{textAlign:"center", textAlignLast: "center", color: "transparent", textShadow: "0px 0px 0px tomato;", caretColor:"transparent"}} />
            </form>
            :
            <p>Click and paste content here</p>
            }
            </Center>
        </Paper>
    )
}
