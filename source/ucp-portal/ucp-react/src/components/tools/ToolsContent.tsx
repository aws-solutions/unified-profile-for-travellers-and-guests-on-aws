// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Tabs } from '@cloudscape-design/components';
import { ReactNode } from 'react';
import { Route, Routes } from 'react-router-dom';
import { ROUTES } from '../../models/constants.ts';
import {
    AIMatchesHelp,
    CacheManagementHelp,
    ConfigHelp,
    DomainSettingsHelp,
    ErrorsHelp,
    HomeHelp,
    JobsHelp,
    NoHelp,
    PrivacyHelp,
    ProfileDisplayHelp,
    ProfileSearchHelp,
    RulesHelp,
} from './HelpPanelContent.tsx';

/**
 * This wrapper component distinguishes between pages that have
 * - both tutorial and help content
 * - only help content
 * to avoid displaying an empty tutorial section to the user.
 * If your solution has tutorial and help on all pages,
 * or no tutorials at all, you can inline the respective component and delete this wrapper.
 */
export const ToolsContent = () => {
    const TutorialsAndHelp = ({ tutorial, help }: { tutorial: ReactNode; help: ReactNode }) => (
        <Tabs
            tabs={[
                {
                    label: 'Tutorial',
                    id: 'tutorial-tab',
                    content: tutorial,
                },
                {
                    label: 'Help',
                    id: 'help',
                    content: help,
                },
            ]}
        />
    );

    return (
        <>
            <Routes>
                <Route path="/" element={<HomeHelp />} />
                <Route path={`/${ROUTES.AI_MATCHES}`} element={<AIMatchesHelp></AIMatchesHelp>} />
                <Route path={`/${ROUTES.CACHE}`} element={<CacheManagementHelp></CacheManagementHelp>} />
                <Route path={`/${ROUTES.CACHE_EDIT}`} element={<CacheManagementHelp></CacheManagementHelp>} />
                <Route path={`/${ROUTES.RULES}`} element={<RulesHelp></RulesHelp>} />
                <Route path={`/${ROUTES.RULES_EDIT}`} element={<RulesHelp></RulesHelp>} />
                <Route path={`/${ROUTES.PRIVACY}`} element={<PrivacyHelp></PrivacyHelp>} />
                <Route path={`/${ROUTES.PRIVACY_RESULTS}`} element={<PrivacyHelp></PrivacyHelp>} />
                <Route path={`/${ROUTES.JOBS}`} element={<JobsHelp></JobsHelp>} />
                <Route path={`/${ROUTES.SETTINGS}`} element={<DomainSettingsHelp></DomainSettingsHelp>} />
                <Route path={`/${ROUTES.PROFILE}`} element={<ProfileSearchHelp></ProfileSearchHelp>} />
                <Route path={`/${ROUTES.PROFILE_PROFILE_ID}`} element={<ProfileDisplayHelp></ProfileDisplayHelp>} />
                <Route path={`/${ROUTES.CONFIG}`} element={<ConfigHelp></ConfigHelp>} />
                <Route path={`/${ROUTES.ERRORS}`} element={<ErrorsHelp></ErrorsHelp>} />
                <Route path="*" element={<NoHelp></NoHelp>} />
            </Routes>
        </>
    );
};
