// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useDispatch } from 'react-redux';
import { clearFormFields } from '../../store/configSlice';

export const ClearFieldsButton = () => {
    const dispatch = useDispatch();

    const onClearFieldsButtonClick = (): void => {
        dispatch(clearFormFields());
    };

    return (
        <Button variant="link" onClick={onClearFieldsButtonClick}>
            Clear Fields
        </Button>
    );
};
