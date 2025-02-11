// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Alert, Box, Button, Modal } from '@cloudscape-design/components';
import { Dispatch, SetStateAction } from 'react';

interface ErrorBodyModalProps {
    isErrorBodyModalVisible: boolean;
    setIsErrorBodyModalVisible: Dispatch<SetStateAction<boolean>>;
    errorBody: string;
}

export const ErrorBodyModal = (props: ErrorBodyModalProps) => {
    let errorBody: string;
    let isInvalidJson = false;
    try {
        errorBody = props.errorBody ? JSON.stringify(JSON.parse(props.errorBody), null, 2) : 'No Error Body Available';
    } catch {
        isInvalidJson = true;
        errorBody = props.errorBody;
    }

    const isRawKinesisRecord = props.errorBody.includes('"eventName": "aws:kinesis:record"');

    return (
        <Modal
            visible={props.isErrorBodyModalVisible}
            size="large"
            header="Ingested Record Detail"
            footer={
                <Box float="right">
                    <Button variant="link" onClick={() => props.setIsErrorBodyModalVisible(false)}>
                        Close
                    </Button>
                </Box>
            }
            onDismiss={() => props.setIsErrorBodyModalVisible(false)}
        >
            {isInvalidJson && invalidJsonAlert}
            {isRawKinesisRecord && rawKinesisRecordAlert}
            <pre>{errorBody}</pre>
        </Modal>
    );
};

const invalidJsonAlert = (
    <Alert statusIconAriaLabel="Warning" type="warning">
        Invalid JSON format detected
    </Alert>
);

const rawKinesisRecordAlert = (
    <Alert statusIconAriaLabel="Warning">
      The Kinesis object's data field is base64 encoded.
      Base64 decode to view the message.
    </Alert>
);
