// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, within, screen, fireEvent } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { http } from 'msw';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { ok } from '../../../mocks/handlers.ts';

const testJob = {
    jobName: 'test_job',
    lastRunTime: new Date(),
    status: 'not_running',
};

it('opens jobs page', async () => {
    await act(async () => {
        renderAppContent({
            initialRoute: '/jobs',
        });
    });
});

it('opens jobs page and loads in one job', async () => {
    server.use(
        http.get(
            MOCK_SERVER_URL + ApiEndpoints.JOB,
            async () =>
                await ok({
                    awsResources: {
                        jobs: [testJob],
                    },
                }),
        ),
    );

    await act(async () => {
        renderAppContent({
            initialRoute: '/jobs',
            preloadedState: {
                user: {
                    appAccessPermission: 2 ** 10,
                },
            },
        });
    });

    const withinMain = within(screen.getByTestId('main-content'));
    const jobTable = await withinMain.findByTestId('jobs-table');
    expect(jobTable).toBeInTheDocument();
    const row = await withinMain.findByTestId('test_job_action');
    expect(row).toBeInTheDocument();

    await act(async () => {
        row.click();
    });

    const runAllButton = await screen.getByText('Run All Jobs');
    expect(runAllButton).toBeInTheDocument();
    await act(async () => {
        fireEvent.click(runAllButton);
    });
});
