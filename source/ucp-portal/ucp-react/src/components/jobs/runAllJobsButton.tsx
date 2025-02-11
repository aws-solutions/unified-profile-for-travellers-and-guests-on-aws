// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useRunAllJobsMutation } from '../../store/jobsApiSlice';
import { useSelector } from 'react-redux';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

export const RunAllJobsButton = () => {
    const [runAllJobsTrigger] = useRunAllJobsMutation();
    const onRunAllJobsButtonClick = (): void => {
        runAllJobsTrigger();
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserRunJobs = (userAppAccess & Permissions.RunGlueJobsPermission) === Permissions.RunGlueJobsPermission;

    return (
        <Button variant="primary" onClick={onRunAllJobsButtonClick} disabled={!canUserRunJobs}>
            Run All Jobs
        </Button>
    );
};
