// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { useContext } from 'react';
import { AppLayout, Flashbar, SplitPanel } from '@cloudscape-design/components';
import SideNavigationBar from './components/navigation/SideNavigationBar.tsx';
import { NotificationContext } from './contexts/NotificationContext.tsx';
import { ToolsContext, ToolsContextType } from './contexts/ToolsContext.tsx';
import { SplitPanelContext, SplitPanelContextType } from './contexts/SplitPanelContext.tsx';
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from './components/navigation/Breadcrumbs.tsx';
import { ToolsContent } from './components/tools/ToolsContent.tsx';
import TopNavigationBar from './components/navigation/TopNavigationBar.tsx';

export default function Layout() {
    const { notifications } = useContext(NotificationContext);
    const { toolsState, setToolsState } = useContext(ToolsContext);
    const { splitPanelState, setSplitPanelState } = useContext(SplitPanelContext);

    return (
        <>
            <div id="top-nav" data-testid={'topNav'}>
                <TopNavigationBar />
            </div>
            <div>
                <AppLayout
                    headerSelector="#top-nav"
                    content={
                        <div data-testid={'main-content'}>
                            <Outlet />
                        </div>
                    }
                    contentType={'dashboard'}
                    breadcrumbs={<Breadcrumbs />}
                    navigation={<SideNavigationBar />}
                    notifications={<Flashbar stackItems={true} items={notifications}></Flashbar>}
                    splitPanelOpen={splitPanelState.isOpen}
                    onSplitPanelToggle={event => {
                        setSplitPanelState((prevState: SplitPanelContextType) => ({
                            ...prevState,
                            isOpen: event.detail.open,
                        }));
                        if (!event.detail.open) splitPanelState.onClose();
                    }}
                    splitPanel={
                        <SplitPanel header={splitPanelState.header} closeBehavior={'hide'} hidePreferencesButton={true}>
                            {splitPanelState.content}
                        </SplitPanel>
                    }
                    stickyNotifications={true}
                    tools={<ToolsContent />}
                    toolsOpen={toolsState.toolsOpen}
                    onToolsChange={e => setToolsState((state: ToolsContextType) => ({ ...state, toolsOpen: e.detail.open }))}
                    ariaLabels={{
                        navigation: 'Navigation drawer',
                        navigationClose: 'Close navigation drawer',
                        navigationToggle: 'Open navigation drawer',
                        notifications: 'Notifications',
                        tools: 'Help panel',
                        toolsClose: 'Close help panel',
                        toolsToggle: 'Open help panel',
                    }}
                />
            </div>
        </>
    );
}
