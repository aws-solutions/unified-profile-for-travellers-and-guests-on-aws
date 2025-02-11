// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Checkbox, Modal, SpaceBetween } from '@cloudscape-design/components';
import { selectCurrentDomain } from '../../store/domainSlice.ts';
import { useSelector } from 'react-redux';
import { Dispatch, SetStateAction, useState } from 'react';
import { RebuildDomainCacheButton } from './rebuildDomainCacheButton.tsx';
import { CacheMode } from '../../models/domain.ts';

interface RebuildCacheModalProps {
    isModalVisible: boolean;
    setIsModalVisible: Dispatch<SetStateAction<boolean>>;
}

export const RebuildCacheModal = ({ isModalVisible, setIsModalVisible }: RebuildCacheModalProps) => {
    const selectedDomain = useSelector(selectCurrentDomain);
    const [cpChecked, setCpChecked] = useState<boolean>(false);
    const [dynamoChecked, setDynamoChecked] = useState<boolean>(false);

    const customerProfileEnabled = (selectedDomain.cacheMode & CacheMode.CP_MODE) === CacheMode.CP_MODE;
    const dynamoEnabled = (selectedDomain.cacheMode & CacheMode.DYNAMO_MODE) === CacheMode.DYNAMO_MODE;

    let cacheMode = CacheMode.NONE;
    if (cpChecked) {
        cacheMode = cacheMode | CacheMode.CP_MODE;
    }
    if (dynamoChecked) {
        cacheMode = cacheMode | CacheMode.DYNAMO_MODE;
    }

    return (
        <Modal
            visible={isModalVisible}
            header="Select which caches to rebuild"
            footer={
                <Box float="right">
                    <SpaceBetween direction="horizontal" size="xs">
                        <Button variant="link" onClick={() => setIsModalVisible(false)}>
                            Cancel
                        </Button>
                        <RebuildDomainCacheButton
                            setIsModalVisible={setIsModalVisible}
                            setCpChecked={setCpChecked}
                            setDynamoChecked={setDynamoChecked}
                            cacheMode={cacheMode}
                        />
                    </SpaceBetween>
                </Box>
            }
            onDismiss={() => setIsModalVisible(false)}
            data-testid="rebuildCacheModal"
        >
            <SpaceBetween direction="vertical" size="xs">
                {customerProfileEnabled ? (
                    <Checkbox
                        onChange={({ detail }) => {
                            setCpChecked(detail.checked);
                        }}
                        checked={cpChecked}
                    >
                        Customer Profiles
                    </Checkbox>
                ) : (
                    <></>
                )}
                {dynamoEnabled ? (
                    <Checkbox
                        onChange={({ detail }) => {
                            setDynamoChecked(detail.checked);
                        }}
                        checked={dynamoChecked}
                    >
                        DynamoDB
                    </Checkbox>
                ) : (
                    <></>
                )}
            </SpaceBetween>
        </Modal>
    );
};
