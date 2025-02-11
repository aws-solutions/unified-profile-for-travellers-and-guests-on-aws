// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act } from '@testing-library/react';
import { renderAppContent } from '../../../test-utils.tsx';

it('Opens domain create page', async () => {
    await act(async () => {
        renderAppContent({
            initialRoute: '/create',
        });
    });
});
