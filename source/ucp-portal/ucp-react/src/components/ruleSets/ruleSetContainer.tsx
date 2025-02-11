// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Container, Header, SpaceBetween } from '@cloudscape-design/components';
import RuleContainer from './ruleContainer';
import { useDispatch, useSelector } from 'react-redux';
import { addRule, selectIsEditPage } from '../../store/ruleSetSlice';
import { IconName } from '../../models/iconName';
import CancelEditRuleSetButton from './cancelEditRuleSetButton';
import SaveRuleSetButton from './saveRuleSetButton';
import EditRuleSetButton from './editRuleSetButton';
import { DisplayRule, SaveRuleSetRequest } from '../../models/ruleSet';
import { ValidationResult } from '../../utils/ruleSetUtil';

interface RuleSetContainerProps {
    headerText: string;
    headerDescription?: string;
    rules: DisplayRule[];
    suppressMatch: boolean;
    onSave: (rq: SaveRuleSetRequest) => void;
    onSetSelected: (value: string) => void;
    isSuccessSave: boolean;
    validationFunctionSave: (rules: DisplayRule[]) => ValidationResult;
}

export default function RuleSetContainer({ headerText, headerDescription, rules, suppressMatch, onSave, onSetSelected, isSuccessSave, validationFunctionSave }: RuleSetContainerProps) {
    const dispatch = useDispatch();
    const onAddRuleButtonClick = () => {
        dispatch(addRule());
    };

    const isEditPage = useSelector(selectIsEditPage);

    const headerActions = isEditPage ? (
        <SpaceBetween direction="horizontal" size="s">
            <CancelEditRuleSetButton />
            <SaveRuleSetButton 
                onSave={onSave} 
                onSetSelected={onSetSelected} 
                isSuccess={isSuccessSave} 
                validationFunction={validationFunctionSave}/>
        </SpaceBetween>
    ) : (
        <EditRuleSetButton rules={rules} />
    );

    const header = (
        <Header description={headerDescription} actions={headerActions}>
            {headerText}
        </Header>
    );

    return (
        <Container header={header}>
            <SpaceBetween size="s">
                {rules &&
                    rules.map(rule => (
                        <RuleContainer ruleIndex={rule.index} rule={rule} key={rule.name + '_' + rule.index} suppressMatch={suppressMatch} />
                    ))}
                {isEditPage && (
                    <Box textAlign="center" key="add_rule_button">
                        <Button iconName={IconName.ADD} onClick={onAddRuleButtonClick}>
                            Add New Rule
                        </Button>
                    </Box>
                )}
            </SpaceBetween>
        </Container>
    );
}
