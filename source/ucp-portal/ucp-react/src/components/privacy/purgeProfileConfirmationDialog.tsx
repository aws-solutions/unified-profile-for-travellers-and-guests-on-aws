import { Box, Button, Flashbar, Modal, SpaceBetween } from '@cloudscape-design/components';

export type PurgeProfileConfirmationDialogProps = {
    headerText: string;
    isVisible: boolean;
    onDismiss: () => void;
    onConfirm: () => void;
    selectedProfiles?: string[] | undefined;
};

export function PurgeProfileConfirmationDialog({
    headerText,
    isVisible,
    onDismiss,
    onConfirm,
    selectedProfiles,
}: PurgeProfileConfirmationDialogProps) {
    return (
        <Modal
            visible={isVisible}
            onDismiss={onDismiss}
            header={headerText}
            footer={
                <Box float="right">
                    <SpaceBetween direction="horizontal" size="xs">
                        <Button onClick={onDismiss}>Cancel</Button>
                        <Button onClick={onConfirm} variant="primary">
                            Purge
                        </Button>
                    </SpaceBetween>
                </Box>
            }
        >
            <Flashbar
                items={[
                    {
                        type: 'warning',
                        content: 'Purged profiles are deleted permanently and cannot be recovered.',
                    },
                ]}
            ></Flashbar>
            {selectedProfiles && selectedProfiles.length > 0 && (
                <div>
                    <h3>Profiles to purge:</h3>
                    <ul>
                        {selectedProfiles.map(profile => (
                            <li key={profile}>{profile}</li>
                        ))}
                    </ul>
                </div>
            )}
        </Modal>
    );
}
