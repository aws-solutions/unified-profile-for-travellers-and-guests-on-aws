// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useSelector } from 'react-redux';
import { Box, Button, Header, Icon, SpaceBetween, Textarea, Toggle, Popover, Slider, Input } from '@cloudscape-design/components';
import { selectIsAutoMatchActive, selectIsPromptActive, selectMatchValue, selectPromptValue } from '../../store/domainSlice';
import { SetStateAction, useState } from 'react';
import { useGetPromptConfigQuery, useSavePromptConfigMutation } from '../../store/domainApiSlice';
import { IconName } from '../../models/iconName';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

function ConfigToggle(
    toggleText: string,
    editedIsActive: boolean,
    setSummaryGeneration: (value: SetStateAction<boolean>) => void,
    canUserEditPrompt: boolean,
    label: string,
) {
    return (
        <SpaceBetween direction="horizontal" size="s">
            <Box padding={{ right: 'xxxl' }}>{toggleText}</Box>
            <Box>Off</Box>
            <Toggle
                checked={editedIsActive}
                onChange={({ detail }) => {
                    setSummaryGeneration(detail.checked);
                }}
                ariaLabel={label}
                disabled={!canUserEditPrompt}
            ></Toggle>
            <Box>On</Box>
        </SpaceBetween>
    );
}

export default function DomainConfiguration() {
    useGetPromptConfigQuery();
    const promptValue = useSelector(selectPromptValue);
    const [editedPromptValue, setEditedPromptValue] = useState(promptValue);

    const isSummaryGenActive = useSelector(selectIsPromptActive);
    const [editSummaryGeneration, setSummaryGeneration] = useState(isSummaryGenActive);

    const matchValue = useSelector(selectMatchValue);
    const [editedMatchValue, setEditedMatchValue] = useState(Number(matchValue));

    const isAutoMatchActive = useSelector(selectIsAutoMatchActive);
    const [editIsMatchActive, setIsMatchActive] = useState(isAutoMatchActive);

    const [savePromptTrigger] = useSavePromptConfigMutation();

    const savePromptConfig = () => {
        savePromptTrigger({
            domainSetting: {
                promptConfig: {
                    isActive: editSummaryGeneration,
                    value: editedPromptValue,
                },
                matchConfig: {
                    isActive: editIsMatchActive,
                    value: String(editedMatchValue),
                },
            },
        });
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserEditPrompt = (userAppAccess & Permissions.ConfigureGenAiPermission) === Permissions.ConfigureGenAiPermission;
    const hasPromptConfigChanged = promptValue !== editedPromptValue || isSummaryGenActive !== editSummaryGeneration;
    const hasMatchConfigChanged = Number(matchValue) !== editedMatchValue || isAutoMatchActive !== editIsMatchActive;
    const hasConfigChanged = hasPromptConfigChanged || hasMatchConfigChanged;

    const actionButton = (
        <Button onClick={() => savePromptConfig()} disabled={!(hasConfigChanged && canUserEditPrompt)} ariaLabel="savePromptConfig">
            Save
        </Button>
    );

    const settingsHeader = (
        <Header variant="h1" actions={actionButton}>
            Domain settings
        </Header>
    );

    const genaiHeader = (
        <Header variant="h3">
            Generative-AI
            <Icon name={IconName.GEN_AI} /> Prompt Configuration{' '}
            <Popover dismissButton={false} position="top" size="small" triggerType="text" content="Powered by Anthropic Sonnet v3">
                <Icon name={IconName.STATUS_INFO} />
            </Popover>
        </Header>
    );

    const genaiTextbox = (
        <Textarea
            value={editedPromptValue}
            onChange={({ detail }) => {
                setEditedPromptValue(detail.value);
            }}
            placeholder="Enter a prompt for generating profile summary"
            ariaLabel="summaryGenTextbox"
            disabled={!canUserEditPrompt}
        />
    );

    return (
        <>
            <SpaceBetween direction="vertical" size="s">
                {settingsHeader}
                {genaiHeader}
                {ConfigToggle(
                    'Is summary generation active?',
                    editSummaryGeneration,
                    setSummaryGeneration,
                    canUserEditPrompt,
                    'summaryGenToggle',
                )}
                {editSummaryGeneration && genaiTextbox}
                <Header variant="h3">
                    Auto Merging Threshold{' '}
                    <Popover
                        dismissButton={false}
                        position="top"
                        size="small"
                        triggerType="text"
                        content="Turning this on will auto-merge matches with a confidence score above the threshold"
                    >
                        <Icon name={IconName.STATUS_INFO} />
                    </Popover>
                </Header>
                {ConfigToggle('Is auto merging active?', editIsMatchActive, setIsMatchActive, canUserEditPrompt, 'matchToggle')}
                {editIsMatchActive && (
                    <>
                        <Box>Threshold value: 0.{editedMatchValue}</Box>
                        <Slider
                            onChange={({ detail }) => setEditedMatchValue(detail.value)}
                            value={Number(editedMatchValue)}
                            valueFormatter={value => '0.' + value}
                            max={99}
                            min={0}
                            disabled={!canUserEditPrompt}
                        />
                    </>
                )}
            </SpaceBetween>
        </>
    );
}
