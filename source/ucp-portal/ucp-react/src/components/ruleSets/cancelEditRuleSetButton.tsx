// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useNavigate } from 'react-router-dom';
import { ROUTES } from '../../models/constants';

export default function CancelEditRuleSetButton() {
    const navigate = useNavigate();
    const routeSuffix = `/${ROUTES.EDIT}`
    const currentUrl = location.pathname;
    let targetUrl = currentUrl
    if (currentUrl.endsWith(routeSuffix)) {
        targetUrl = currentUrl.slice(0, -routeSuffix.length);
    }
    return (
        <Button
            onClick={() => {
                navigate(targetUrl);
            }}
        >
            Cancel Edit
        </Button>
    );
}
