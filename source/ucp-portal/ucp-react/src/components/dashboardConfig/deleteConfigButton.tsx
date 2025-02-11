// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { IconName } from '../../models/iconName';
import { HyperLinkMapping } from '../../models/config';
import { usePostPortalConfigMutation } from '../../store/configApiSlice';
import { useSelector } from 'react-redux';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

interface DeleteConfigButtonProps {
    configData: HyperLinkMapping[] | undefined;
    configToDelete: HyperLinkMapping;
}

export const DeleteConfigButton = (props: DeleteConfigButtonProps) => {
    const filteredConfigData = props.configData?.filter(config => JSON.stringify(config) !== JSON.stringify(props.configToDelete));

    const [postPortalConfigTrigger] = usePostPortalConfigMutation();

    const onDeleteConfigButtonClick = (): void => {
        if (filteredConfigData) {
            postPortalConfigTrigger(filteredConfigData);
        }
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserDeleteConfig = (userAppAccess & Permissions.SaveHyperlinkPermission) === Permissions.SaveHyperlinkPermission;

    return <Button variant="icon" onClick={onDeleteConfigButtonClick} iconName={IconName.REMOVE} disabled={!canUserDeleteConfig}/>;
};
