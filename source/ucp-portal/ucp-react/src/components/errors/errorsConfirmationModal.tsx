// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Modal, SpaceBetween } from '@cloudscape-design/components';
import { DeleteAllErrorsButton } from './deleteAllErrorsButton';
import { Dispatch, SetStateAction } from 'react';

interface ErrorsConfirmationModalProps {
    isConfirmationModalVisible: boolean;
    setIsConfirmationModalVisible: Dispatch<SetStateAction<boolean>>;
    refetch: () => void;
}

export const ErrorsConfirmationModal = (props: ErrorsConfirmationModalProps) => {
    return (
        <Modal
            visible={props.isConfirmationModalVisible}
            header="Are you sure you want to clear all errors?"
            footer={
                <Box float="right">
                    <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => props.setIsConfirmationModalVisible(false)}>
                            Cancel
                        </Button>
                        <DeleteAllErrorsButton
                            refetch={props.refetch}
                            setIsConfirmationModalVisible={props.setIsConfirmationModalVisible}
                        />
                    </SpaceBetween>
                </Box>
            }
            onDismiss={() => props.setIsConfirmationModalVisible(false)}
        />
    );
};
