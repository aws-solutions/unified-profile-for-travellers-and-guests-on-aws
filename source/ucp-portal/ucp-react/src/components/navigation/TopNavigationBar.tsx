// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { TopNavigation, TopNavigationProps } from '@cloudscape-design/components';
import { DomainSelector } from '../DomainSelector.tsx';
import { ROOT_TITLE, ROUTES } from '../../models/constants.ts';
import { useNavigate } from 'react-router-dom';
import { useGetAllDomainsQuery } from '../../store/domainApiSlice.ts';

export default function TopNavigationBar() {
    const { topNavProps } = DomainSelector();
    useGetAllDomainsQuery();
    const navigate = useNavigate();
    const solutionIdentity: TopNavigationProps.Identity = {
        href: '/',
        title: ROOT_TITLE,
    };

    const i18nStrings: TopNavigationProps.I18nStrings = {
        overflowMenuTitleText: 'All',
        overflowMenuTriggerText: 'More',
    };

    const utilities: TopNavigationProps.Utility[] = [
        {
            type: 'button',
            text: 'Create Domain',
            iconName: 'add-plus',
            onClick: () => {
                navigate(ROUTES.CREATE);
            },
            ariaLabel: 'createDomainTopNav',
        },
    ];
    return <TopNavigation identity={solutionIdentity} i18nStrings={i18nStrings} utilities={[...topNavProps, ...utilities]} />;
}
