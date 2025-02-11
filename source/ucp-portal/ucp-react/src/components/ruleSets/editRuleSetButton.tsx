// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useLocation, useNavigate } from 'react-router-dom';
import { ROUTES } from '../../models/constants';
import { useDispatch, useSelector } from 'react-redux';
import { setEditableRuleSet } from '../../store/ruleSetSlice';
import { DisplayRule } from '../../models/ruleSet';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

interface EditRuleSetButtonProps {
    rules: DisplayRule[];
}

export default function EditRuleSetButton({ rules }: EditRuleSetButtonProps) {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const location = useLocation();
    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserEditRuleSet = (userAppAccess & Permissions.SaveRuleSetPermission) === Permissions.SaveRuleSetPermission;
    return (
        <Button
            onClick={() => {
                dispatch(setEditableRuleSet(rules));
                navigate(`${location.pathname}/${ROUTES.EDIT}`);
            }}
            disabled={!canUserEditRuleSet}
        >
            Edit this rule set
        </Button>
    );
}
