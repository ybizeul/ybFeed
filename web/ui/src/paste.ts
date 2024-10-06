import { Y } from "./YBFeedClient"

export const PasteToFeed = (event: ClipboardEvent, feedName: string) => {
    if (event.clipboardData === null) {
        return
    }

    const formData = new FormData();

    const items = event.clipboardData.items
    let data: File|string|null = null
    let type: string|null = null

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

    if (type === null || data === null) {
        return
    }

    const blob = new Blob([data], { type: type });
    formData.append("clipboard", blob);
    
    const options={ headers: {"Content-Type": type}}
    Y.post("/feeds/" + encodeURIComponent(feedName), formData, options)
}
