// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act } from '@testing-library/react';
import { ROUTES } from '../../../models/constants.ts';
import { renderAppContent } from '../../test-utils.tsx';

it('Opens cache edit page', async () => {
    await act(async () => {
        renderAppContent({
            initialRoute: `/${ROUTES.CACHE_EDIT}`,
            preloadedState: {
                ruleSet: {
                    allRuleSets: [],
                    allRuleSetsCache: [],
                    editableRuleSet: [],
                    selectedRuleSet: 'active',
                    selectedRuleSetCache: 'active',
                    isEditPage: true,
                    profileMappings: [],
                },
            },
        });
    });
});
