import { useState,useEffect } from 'react'

import { redirect, useParams } from 'react-router-dom'

import { Textarea, Center, Text } from '@mantine/core';
import { Dropzone } from '@mantine/dropzone';
// import { useForm } from '@mantine/form';

import './YBPasteCardComponent.css'
import { PasteToFeed } from '../../paste';
import { Y } from '../../YBFeedClient';

interface YBPasteCardComponentProps {
    onPaste?: () => void
}

export function YBPasteCardComponent(props:YBPasteCardComponentProps) {
    const [isMobile, setIsMobile] = useState(false)
    const {feedName} = useParams()

    if (!feedName) {
        redirect("/")
        return
    }
    // const form = useForm({
    //     initialValues: {
    //       text: '',
    //     },
    // })

    // //
    // // Pasting Data
    // //
    
    // const handleOnPaste = (event: React.ClipboardEvent) => {
    //     const items = event.clipboardData.items
    //     let data, type
    //     form.setFieldValue("text","")
    //     for (let i=0; i<items.length;i++) {
    //         if (items[i].type.indexOf("image") === 0 && items[i].kind === "file") {
    //             type = items[i].type
    //             data = items[i].getAsFile()
    //             break
    //         }
    //         else if (items[i].type === "text/plain") {
    //             type = items[i].type
    //             data = event.clipboardData.getData('text')
    //             break
    //         }
    //     }

    //     if (type === undefined) {
    //         return
    //     }

    //     const requestHeaders: HeadersInit = new Headers();
    //     requestHeaders.set("Content-Type", type)
    //     fetch("/api/feeds/" + encodeURIComponent(feed!),{
    //         method: "POST",
    //         body: data,
    //         headers: requestHeaders,
    //         credentials: "include"
    //       })
    //       .then(() => {
    //         form.setFieldValue("text","")
    //         if (props.onPaste) {
    //             props.onPaste()
    //         }
    //       })
    // }

    // const handleFinish = (text:string) => {
    //     const requestHeaders: HeadersInit = new Headers();
    //     requestHeaders.set("Content-Type", "text/plain")
    //     fetch("/api/feeds/" + encodeURIComponent(feed!),{
    //         method: "POST",
    //         body: text,
    //         headers: requestHeaders,
    //         credentials: "include"
    //       })
    //       .then(() => {
    //         form.setFieldValue("text","")
    //     })
    // }

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

    useEffect(() => {
        document.onpaste = (e) => {
            PasteToFeed(e,feedName)
        }
        return () => {
            document.onpaste = null
        }
    })

    return (
        <Center my="2em" h="100%" style={{ flexDirection:"column"}}>
            {isMobile&&
            // <form style={{width:"100%"}} onSubmit={form.onSubmit((values) => handleFinish(values.text))} >
                <Textarea ta="center" pt="1em" variant="unstyled" placeholder='Paste Here' value={""} onChange={() => {}}
                style={{textAlign:"center", textAlignLast: "center", color: "transparent", textShadow: "0px 0px 0px tomato", caretColor:"transparent"}} />
            // </form>
            }
            <Dropzone.FullScreen w="100%" ta="center"
                onDrop={(files) => {
                    const formData = new FormData();
                    formData.append("file", files[0]);
                    Y.post("/feeds/" + encodeURIComponent(feedName), formData)
                }}
                onReject={(files) => console.log('rejected files', files)}
                maxSize={5 * 1024 ** 2}
                ><Center h={"100vh"}>
                    Drop files here
                    </Center>
            </Dropzone.FullScreen>
            
        </Center>
    )
}
