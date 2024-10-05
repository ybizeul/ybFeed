export const PasteToFeed = (event: ClipboardEvent, feedName: string) => {
    if (event.clipboardData === null) {
        return
    }
    const items = event.clipboardData.items
    let data, type
    //form.setFieldValue("text","")
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
    fetch("/api/feeds/" + encodeURIComponent(feedName),{
        method: "POST",
        body: data,
        headers: requestHeaders,
        credentials: "include"
      })
    //   .then(() => {
    //     form.setFieldValue("text","")
    //     if (props.onPaste) {
    //         props.onPaste()
    //     }
    //   })
}