import { useState,useEffect } from 'react'
import { Input, Form } from 'antd'
import { useParams } from 'react-router-dom'
interface PasteCardProps {
    empty?: boolean
}

export function YBPasteCard(props:PasteCardProps) {
    const [isMobile, setIsMobile] = useState(false)
    const [textFieldValue,setTextFieldValue] = useState("")
    const {feed} = useParams()

       //
    // Pasting Data
    //
    
    const handleOnPaste = (event: React.ClipboardEvent) => {
        const items = event.clipboardData.items
        var data, type
        setTextFieldValue("")
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
        fetch("/api/feed/" + feed,{
            method: "POST",
            body: data,
            headers: requestHeaders,
            credentials: "include"
          })
          .then(() => {
            setTextFieldValue("")
          })
    }
    const handleFinish = () => {
        const requestHeaders: HeadersInit = new Headers();
        requestHeaders.set("Content-Type", "text/plain")
        fetch("/api/feed/" + feed,{
            method: "POST",
            body: textFieldValue,
            headers: requestHeaders,
            credentials: "include"
          })
          .then(() => {
            setTextFieldValue("")
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
        <div className="pasteDiv d-xs-none" tabIndex={0} onPaste={handleOnPaste}>
            {(props.empty === true)?<p>Your feed is empty</p>:""}
            {isMobile?
                <Form action="/" onFinish={handleFinish}>
                    <Input placeholder='Paste Here' value={textFieldValue} onChange={(e) => setTextFieldValue(e.currentTarget.value)}/>
                </Form>
            :
                "Paste content here"
            }
        </div>
    )
}
