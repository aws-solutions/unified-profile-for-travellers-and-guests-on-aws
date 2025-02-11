// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { SideNavigation, SideNavigationProps } from '@cloudscape-design/components';
import { Auth } from 'aws-amplify';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSelector } from 'react-redux';
import { NavigateFunction, useLocation, useNavigate } from 'react-router-dom';
import { ROUTES } from '../../models/constants';
import { selectCurrentDomain } from '../../store/domainSlice';

export default function SideNavigationBar() {
    const navigate: NavigateFunction = useNavigate();
    const { t } = useTranslation();
    const [activeHref, setActiveHref] = useState('/');
    const currentDomainName = useSelector(selectCurrentDomain).customerProfileDomain;
    const domainSpecificSection: SideNavigationProps.Item[] =
        currentDomainName === ''
            ? []
            : [
                  {
                      type: 'section-group',
                      title: `${currentDomainName}`,
                      items: [
                          { type: 'link', text: 'Profiles', href: `/${ROUTES.PROFILE}` },
                          { type: 'link', text: 'Domain Settings', href: `/${ROUTES.SETTINGS}` },
                          { type: 'link', text: 'Jobs', href: `/${ROUTES.JOBS}` },
                          { type: 'link', text: 'Privacy', href: `/${ROUTES.PRIVACY}` },
                          { type: 'link', text: 'Rules', href: `/${ROUTES.RULES}` },
                          { type: 'link', text: 'Cache Management', href: `/${ROUTES.CACHE}` },
                          { type: 'link', text: 'AI Matches', href: `/${ROUTES.AI_MATCHES}` },
                      ],
                  },
              ];

    const navigationItems: SideNavigationProps.Item[] = [
        { type: 'link', text: 'Welcome', href: '/' },
        { type: 'divider' },
        {
            type: 'section-group',
            title: 'Overview',
            items: [
                { type: 'link', text: 'Errors', href: `/${ROUTES.ERRORS}` },
                { type: 'link', text: 'Config', href: `/${ROUTES.CONFIG}` },
            ],
        },
        ...domainSpecificSection,
        { type: 'divider' },
        { type: 'link', text: 'Logout', href: '/logout' },
    ];

    // follow the given router link and update the store with active path
    const handleFollow = useCallback(
        (event: Readonly<CustomEvent>): void => {
            if (event.detail.external || !event.detail.href) return;
            if (event.detail.href === '/logout') {
                Auth.signOut();
                event.preventDefault();
                navigate('/');
            } else {
                event.preventDefault();
                const path = event.detail.href;
                navigate(path);
            }
        },
        [navigate],
    );

    const location = useLocation();
    useEffect(() => {
        setActiveHref(location.pathname);
    }, [location]);

    const navHeader: SideNavigationProps.Header = {
        href: '/',
        text: t('sidenav.header'),
    };

    return (
        <SideNavigation
            header={navHeader}
            activeHref={activeHref}
            onFollow={handleFollow}
            items={navigationItems}
            data-testid={'sideNav'}
        />
    );
}
