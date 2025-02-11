// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { combineReducers, configureStore, createListenerMiddleware, isAnyOf, PreloadedState } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { FLUSH, PAUSE, PERSIST, persistReducer, PURGE, REGISTER, REHYDRATE } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import { AsyncEventType } from '../models/async.ts';
import { asyncApiSlice } from './asyncApiSlice.ts';
import { configSlice } from './configSlice.ts';
import { domainApiSlice } from './domainApiSlice.ts';
import { domainSlice } from './domainSlice.ts';
import { matchesSlice } from './matchesSlice.tsx';
import { notificationsSlice } from './notificationsSlice.ts';
import { permissionSystemSlice } from './permissionSystemSlice.ts';
import { privacyApiSlice } from './privacyApiSlice.ts';
import { privacySlice } from './privacySlice.ts';
import { profileSlice } from './profileSlice.ts';
import { ruleSetSlice } from './ruleSetSlice.ts';
import { solutionApi } from './solutionApi.ts';
import { userSlice } from './userSlice.ts';

const domainConfig = {
    key: 'domains',
    storage,
};

export const rootReducer = combineReducers({
    [solutionApi.reducerPath]: solutionApi.reducer,
    [notificationsSlice.name]: notificationsSlice.reducer,
    [profileSlice.name]: profileSlice.reducer,
    [configSlice.name]: configSlice.reducer,
    [ruleSetSlice.name]: ruleSetSlice.reducer,
    [domainSlice.name]: persistReducer(domainConfig, domainSlice.reducer),
    [privacySlice.name]: privacySlice.reducer,
    [matchesSlice.name]: matchesSlice.reducer,
    [userSlice.name]: userSlice.reducer,
    [permissionSystemSlice.name]: permissionSystemSlice.reducer,
});

// Infer the `RootState` types from the store itself
export type RootState = ReturnType<typeof rootReducer>;

const chainingMiddleware = createListenerMiddleware<RootState>();

chainingMiddleware.startListening({
    matcher: isAnyOf(asyncApiSlice.endpoints.getAsyncStatus.matchFulfilled),
    effect: (action, api) => {
        if (action.payload.status === 'success' || action.payload.status === 'failed') {
            if (action.payload.item_id === AsyncEventType.CREATE_DOMAIN || action.payload.item_id === AsyncEventType.DELETE_DOMAIN) {
                api.dispatch(domainApiSlice.endpoints.getAllDomains.initiate());
            }
            if (action.payload.item_id === AsyncEventType.CREATE_PRIVACY_SEARCH) {
                api.dispatch(privacyApiSlice.endpoints.listSearches.initiate());
            }
        }
    },
});

// See https://redux-toolkit.js.org/usage/usage-guide#use-with-redux-persist for more info on ignoredActions
export const setupStore = (preloadedState?: PreloadedState<RootState>) => {
    const store = configureStore({
        reducer: rootReducer,
        preloadedState,
        middleware: getDefaultMiddleware =>
            getDefaultMiddleware({
                serializableCheck: {
                    ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER],
                },
            })
                .prepend(chainingMiddleware.middleware)
                .concat(solutionApi.middleware),
    });
    setupListeners(store.dispatch);
    return store;
};
