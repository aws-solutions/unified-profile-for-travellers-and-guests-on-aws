// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { AmazonConnectApp } from '@amazon-connect/app';
import { ContactClient } from '@amazon-connect/contact';
import { VoiceClient } from '@amazon-connect/voice';
import { withAuthenticator } from '@aws-amplify/ui-react';
import '@aws-amplify/ui-react/styles.css';
import { AppRoutes } from './AppRoutes.tsx';

import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { addNotification, deleteNotification } from './store/notificationsSlice.ts';
import { useBuildAccessPermission } from './utils/authUtil.ts';

let connectHasBeenInitialized = false;

const AppComponent = () => {
    const dispatch = useDispatch();

    useBuildAccessPermission(dispatch);

    //This section setups the necessary hooks when the solution is running withing the Amazon Connect Agent Workspace
    useEffect(() => {
        //prevent connect to be initialized twice
        if (!connectHasBeenInitialized) {
            connectHasBeenInitialized = true;
            console.log('[upt] loading App component');
            //initialize connect app if UPT is running in the agent workspace
            try {
                const { provider } = AmazonConnectApp.init({
                    onCreate: event => {
                        const { appInstanceId } = event.context;
                        console.log('[upt] UPT is running within the agent workspace. App instance ID: ', appInstanceId);
                        // A simple callback that just console logs the contact accepted event data
                        // returned by the workspace whenever the current contact is accepted
                        dispatch(
                            addNotification({
                                id: 'connect-agent-workspace-notification',
                                type: 'in-progress',
                                content: 'UPT is running within the Amazon Connect Agent workspace. Waiting for contact...',
                                loading: true,
                            }),
                        );
                        return new Promise((resolve, reject) => { });
                    },
                    onDestroy: event => {
                        console.log('[upt] App being destroyed');
                        return new Promise((resolve, reject) => { });
                    },
                });

                const contactClient = new ContactClient();
                const voiceClient = new VoiceClient();

                //handler for the "Contact Accepter" Hook
                const handler = async (data: any) => {
                    console.log('[upt] Contact accepted');
                    console.log(`[upt] Contact event callback value`, data);
                    const phone = await voiceClient.getPhoneNumber(data.contactId);
                    console.log(`[upt] phone number: `, phone);

                    dispatch(deleteNotification({ id: 'connect-agent-workspace-notification' }));
                    dispatch(
                        addNotification({
                            id: 'connect-contact-notification',
                            type: 'in-progress',
                            content: `Contact Accepted with id ${data.contactId} and phone ${phone}`,
                            loading: true,
                        }),
                    );
                };

                //handler for teh contact terminated hook
                const disconnectedHandler = async (data: any) => {
                    console.log('[upt] Contact disconnected');
                    dispatch(deleteNotification({ id: 'connect-contact-notification' }));
                    dispatch(
                        addNotification({
                            id: 'connect-agent-workspace-notification',
                            type: 'in-progress',
                            content: 'UPT is running within the Amazon Connect Agent workspace. Waiting for contact...',
                            loading: true,
                        }),
                    );
                };
                // Subscribe to the contact accepted topic using the above handler
                console.log('[upt] Subscribing to contact accepted topic');
                contactClient.onConnected(handler);
                contactClient.onDestroyed(disconnectedHandler);
            } catch (e) {
                console.log('[upt] Error launching the agent workspace. UPT is maybe running standalone', e);
            }
        }
    }, []); // <-- empty dependency array to make sure the connect init only happens once https://stackoverflow.com/questions/58101018/react-calling-a-method-on-load-only-once

    /**
     * Load base data here that should be available on app start up for all pages.
     * Other data will only load once the user navigates to pages that require it.
     */

    return <AppRoutes></AppRoutes>;
};
/**
 * withAuthenticator wraps our App into AWS Amplify's authenticator.
 * This displays a login screen and manages the user token.
 *
 * To try out this frontend template without deploying the backend/authentication infrastructure, comment out the following lines (withAuthenticator...) and replace them by:
 * export const App = AppComponent;
 */
export const App = withAuthenticator(AppComponent, {
    loginMechanisms: ['email', 'username'],
    hideSignUp: true,
});
