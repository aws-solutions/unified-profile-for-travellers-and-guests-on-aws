// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act } from '@testing-library/react';
import { renderAppContent } from '../../test-utils.tsx';
import { ROUTES } from '../../../models/constants.ts';

it('Opens cache page', async () => {
    await act(async () => {
        renderAppContent({
            initialRoute: `/${ROUTES.CACHE}`,
        });
    });
});
