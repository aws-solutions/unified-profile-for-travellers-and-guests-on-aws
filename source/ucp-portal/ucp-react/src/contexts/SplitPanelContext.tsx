// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { createContext, ReactNode, useState } from 'react';

export type SplitPanelContextType = {
    header: string;
    content: ReactNode;
    isOpen: boolean;
    onClose: () => void;
};

const DEFAULT_STATE = {
    header: '',
    content: <></>,
    isOpen: false,
    onClose: () => {},
};
export const SplitPanelContext = createContext<{
    splitPanelState: SplitPanelContextType;
    setSplitPanelState: (value: ((prevState: SplitPanelContextType) => SplitPanelContextType) | SplitPanelContextType) => void;
}>(null as any);
export const SplitPanelContextProvider = (props: { children: ReactNode }) => {
    const [splitPanelState, setSplitPanelState] = useState<SplitPanelContextType>(DEFAULT_STATE);

    return (
        <>
            <SplitPanelContext.Provider value={{ splitPanelState, setSplitPanelState }}>{props.children}</SplitPanelContext.Provider>
        </>
    );
};
