// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Container, ContentLayout, Header, SpaceBetween } from '@cloudscape-design/components';
import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import ActivateRuleSetModal from '../../components/ruleSets/activateRuleSetModal.tsx';
import RuleSetContainer from '../../components/ruleSets/ruleSetContainer.tsx';
import RuleSetSelect from '../../components/ruleSets/ruleSetSelect.tsx';
import { Permissions } from '../../constants.ts';
import { ROUTES } from '../../models/constants.ts';
import { useActivateRuleSetMutation, useListRuleSetsQuery, useSaveRuleSetMutation } from '../../store/ruleSetApiSlice.ts';
import { setSelectedRuleSet } from '../../store/ruleSetSlice';
import { selectAllRuleSets, selectSelectedRuleSet, setEditableRuleSet, setIsEditPage } from '../../store/ruleSetSlice.ts';
import { selectAppAccessPermission } from '../../store/userSlice.ts';
import { rulesToDisplayRules, validateRules } from '../../utils/ruleSetUtil.ts';

export const RulesPage = () => {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    dispatch(setIsEditPage(false));
    const [showActivateModal, setShowActivateModal] = useState<boolean>(false);
    useListRuleSetsQuery();
    const allRuleSets = useSelector(selectAllRuleSets);
    const selectedRuleSet = useSelector(selectSelectedRuleSet);
    const ruleSet = allRuleSets.find(ruleSet => ruleSet.name === selectedRuleSet);
    const rulesOnDisplay = ruleSet ? rulesToDisplayRules(ruleSet.rules) : [];
    const [activateRuleSetTrigger, { isSuccess }] = useActivateRuleSetMutation();
    const [saveRuleSetTrigger, { isSuccess: isSuccessSave }] = useSaveRuleSetMutation();
    const setSelectedRuleSetFunc = (value: string) => {
        dispatch(setSelectedRuleSet(value));
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserCreateRuleSet = (userAppAccess & Permissions.SaveRuleSetPermission) === Permissions.SaveRuleSetPermission;
    const canUserActivateRuleSet = (userAppAccess & Permissions.ActivateRuleSetPermission) === Permissions.ActivateRuleSetPermission;

    return (
        <div data-testid="rulesPage">
            <ContentLayout
                header={
                    <Header
                        data-testid="rulesHeader"
                        variant="h1"
                        actions={
                            <SpaceBetween direction="horizontal" size="m">
                                <Button
                                    onClick={() => {
                                        dispatch(setEditableRuleSet([]));
                                        navigate(`/${ROUTES.RULES_EDIT}`);
                                    }}
                                    disabled={!canUserCreateRuleSet}
                                >
                                    Create New Rule Set
                                </Button>
                                {selectedRuleSet === 'draft' && (
                                    <Button
                                        variant="primary"
                                        onClick={() => {
                                            setShowActivateModal(true);
                                        }}
                                        disabled={!canUserActivateRuleSet}
                                    >
                                        Activate Draft Rule Set
                                    </Button>
                                )}
                                <RuleSetSelect
                                    ruleSetNames={allRuleSets.map(ruleSet => ruleSet.name)}
                                    selectedRuleSetName={selectedRuleSet}
                                    onSetSelected={setSelectedRuleSetFunc}
                                />
                            </SpaceBetween>
                        }
                    >
                        Rule Sets
                    </Header>
                }
                data-testid="rulesContent"
            >
                {rulesOnDisplay && rulesOnDisplay.length > 0 ? (
                    <RuleSetContainer
                        headerText={`Rule Set: ${selectedRuleSet}`}
                        rules={rulesOnDisplay}
                        suppressMatch={false}
                        onSave={saveRuleSetTrigger}
                        onSetSelected={setSelectedRuleSetFunc}
                        isSuccessSave={isSuccessSave}
                        validationFunctionSave={validateRules}
                    />
                ) : (
                    <Container>
                        <Box textAlign="center">No Rule Set Selected</Box>
                    </Container>
                )}
                <ActivateRuleSetModal
                    isVisible={showActivateModal}
                    setVisible={setShowActivateModal}
                    onActivate={activateRuleSetTrigger}
                    isSuccess={isSuccess}
                    isCache={false}
                />
            </ContentLayout>
        </div>
    );
};
