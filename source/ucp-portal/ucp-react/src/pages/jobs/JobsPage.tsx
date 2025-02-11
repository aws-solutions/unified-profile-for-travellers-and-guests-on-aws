// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ContentLayout, Header } from '@cloudscape-design/components';
import { useGetJobsQuery } from '../../store/jobsApiSlice.ts';
import { JobsTable } from '../../components/jobs/jobsTable.tsx';

export const JobsPage = () => {
    useGetJobsQuery();

    return (
        <>
            <ContentLayout data-testid="jobsPage" header={<Header variant="h1">Jobs</Header>}>
                <JobsTable />
            </ContentLayout>
        </>
    );
};
