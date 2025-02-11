// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Badge, BadgeProps } from '@cloudscape-design/components';

export interface BaseTableBadgeProps {
    color?: BadgeProps['color'];
    text?: string;
    content?: any;
}

export default function BaseTableBadge({ color, text, content }: BaseTableBadgeProps) {
    return (
        <div>
            <Badge color={color}>{text}</Badge>
            &nbsp; {content}
        </div>
    );
}
