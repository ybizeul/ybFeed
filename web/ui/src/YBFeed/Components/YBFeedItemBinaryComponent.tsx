
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function YBFeedItemBinaryComponent(props: any) {
    return(
        <div className="itemContainer">
            <div className="itemText">
                <pre style={{overflowY:"scroll", fontSize: "0.8em"}}>{props.children}</pre>
            </div>
        </div>
    )
}