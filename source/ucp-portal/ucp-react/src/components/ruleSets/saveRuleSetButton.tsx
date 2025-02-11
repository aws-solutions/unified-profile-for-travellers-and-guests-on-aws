// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { v4 } from 'uuid';
import { Permissions } from '../../constants';
import { ROUTES } from '../../models/constants';
import { DisplayRule, SaveRuleSetRequest } from '../../models/ruleSet';
import { NotificationPayload, addNotification } from '../../store/notificationsSlice';
import { selectEditableRuleSet } from '../../store/ruleSetSlice';
import { selectAppAccessPermission } from '../../store/userSlice';
import { ValidationResult, displayRulesToRules } from '../../utils/ruleSetUtil';

interface SaveRuleSetModelProps {
    onSave: (rq: SaveRuleSetRequest) => void;
    onSetSelected: (value: string) => void;
    isSuccess: boolean;
    validationFunction: (rules: DisplayRule[]) => ValidationResult;
}

export default function SaveRuleSetButton({ onSave, onSetSelected, isSuccess, validationFunction }: SaveRuleSetModelProps) {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const routeSuffix = `/${ROUTES.EDIT}`;
    const currentUrl = location.pathname;
    let targetUrl = currentUrl;
    if (currentUrl.endsWith(routeSuffix)) {
        targetUrl = currentUrl.slice(0, -routeSuffix.length);
    }
    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserSaveRuleSet = (userAppAccess & Permissions.SaveRuleSetPermission) === Permissions.SaveRuleSetPermission;

    useEffect(() => {
        if (isSuccess) {
            navigate(targetUrl);
        }
    }, [isSuccess]);

    const displayRules = useSelector(selectEditableRuleSet);
    const onSaveRuleSetButtonClick = () => {
        const validationResult = validationFunction(displayRules);
        if (validationResult.isValid) {
            const rules = displayRulesToRules(displayRules);
            const saveRuleSetRequest: SaveRuleSetRequest = { rules };
            onSave(saveRuleSetRequest);
            onSetSelected('draft');
        } else {
            const notification: NotificationPayload = {
                id: v4(),
                type: 'error',
                content: 'Rule set is not valid: ' + validationResult.issue,
            };
            dispatch(addNotification(notification));
        }
    };
    return (
        <Button variant="primary" onClick={onSaveRuleSetButtonClick} disabled={!canUserSaveRuleSet}>
            Save Rule Set to Draft
        </Button>
    );
}
