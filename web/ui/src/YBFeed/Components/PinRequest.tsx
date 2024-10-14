import { Center, Text, PinInput } from "@mantine/core";
import { useFocusTrap } from "@mantine/hooks";

interface PinRequestProps {
    sendPIN: (pin: string) => void;
}

export function PinRequest(props:PinRequestProps) {
    const { sendPIN } = props
    const focusTrapRef = useFocusTrap();
    return (
        <>
            <Text mt="2em" ta="center">This feed is protected by a PIN.</Text>
            <Center>
                <PinInput ref={focusTrapRef} mt="2em" type="number" mask onComplete={(v) => { sendPIN(v)}}/>
            </Center>
        </>
    )
}