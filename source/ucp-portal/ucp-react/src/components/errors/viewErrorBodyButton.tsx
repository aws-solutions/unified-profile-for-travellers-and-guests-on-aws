// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { IconName } from '../../models/iconName';
import { Dispatch, SetStateAction } from 'react';

interface ViewErrorBodyButtonProps {
    errorBody: string;
    setIsErrorBodyModalVisible: Dispatch<SetStateAction<boolean>>;
    setErrorBody: Dispatch<SetStateAction<string>>;
}

export const ViewErrorBodyButton = (props: ViewErrorBodyButtonProps) => {
    const onViewErrorBodyButtonClick = () => {
        props.setErrorBody(props.errorBody);
        props.setIsErrorBodyModalVisible(true);
    };

    return <Button variant="icon" onClick={onViewErrorBodyButtonClick} iconName={IconName.STATUS_INFO} />;
};
