import { Center, Modal, PinInput } from "@mantine/core";
import { useFocusTrap } from "@mantine/hooks";

interface PinModalProps {
    opened: boolean;
    setOpened: (open: boolean) => void;
    setPIN: (pin: string) => void;
}

export function PinModal(props:PinModalProps) {
    const {opened, setOpened, setPIN } = props
    const focusTrapRef = useFocusTrap();
    return (
        <Modal title="Set Temporary PIN" className="PINModal" opened={opened} onClose={() => setOpened(false)}>
            <div className="text-center">
                Please choose a PIN, it will expire after 2 minutes:
            </div>
            <Center>
            <PinInput ref={focusTrapRef} data-autofocus mt="1em" mb="1em" type="number" mask onComplete={(v) => { setPIN(v)}}/>
            </Center>
        </Modal>
    )
}