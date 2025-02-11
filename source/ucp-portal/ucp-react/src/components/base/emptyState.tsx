// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box } from '@cloudscape-design/components';
import { ReactNode } from 'react';

interface EmptyStateProps {
    title: string;
    subtitle: string;
    action: ReactNode;
}

/**
 *
 * @param props
 * @returns {ReactNode} to replace empty state in table
 */
export default function EmptyState(props: EmptyStateProps) {
    return (
        <Box textAlign="center" color="inherit">
            <Box variant="strong" textAlign="center" color="inherit">
                {props.title}
            </Box>
            <Box variant="p" padding={{ bottom: 's' }} color="inherit">
                {props.subtitle}
            </Box>
            {props.action}
        </Box>
    );
}
