// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, Container, ContentLayout, Header } from '@cloudscape-design/components';
import { Trans, useTranslation } from 'react-i18next';
import SpaceBetween from '@cloudscape-design/components/space-between';

export const InternationalizationPage = () => {
    const { t, i18n } = useTranslation();

    const languages: Record<string, { nativeName: string }> = {
        en: { nativeName: 'English' },
        'en-US': { nativeName: 'English (US)' },
        'en-GB': { nativeName: 'English (UK)' },
        de: { nativeName: 'Deutsch' },
        'de-AT': { nativeName: 'Deutsch (Österreich)' },
    };

    return (
        <ContentLayout
            header={
                <Header variant="h1">
                    <Trans i18nKey="i18nPage.i18nAndl10n" />
                </Header>
            }
        >
            <SpaceBetween size={'m'}>
                <Container
                    header={
                        <Header variant={'h2'}>
                            <Trans i18nKey="i18nPage.whatIsI18n" />
                        </Header>
                    }
                >
                    <p>
                        <Trans i18nKey="i18nPage.explanation1" />
                    </p>
                    <p>
                        <Trans i18nKey="i18nPage.explanation2" />
                    </p>
                </Container>
                <Container
                    header={
                        <Header variant={'h2'}>
                            <Trans i18nKey="i18nPage.languageSwitching" />
                        </Header>
                    }
                >
                    <p>
                        <Trans i18nKey="i18nPage.localeText" />
                    </p>
                    <p>
                        <Trans i18nKey="i18nPage.languageSwitcherText" />
                    </p>
                    <SpaceBetween size={'xs'} direction={'horizontal'}>
                        {Object.keys(languages).map(lng => (
                            <Button key={lng} onClick={() => i18n.changeLanguage(lng)}>
                                {languages[lng].nativeName}
                            </Button>
                        ))}
                    </SpaceBetween>
                    <p>
                        <Trans i18nKey="i18nPage.fallback" />
                    </p>
                </Container>
                <Container
                    header={
                        <Header variant={'h2'}>
                            <Trans i18nKey="i18nPage.dateFormats" />
                        </Header>
                    }
                >
                    <p>
                        <div>{t('i18nPage.todayIs', { date: new Date() })}</div>
                        <div>{t('i18nPage.nowIs', { date: new Date() })}</div>
                    </p>
                </Container>
            </SpaceBetween>
        </ContentLayout>
    );
};
