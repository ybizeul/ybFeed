import { useState,useEffect } from 'react'

import { useParams } from 'react-router-dom'

import { Textarea } from '@mantine/core';
import { useForm } from '@mantine/form';

import './YBPasteCardComponent.css'

interface YBPasteCardComponentProps {
    empty?: boolean
    onPaste: () => void
}

export function YBPasteCardComponent(props:YBPasteCardComponentProps) {
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
        <div id="pasteCard" className="pasteDiv" tabIndex={0} onPaste={handleOnPaste} >
            {(props.empty === true)?<p>Your feed is empty</p>:""}
            {isMobile?
            <form onSubmit={form.onSubmit((values) => handleFinish(values.text))}>
            <Textarea {...form.getInputProps('text')} placeholder='Paste Here' />
            </form>
            :
            <p>Click and paste content here</p>
            }
        </div>
    )
}
