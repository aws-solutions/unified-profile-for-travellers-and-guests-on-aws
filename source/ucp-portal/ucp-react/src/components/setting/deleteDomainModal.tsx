// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Modal, SpaceBetween } from '@cloudscape-design/components';
import { Dispatch, SetStateAction } from 'react';
import { useSelector } from 'react-redux';
import { DeleteDomainButton } from '../../components/setting/deleteDomainButton.tsx';
import { selectCurrentDomain } from '../../store/domainSlice.ts';

interface DeleteDomainModalProps {
    isModalVisible: boolean;
    setIsModalVisible: Dispatch<SetStateAction<boolean>>;
    setIsAsyncRunning: Dispatch<SetStateAction<boolean>>;
}

export const DeleteDomainModal = ({ isModalVisible, setIsModalVisible, setIsAsyncRunning }: DeleteDomainModalProps) => {
    const selectedDomain = useSelector(selectCurrentDomain);
    const domainName = selectedDomain.customerProfileDomain;

    return (
        <Modal
            visible={isModalVisible}
            header="Are you sure you want to delete this domain?"
            footer={
                <Box float="right">
                    <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setIsModalVisible(false)}>
                            Cancel
                        </Button>
                        <DeleteDomainButton setIsModalVisible={setIsModalVisible} setIsAsyncRunning={setIsAsyncRunning} />
                    </SpaceBetween>
                </Box>
            }
            onDismiss={() => setIsModalVisible(false)}
            data-testid="deleteDomainModal"
        >
            Selected Domain: {domainName}<br></br><br></br>
            Please note: if you wish to recreate this domain with the same name, and are using Amazon Connect Customer Profiles, please wait a few minutes after deletion before recreation
        </Modal>
    );
};
