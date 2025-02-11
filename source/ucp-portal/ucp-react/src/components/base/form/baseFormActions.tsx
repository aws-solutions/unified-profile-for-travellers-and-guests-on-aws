// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, BoxProps, Button, SpaceBetween } from '@cloudscape-design/components';

interface BaseFormActionsProps {
    submitTitle: string;
    handleCancel: () => void;
    handleSubmit: () => void;
    cancelLabel?: string;
    submitLabel?: string;
    disableSubmit?: boolean;
    float?: BoxProps.Float;
}
export default function BaseFormActions({
    submitTitle,
    handleCancel,
    handleSubmit,
    cancelLabel = 'cancel',
    submitLabel = 'submit',
    disableSubmit = false,
    float = 'right',
}: BaseFormActionsProps) {
    return (
        <>
            <Box float={float}>
                <SpaceBetween direction="horizontal" size="xs">
                    <Button variant="link" ariaLabel={cancelLabel} onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button variant="primary" ariaLabel={submitLabel} onClick={handleSubmit} disabled={disableSubmit}>
                        {submitTitle}
                    </Button>
                </SpaceBetween>
            </Box>
        </>
    );
}
