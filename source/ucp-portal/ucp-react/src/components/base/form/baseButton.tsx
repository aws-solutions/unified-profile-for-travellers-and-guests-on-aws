// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ButtonProps } from '@cloudscape-design/components';
import { IconName } from '../../../models/iconName';

export interface BaseButtonProps {
    title?: string;
    iconName?: IconName;
    disabled?: boolean;
    variant?: ButtonProps.Variant;
    arialabel?: string;
    handleClick?: () => void;
    href?: string;
    isExternal?: boolean;
}

export default function BaseButton({
    title,
    iconName,
    disabled,
    variant = 'primary',
    arialabel,
    handleClick,
    href,
    isExternal,
}: BaseButtonProps) {
    return (
        <>
            <Button
                variant={variant}
                iconName={iconName}
                onClick={handleClick}
                disabled={disabled}
                ariaLabel={arialabel}
                href={href}
                target={isExternal ? '_blank' : ''}
            >
                {title}
            </Button>
        </>
    );
}
