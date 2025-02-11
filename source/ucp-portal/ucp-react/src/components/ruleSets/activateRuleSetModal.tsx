// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Alert, Button, Modal, SpaceBetween } from '@cloudscape-design/components';
import { Dispatch, SetStateAction, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

interface ActivateRuleSetModalProps {
    isVisible: boolean;
    setVisible: Dispatch<SetStateAction<boolean>>;
    onActivate: () => void;
    isSuccess: boolean;
    isCache: boolean;
}

export default function ActivateRuleSetModal({ isVisible, setVisible, onActivate, isSuccess, isCache }: ActivateRuleSetModalProps) {
    const navigate = useNavigate();

    useEffect(() => {
        if (isSuccess) {
            navigate(0);
        }
    }, [isSuccess]);

    const onActivateRuleSetButtonClick = () => {
        onActivate();
    };

    return (
        <Modal
            visible={isVisible}
            header="Activate Rule Set"
            footer={
                <SpaceBetween direction="horizontal" size="xs">
                    <Button
                        onClick={() => {
                            setVisible(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={onActivateRuleSetButtonClick}>
                        Activate
                    </Button>
                </SpaceBetween>
            }
            onDismiss={() => setVisible(false)}
        >
            <SpaceBetween direction="vertical" size="m">
                This will replace the current active rule set with the draft rule set.
                {!isCache && (
                    <Alert statusIconAriaLabel="Warning" type="warning" header="Rule-based identity resolution">
                        Identity resolution will be re-run on all objects. This can take a significant amount of time and compute resources
                        depending on the total number of objects.
                    </Alert>
                )}
            </SpaceBetween>
        </Modal>
    );
}
