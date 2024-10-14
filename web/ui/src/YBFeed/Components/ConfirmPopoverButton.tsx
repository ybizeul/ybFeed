import { Box, Button, Popover, Text } from "@mantine/core";
import { PropsWithChildren } from "react";
import classes from "./ConfirmPopoverButton.module.css";
import { useToggle } from "@mantine/hooks";

interface ConfirmPopoverButtonProps {
    message: string
    buttonTitle: string
    onConfirm: () => void
}
export function ConfirmPopoverButton(props:PropsWithChildren<ConfirmPopoverButtonProps>) {
    const [opened,setOpened] = useToggle()

    return (
        <Popover opened={opened} onClose={() => setOpened(false)} width={200} position="bottom" withArrow shadow="md">
            <Popover.Target>
                <Box onClick={() => setOpened(true)}>
                {props.children}
                </Box>
            </Popover.Target>
            <Popover.Dropdown className={classes.popover} >
                <Text ta="center" size="xs" mb="xs">{props.message}</Text>
                <Button w="100%" size="xs" variant="default" c="red" onClick={() => {props.onConfirm(); setOpened(false)}}>{props.buttonTitle}</Button>
            </Popover.Dropdown>
        </Popover>
    )
}