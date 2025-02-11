// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Auth } from 'aws-amplify';
import { useSelector } from 'react-redux';
import { AnyAction, Dispatch } from 'redux';
import { Permissions } from '../constants';
import { selectCurrentDomain } from '../store/domainSlice';
import { PermissionSystemType, selectPermissionSystem } from '../store/permissionSystemSlice';
import { setAppAccessPermission } from '../store/userSlice';

export function useBuildAccessPermission(dispatch: Dispatch<AnyAction>) {
    const permissionSystem = useSelector(selectPermissionSystem);
    const domain = useSelector(selectCurrentDomain);
    const appGlobalPrefix = 'app-global';
    const appDomainPrefix = 'app-' + (domain.customerProfileDomain == '' ? '-' : domain.customerProfileDomain);
    Auth.currentAuthenticatedUser().then(data => {
        const userGroups: string[] | undefined = data.signInUserSession.accessToken.payload['cognito:groups'];
        if (permissionSystem.type === PermissionSystemType.Disabled) {
            dispatch(setAppAccessPermission(Permissions.AdminPermission)); 
            return;
        }
        if (userGroups) {
            let appPermissionValue = 0;
            for (const group of userGroups) {
                if (group.startsWith(appGlobalPrefix) || group.startsWith(appDomainPrefix)) {
                    const permissionStringParts = group.split('/');
                    if (permissionStringParts.length === 2) {
                        appPermissionValue = (appPermissionValue | parseInt(permissionStringParts[1], 16)) >>> 0;
                    }
                }
            }
            dispatch(setAppAccessPermission(appPermissionValue));
        }
    });
}
