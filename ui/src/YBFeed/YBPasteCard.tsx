interface PasteCardProps {
    empty?: boolean
}

export function YBPasteCard(props:PasteCardProps) {

//export default const PasteCard: FC<PasteCardProps> = (props:PasteCardProps) => {
    return (
            <div className="pasteDiv" tabIndex={0}>
                {(props.empty === true)?<p>Your feed is empty</p>:""}
                Paste content here
            </div>
    )
}
