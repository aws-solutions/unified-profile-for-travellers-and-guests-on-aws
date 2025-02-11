// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { Container, ContentLayout, Header } from '@cloudscape-design/components';
import { Route, Routes } from 'react-router-dom';
import Layout from './Layout.tsx';
import { ROUTES } from './models/constants.ts';
import MatchesPage from './pages/aiMatches/matchesPage.tsx';
import { ConfigPage } from './pages/config/ConfigPage.tsx';
import { DomainCreateForm } from './pages/domain/domainCreate/DomainCreateForm.tsx';
import { ErrorsPage } from './pages/errors/ErrorsPage.tsx';
import IntroductionPage from './pages/introduction/IntroductionPage.tsx';
import { JobsPage } from './pages/jobs/JobsPage.tsx';
import PrivacyResults from './pages/privacy/PrivacyResults.tsx';
import PrivacyScreen from './pages/privacy/PrivacyScreen.tsx';
import ProfileDisplay from './pages/profile/profileDisplay.tsx';
import ProfileSearch from './pages/profile/profileSearch.tsx';
import { RulesEditPage } from './pages/rules/RulesEditPage.tsx';
import { RulesPage } from './pages/rules/RulesPage.tsx';
import { CacheEditPage } from './pages/rulesCache/CacheEditPage.tsx';
import { CachePage } from './pages/rulesCache/CachePage.tsx';
import { DomainSettingsPage } from './pages/settings/DomainSettingsPage.tsx';

export const AppRoutes = () => (
    <Routes>
        <Route path={'/*'} element={<Layout />}>
            <Route index element={<IntroductionPage />} />
            <Route path={ROUTES.CREATE} element={<DomainCreateForm />} />
            <Route path={ROUTES.SETTINGS} element={<DomainSettingsPage />} />
            <Route path={ROUTES.ERRORS} element={<ErrorsPage />} />
            <Route path={ROUTES.JOBS} element={<JobsPage />} />
            <Route path={ROUTES.CONFIG} element={<ConfigPage />} />
            <Route path={ROUTES.PROFILE} element={<ProfileSearch />} />
            <Route path={ROUTES.PROFILE_PROFILE_ID} element={<ProfileDisplay />} />
            <Route path={ROUTES.IDENTITY_RESOLUTION} element={<IntroductionPage />} />
            <Route path={ROUTES.DATA_QUALITY} element={<IntroductionPage />} />
            <Route path={ROUTES.SETUP} element={<IntroductionPage />} />
            <Route path={ROUTES.PRIVACY} element={<PrivacyScreen />} />
            <Route path={ROUTES.PRIVACY_RESULTS} element={<PrivacyResults />} />
            <Route path={ROUTES.RULES} element={<RulesPage />} />
            <Route path={ROUTES.RULES_EDIT} element={<RulesEditPage />} />
            <Route path={ROUTES.CACHE} element={<CachePage />} />
            <Route path={ROUTES.CACHE_EDIT} element={<CacheEditPage />} />
            <Route path={ROUTES.AI_MATCHES} element={<MatchesPage />} />

            {/* Add more child routes that use the same Layout here */}

            <Route
                path="*"
                element={
                    <ContentLayout header={<Header>Error</Header>}>
                        <Container header={<Header>Page not found 😿</Header>}></Container>
                    </ContentLayout>
                }
            />
        </Route>

        {/* Add another set of routes with a different layout here */}
    </Routes>
);
