// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ButtonDropdownProps, NonCancelableCustomEvent, TopNavigationProps } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { Domain } from '../models/config.ts';
import { ROUTES } from '../models/constants.ts';
import { useLazyGetDomainQuery } from '../store/domainApiSlice.ts';
import { selectAllDomains, setSelectedDomain } from '../store/domainSlice.ts';
import { resetState as resetProfileSliceState } from '../store/profileSlice.ts';
import { getDomainValueStorage } from '../utils/localStorageUtil.ts';

export const DomainSelector = () => {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const domainName = getDomainValueStorage() || 'Select a domain to get started';
    const domains = useSelector(selectAllDomains);
    const [triggerGetDomainQuery] = useLazyGetDomainQuery();

    const handleSelectChange = ({ detail }: NonCancelableCustomEvent<ButtonDropdownProps.ItemClickDetails>): void => {
        const selectedDomain = domains.find(domain => domain.customerProfileDomain === detail.id);
        if (selectedDomain !== undefined) {
            dispatch(resetProfileSliceState());
            dispatch(setSelectedDomain(selectedDomain));
        }
        triggerGetDomainQuery(detail.id);
        navigate(ROUTES.SETTINGS);
    };

    const dropdownItems = domains.map((item: Domain) => ({
        text: item.customerProfileDomain,
        id: item.customerProfileDomain,
        disabled: false,
    }));
    const topNavProps: TopNavigationProps.Utility[] = [
        {
            type: 'menu-dropdown',
            text: domainName,
            items: dropdownItems,
            onItemClick: handleSelectChange,
            ariaLabel: 'domainSelector',
        },
    ];

    return { topNavProps };
};
