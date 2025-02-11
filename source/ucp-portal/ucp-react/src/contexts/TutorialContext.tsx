// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ReactNode, useState } from 'react';
import { AnnotationContext, AnnotationContextProps } from '@cloudscape-design/components';

export const TutorialContextProvider = ({ children }: { children: ReactNode }) => {
    // const navigate = useNavigate(); // if needed
    const [currentTutorial, setCurrentTutorial] = useState<AnnotationContextProps.Tutorial | null>(null);
    return (
        <AnnotationContext
            currentTutorial={currentTutorial}
            children={children}
            i18nStrings={{
                stepCounterText: (stepIndex, totalStepCount) => 'Step ' + (stepIndex + 1) + '/' + totalStepCount,
                taskTitle: (taskIndex, taskTitle) => 'Task ' + (taskIndex + 1) + ': ' + taskTitle,
                labelHotspot: (openState, stepIndex, totalStepCount) =>
                    openState
                        ? 'close annotation for step ' + (stepIndex + 1) + ' of ' + totalStepCount
                        : 'open annotation for step ' + (stepIndex + 1) + ' of ' + totalStepCount,
                nextButtonText: 'Next',
                previousButtonText: 'Previous',
                finishButtonText: 'Finish',
                labelDismissAnnotation: 'dismiss annotation',
            }}
            onExitTutorial={() => {
                setCurrentTutorial(null);
            }}
            onStartTutorial={props => {
                setCurrentTutorial(props.detail.tutorial);
                // add extra logic here, e.g. navigating to a page
            }}
            onFinish={() => {
                setCurrentTutorial(state => (state ? { ...state, completed: true } : null));
            }}
        ></AnnotationContext>
    );
};
