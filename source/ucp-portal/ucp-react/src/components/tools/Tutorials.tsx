// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { AnnotationContextProps, TutorialPanel, TutorialPanelProps } from '@cloudscape-design/components';

export const ProjectsTutorial = () => {
    const onboardingTutorial: AnnotationContextProps.Tutorial = {
        completed: false,
        completedScreenDescription: 'Outcome of the tutorial.',
        description: 'Description of the tutorial',
        title: 'Title',
        prerequisitesNeeded: false,
        prerequisitesAlert: 'Before starting this tutorial, take some prerequisite action.',
        tasks: [
            {
                title: 'Foo',
                steps: [
                    {
                        title: 'Foo 1',
                        content: (
                            <>
                                <p>Choose a Portfolio</p>
                            </>
                        ),
                        hotspotId: 'foo-tutorial-hotspot-1',
                    },
                    {
                        title: 'Foo 2',
                        content: <>Click this button</>,
                        hotspotId: 'foo-tutorial-hotspot-2',
                    },
                ],
            },
            {
                title: 'Navigate to the Create Item page',
                steps: [
                    {
                        title: 'Create item',
                        content: <>This is the create form.</>,
                        hotspotId: 'create-item-hotspot',
                    },
                    {
                        title: 'Press the Create button',
                        content: <>Click 'Create' now.</>,
                        hotspotId: 'create-button',
                    },
                ],
            },
        ],
    };

    return <TutorialPanel tutorials={[onboardingTutorial]} i18nStrings={tutorialPanelI18nStrings} />;
};
const tutorialPanelI18nStrings: TutorialPanelProps.I18nStrings = {
    labelsTaskStatus: {
        pending: 'Pending',
        'in-progress': 'In progress',
        success: 'Success',
    },
    loadingText: 'Loading',
    tutorialListTitle: 'Choose a tutorial',
    tutorialListDescription: 'Use our walk-through tutorials to learn how to achieve your desired objectives with this solution.',
    tutorialListDownloadLinkText: '',
    tutorialCompletedText: 'Tutorial completed',
    labelExitTutorial: 'dismiss tutorial',
    learnMoreLinkText: 'Learn more',
    startTutorialButtonText: 'Start tutorial',
    restartTutorialButtonText: 'Restart tutorial',
    completionScreenTitle: 'Congratulations! You completed the tutorial.',
    feedbackLinkText: 'Feedback',
    dismissTutorialButtonText: 'Dismiss tutorial',
    taskTitle: (taskIndex, taskTitle) => `Task ${taskIndex + 1}: ${taskTitle}`,
    stepTitle: (stepIndex, stepTitle) => `Step ${stepIndex + 1}: ${stepTitle}`,
    labelTotalSteps: totalStepCount => `Total steps: ${totalStepCount}`,
    labelLearnMoreExternalIcon: 'Opens in a new tab',
    labelTutorialListDownloadLink: 'Download PDF version of this tutorial',
};
