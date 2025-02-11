// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useRunJobMutation } from '../../store/jobsApiSlice';
import { RunGlueJobRequest } from '../../models/job';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';
import { useSelector } from 'react-redux';

interface RunJobButtonProps {
    jobName: string;
}

export const RunJobButton = (props: RunJobButtonProps) => {
    const runJobRequest: RunGlueJobRequest = {
        startJobRq: { jobName: props.jobName },
    };

    const [runJobTrigger] = useRunJobMutation();
    const onRunJobButtonClick = (): void => {
        runJobTrigger({ runJobRequest });
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserRunJobs = (userAppAccess & Permissions.RunGlueJobsPermission) === Permissions.RunGlueJobsPermission;

    return (
        <Button variant="inline-link" onClick={onRunJobButtonClick} data-testid={props.jobName + '_action'} disabled={!canUserRunJobs}>
            Run Job
        </Button>
    );
};
