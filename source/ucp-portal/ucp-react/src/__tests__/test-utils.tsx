// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { SplitPanelContextProvider } from '../contexts/SplitPanelContext';
// import { DEFAULT_INITIAL_STATE } from '../store/types.ts';
import { configureStore, createListenerMiddleware, PreloadedState } from '@reduxjs/toolkit';
import { Provider } from 'react-redux';
import { NotificationContextProvider } from '../contexts/NotificationContext.tsx';
import { MemoryRouter } from 'react-router-dom';
import { AppRoutes } from '../AppRoutes.tsx';
import { render } from '@testing-library/react';
import { rootReducer, RootState, setupStore } from '../store/store.ts';
import { ToolsContextProvider } from '../contexts/ToolsContext.tsx';
import { solutionApi } from '../store/solutionApi.ts';
import { asyncApiSlice } from '../store/asyncApiSlice.ts';
import { AsyncEventType } from '../models/async.ts';
import { domainApiSlice } from '../store/domainApiSlice.ts';
import { FLUSH, PAUSE, PERSIST, PURGE, REGISTER, REHYDRATE } from 'redux-persist';

/*
 * Render a page within the context of a Router, redux store and NotificationContext.
 *
 * This function provides setup for component tests that
 * - interact with the store state,
 *  -navigate between pages
 *  and/or
 * - emit notifications.
 */
export function renderAppContent(props?: { preloadedState?: PreloadedState<RootState>; initialRoute: string }) {
    const store = setupStore(props?.preloadedState);

    const renderResult = render(
        <MemoryRouter initialEntries={[props?.initialRoute ?? '/']}>
            <Provider store={store}>
                <NotificationContextProvider>
                    <ToolsContextProvider>
                        <SplitPanelContextProvider>
                            <AppRoutes></AppRoutes>
                        </SplitPanelContextProvider>
                    </ToolsContextProvider>
                </NotificationContextProvider>
            </Provider>
        </MemoryRouter>,
    );

    return {
        renderResult,
        store,
    };
}
