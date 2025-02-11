// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ApiEndpoints, solutionApi } from './solutionApi.ts';
import { ConfigResponse, Job } from '../models/config.ts';
import { RunGlueJobRequest } from '../models/job.ts';

export const jobsApiSlice = solutionApi.injectEndpoints({
    endpoints: builder => ({
        getJobs: builder.query<Job[], void>({
            query: () => ApiEndpoints.JOB,
            providesTags: ['Job'],
            transformResponse: (response: ConfigResponse) => {
                return response.awsResources.jobs;
            },
        }),
        runJob: builder.mutation<ConfigResponse, { runJobRequest: RunGlueJobRequest }>({
            query: ({ runJobRequest }) => ({
                url: ApiEndpoints.JOB,
                method: 'POST',
                body: runJobRequest,
            }),
            invalidatesTags: ['Job'],
        }),
        runAllJobs: builder.mutation<ConfigResponse, void>({
            query: () => ({
                url: ApiEndpoints.JOB,
                method: 'POST',
                body: {
                    startJobRq: {
                        jobName: 'run-all-jobs',
                    },
                },
            }),
            invalidatesTags: ['Job'],
        }),
    }),
});

export const { useGetJobsQuery, useRunJobMutation, useRunAllJobsMutation } = jobsApiSlice;
