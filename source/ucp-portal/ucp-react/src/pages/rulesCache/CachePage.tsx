// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Container, ContentLayout, Header, SpaceBetween } from '@cloudscape-design/components';
import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate } from 'react-router-dom';
import ActivateRuleSetModal from '../../components/ruleSets/activateRuleSetModal.tsx';
import RuleSetContainer from '../../components/ruleSets/ruleSetContainer.tsx';
import RuleSetSelect from '../../components/ruleSets/ruleSetSelect.tsx';
import { Permissions } from '../../constants';
import { ROUTES } from '../../models/constants.ts';
import { useActivateCacheRuleSetMutation, useListRuleSetsCacheQuery, useSaveCacheRuleSetMutation } from '../../store/ruleSetApiSlice.ts';
import {
    selectAllRuleSetsCache,
    selectSelectedRuleSetCache,
    setEditableRuleSet,
    setIsEditPage,
    setSelectedRuleSetCache,
} from '../../store/ruleSetSlice.ts';
import { selectAppAccessPermission } from '../../store/userSlice';
import { rulesToDisplayRules, validateCacheRules } from '../../utils/ruleSetUtil.ts';

export const CachePage = () => {
    const dispatch = useDispatch();
    const navigate = useNavigate();
    dispatch(setIsEditPage(false));
    const [showActivateModal, setShowActivateModal] = useState<boolean>(false);
    useListRuleSetsCacheQuery();
    const allRuleSets = useSelector(selectAllRuleSetsCache);
    const selectedRuleSet = useSelector(selectSelectedRuleSetCache);
    const ruleSet = allRuleSets.find(ruleSet => ruleSet.name === selectedRuleSet);
    const rulesOnDisplay = ruleSet ? rulesToDisplayRules(ruleSet.rules) : [];
    const [activateRuleSetTrigger, { isSuccess }] = useActivateCacheRuleSetMutation();
    const location = useLocation();
    const [saveRuleSetTrigger, { isSuccess: isSuccessSave }] = useSaveCacheRuleSetMutation();
    const setSelectedRuleSetFunc = (value: string) => {
        dispatch(setSelectedRuleSetCache(value));
    };
    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserCreateRuleSetCache = (userAppAccess & Permissions.SaveRuleSetPermission) === Permissions.SaveRuleSetPermission;
    const canUserActivateRuleSetCache = (userAppAccess & Permissions.ActivateRuleSetPermission) === Permissions.ActivateRuleSetPermission;

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
                                        navigate(`${location.pathname}/${ROUTES.EDIT}`);
                                    }}
                                    disabled={!canUserCreateRuleSetCache}
                                >
                                    Create New Cache Set
                                </Button>
                                {selectedRuleSet === 'draft' && (
                                    <Button
                                        variant="primary"
                                        onClick={() => {
                                            setShowActivateModal(true);
                                        }}
                                        disabled={!canUserActivateRuleSetCache}
                                    >
                                        Activate Draft Cache Set
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
                        Rule Sets For Cache
                    </Header>
                }
                data-testid="rulesContent"
            >
                {rulesOnDisplay && rulesOnDisplay.length > 0 ? (
                    <RuleSetContainer
                        headerText={`Rule Set: ${selectedRuleSet}`}
                        rules={rulesOnDisplay}
                        suppressMatch={true}
                        onSave={saveRuleSetTrigger}
                        onSetSelected={setSelectedRuleSetFunc}
                        isSuccessSave={isSuccessSave}
                        validationFunctionSave={validateCacheRules}
                    />
                ) : (
                    <Container>
                        <Box textAlign="center">No Cache Set Selected</Box>
                    </Container>
                )}
                <ActivateRuleSetModal
                    isVisible={showActivateModal}
                    setVisible={setShowActivateModal}
                    onActivate={activateRuleSetTrigger}
                    isSuccess={isSuccess}
                    isCache={true}
                />
            </ContentLayout>
        </div>
    );
};
