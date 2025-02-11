// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ContentLayout, Header, SpaceBetween } from '@cloudscape-design/components';
import { selectCurrentDomain } from '../../store/domainSlice.ts';
import { useSelector } from 'react-redux';
import { OverviewContainers } from '../../components/setting/overviewContainers.tsx';
import { OverviewDataRepositories } from '../../components/setting/overviewDataRepositories.tsx';
import { useGetDomainQuery } from '../../store/domainApiSlice.ts';
import { skipToken } from '@reduxjs/toolkit/query';
import { useState } from 'react';
import DomainConfiguration from '../../components/setting/domainConfig.tsx';
import { DeleteDomainModal } from '../../components/setting/deleteDomainModal.tsx';
import { RebuildCacheModal } from '../../components/setting/rebuildCacheModal.tsx';
import { selectAppAccessPermission } from '../../store/userSlice.ts';
import { Permissions } from '../../constants.ts';

export const DomainSettingsPage = () => {
    const selectedDomain = useSelector(selectCurrentDomain);
    const domainName = selectedDomain.customerProfileDomain;
    const [isDeleteDomainModalVisible, setIsDeleteDomainModalVisible] = useState<boolean>(false);
    const [isRebuildCacheModalVisible, setIsRebuildCacheModalVisible] = useState<boolean>(false);
    const [isAsyncRunning, setIsAsyncRunning] = useState<boolean>(false);
    const { data: configData, isFetching, isUninitialized } = useGetDomainQuery(domainName !== '' ? domainName : skipToken);
    const isLoading = isFetching || isUninitialized || isAsyncRunning;

    const headerText = isFetching || isUninitialized ? 'Loading: ' + domainName : domainName;

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserRebuildCache = (userAppAccess & Permissions.RebuildCachePermission) === Permissions.RebuildCachePermission;
    const canUserDeleteDomain = (userAppAccess & Permissions.DeleteDomainPermission) === Permissions.DeleteDomainPermission;

    return (
        <div data-testid={'settingsPage'}>
            <ContentLayout
                header={
                    <Header
                        data-testid="settingsPageHeader"
                        variant="h1"
                        actions={
                            <SpaceBetween direction="horizontal" size="s">
                                <Button
                                    data-testid="rebuildCacheButton"
                                    variant="primary"
                                    disabled={isLoading || !canUserRebuildCache}
                                    onClick={() => setIsRebuildCacheModalVisible(true)}
                                >
                                    Rebuild Cache
                                </Button>
                                <Button
                                    data-testid="deleteDomainButton"
                                    variant="primary"
                                    disabled={isLoading || !canUserDeleteDomain}
                                    onClick={() => setIsDeleteDomainModalVisible(true)}
                                >
                                    Delete Domain
                                </Button>
                            </SpaceBetween>
                        }
                    >
                        {headerText}
                    </Header>
                }
                data-testid="settingsContent"
            >
                <OverviewContainers selectedDomain={selectedDomain} isLoading={isLoading} />
                <br />
                <OverviewDataRepositories configData={configData} isDisabled={isLoading} />
                <DeleteDomainModal
                    isModalVisible={isDeleteDomainModalVisible}
                    setIsModalVisible={setIsDeleteDomainModalVisible}
                    setIsAsyncRunning={setIsAsyncRunning}
                />
                <RebuildCacheModal isModalVisible={isRebuildCacheModalVisible} setIsModalVisible={setIsRebuildCacheModalVisible} />
                <br />
                <br />
                <DomainConfiguration />
            </ContentLayout>
        </div>
    );
};
