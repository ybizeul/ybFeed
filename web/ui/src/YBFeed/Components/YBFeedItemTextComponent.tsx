
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function YBFeedItemTextComponent(props: any) {
    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll"}}>{props.children}</pre>
            </div>
        </div>
    )
}