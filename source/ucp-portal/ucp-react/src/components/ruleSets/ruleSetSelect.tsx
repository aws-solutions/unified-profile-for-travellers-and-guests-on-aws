// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useDispatch } from 'react-redux';
import { Select } from '@cloudscape-design/components';

interface RuleSetSelectProps {
    ruleSetNames: string[];
    selectedRuleSetName: string;
    onSetSelected: (value: string) => void;
}

export default function RuleSetSelect({ ruleSetNames, selectedRuleSetName, onSetSelected }: RuleSetSelectProps) {
    const dispatch = useDispatch();
    const options = ruleSetNames.map(name => ({ label: name, value: name }));
    const selectedRuleSetAsOption = options.find(option => option.value === selectedRuleSetName);

    return (
        <Select
            selectedOption={selectedRuleSetAsOption ? selectedRuleSetAsOption : null}
            onChange={({ detail }) => onSetSelected(detail.selectedOption.value!)}
            options={options}
            placeholder="Select a rule set"
            empty="No rule sets found"
        />
    );
}
