// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Header, Link, TableProps } from '@cloudscape-design/components';
import BaseCRUDTable from '../base/baseCrudTable';
import { useGetJobsQuery } from '../../store/jobsApiSlice';
import { useEffect, useState } from 'react';
import { Job } from '../../models/config';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import { AWS_GLUE_PREFIX, DATE_FORMAT, INVALID_DATE } from '../../models/constants';
import { RunAllJobsButton } from './runAllJobsButton';
import { RunJobButton } from './runJobButton';
dayjs.extend(utc);

export const JobsTable = () => {
    const columnDefs: TableProps.ColumnDefinition<Job>[] = [
        {
            id: 'jobName',
            header: 'Job Name',
            cell: job => (
                <Link external href={`${AWS_GLUE_PREFIX}${job.jobName}/runs`}>
                    {job.jobName}
                </Link>
            ),
            ariaLabel: () => 'jobName',
        },
        {
            id: 'lastRun',
            header: 'Last Run',
            cell: job => {
                const date = dayjs.utc(job.lastRunTime).format(DATE_FORMAT);
                if (date === INVALID_DATE) {
                    return 'Never Ran';
                }
                return date;
            },
            ariaLabel: () => 'lastRun',
        },
        {
            id: 'status',
            header: 'Status',
            cell: job => job.status,
            ariaLabel: () => 'status',
        },
        {
            id: 'action',
            header: 'Action',
            cell: job => <RunJobButton jobName={job.jobName}></RunJobButton>,
            ariaLabel: () => 'action',
        },
    ];

    const { data: jobData, isLoading: getJobsLoading } = useGetJobsQuery();
    const [tableItems, setTableItems] = useState<Job[]>([]);
    useEffect(() => {
        if (jobData !== undefined) {
            setTableItems(jobData);
        }
    }, [jobData]);

    const header = (
        <Header variant="h2" actions={<RunAllJobsButton />}>
            Glue Jobs
        </Header>
    );

    return (
        <BaseCRUDTable<Job>
            data-testid="jobs-table"
            columnDefinitions={columnDefs}
            items={tableItems}
            header={header}
            preferences={{}}
            isLoading={getJobsLoading}
            tableVariant="container"
            isFilterable={false}
        />
    );
};
