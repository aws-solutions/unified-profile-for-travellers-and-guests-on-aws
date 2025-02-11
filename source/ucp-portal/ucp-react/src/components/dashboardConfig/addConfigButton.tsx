// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { IconName } from '../../models/iconName';
import { HyperLinkMapping } from '../../models/config';
import { usePostPortalConfigMutation } from '../../store/configApiSlice';
import { useDispatch, useSelector } from 'react-redux';
import {
    clearFormFields,
    selectDataFieldInput,
    selectDataRecordInput,
    selectHyperLinkMappings,
    selectUrlTemplateInput,
} from '../../store/configSlice';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

export const AddConfigButton = () => {
    const dispatch = useDispatch();
    const currentTableItems = useSelector(selectHyperLinkMappings);
    const dataRecord = useSelector(selectDataRecordInput);
    const dataField = useSelector(selectDataFieldInput);
    const urlTemplate = useSelector(selectUrlTemplateInput);

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserAddConfig = (userAppAccess & Permissions.SaveHyperlinkPermission) === Permissions.SaveHyperlinkPermission;

    const isAddConfigButtonDisabled = dataRecord === null || dataField === null || urlTemplate === '' || !canUserAddConfig;

    const [postPortalConfigTrigger] = usePostPortalConfigMutation();

    const onAddConfigButtonClick = (): void => {
        if (dataRecord && dataField && dataRecord.value && dataField.value && urlTemplate) {
            const newConfig: HyperLinkMapping = {
                accpObject: dataRecord.value,
                fieldName: dataField.value,
                hyperlinkTemplate: urlTemplate,
            };
            const tableItemCopy = [...currentTableItems];
            tableItemCopy.push(newConfig);
            postPortalConfigTrigger(tableItemCopy);
            dispatch(clearFormFields());
        }
    };

    return (
        <Button
            variant="primary"
            onClick={onAddConfigButtonClick}
            iconName={IconName.ADD}
            iconAlign="left"
            disabled={isAddConfigButtonDisabled}
            data-testid="addConfigButton"
        >
            Add Config
        </Button>
    );
};
