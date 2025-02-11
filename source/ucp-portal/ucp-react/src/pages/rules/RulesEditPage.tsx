// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ContentLayout, Header } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import RuleSetContainer from '../../components/ruleSets/ruleSetContainer.tsx';
import { selectEditableRuleSet, setIsEditPage, setSelectedRuleSet } from '../../store/ruleSetSlice.ts';
import { useGetPortalConfigQuery } from '../../store/configApiSlice.ts';
import { useListRuleSetsQuery, useSaveRuleSetMutation } from '../../store/ruleSetApiSlice.ts';
import { validateRules } from '../../utils/ruleSetUtil.ts';

export const RulesEditPage = () => {
    const dispatch = useDispatch();
    const editableRuleSet = useSelector(selectEditableRuleSet);
    useListRuleSetsQuery();
    dispatch(setIsEditPage(true));
    useGetPortalConfigQuery();
    const [saveRuleSetTrigger, { isSuccess: isSuccessSave }] = useSaveRuleSetMutation();
    const setSelectedRuleSetFunc = (value: string) => {
        dispatch(setSelectedRuleSet(value));
    };

    return (
        <div data-testid="rulesEditPage">
            <ContentLayout
                header={
                    <Header data-testid="rulesEditHeader" variant="h1" description="Leaving this page will lose all changes made">
                        Edit Rule Set
                    </Header>
                }
                data-testid="rulesEditContent"
            >
                <RuleSetContainer
                    headerText={'Rule Set'}
                    rules={editableRuleSet}
                    suppressMatch={false}
                    onSave={saveRuleSetTrigger}
                    onSetSelected={setSelectedRuleSetFunc}
                    isSuccessSave={isSuccessSave}
                    validationFunctionSave={validateRules}
                />
            </ContentLayout>
        </div>
    );
};
