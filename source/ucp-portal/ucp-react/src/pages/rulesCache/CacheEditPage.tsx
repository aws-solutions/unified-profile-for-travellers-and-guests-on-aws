// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ContentLayout, Header } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import RuleSetContainer from '../../components/ruleSets/ruleSetContainer.tsx';
import { selectEditableRuleSet, setIsEditPage, setSelectedRuleSetCache } from '../../store/ruleSetSlice.ts';
import { useGetPortalConfigQuery } from '../../store/configApiSlice.ts';
import { useEffect } from 'react';
import { useListRuleSetsCacheQuery, useSaveCacheRuleSetMutation } from '../../store/ruleSetApiSlice.ts';
import { validateCacheRules } from '../../utils/ruleSetUtil.ts';

export const CacheEditPage = () => {
    const dispatch = useDispatch();
    const editableRuleSet = useSelector(selectEditableRuleSet);
    useListRuleSetsCacheQuery();
    dispatch(setIsEditPage(true));
    useGetPortalConfigQuery();
    const [saveRuleSetTrigger, { isSuccess: isSuccessSave }] = useSaveCacheRuleSetMutation();
    const setSelectedRuleSetFunc = (value: string) => {
        dispatch(setSelectedRuleSetCache(value));
    };

    return (
        <div data-testid="rulesEditPage">
            <ContentLayout
                header={
                    <Header data-testid="rulesEditHeader" variant="h1" description="Leaving this page will lose all changes made">
                        Edit Cache Skip Conditions
                    </Header>
                }
                data-testid="rulesEditContent"
            >
                <RuleSetContainer headerText={`Rule Set`} 
                        rules={editableRuleSet} 
                        suppressMatch={true}
                        onSave={saveRuleSetTrigger}
                        onSetSelected={setSelectedRuleSetFunc}
                        isSuccessSave={isSuccessSave}
                        validationFunctionSave={validateCacheRules}
                    />
            </ContentLayout>
        </div>
    );
};
