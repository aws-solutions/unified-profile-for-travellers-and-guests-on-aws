// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import i18next from 'i18next';
import I18nextBrowserLanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';
import { DateTime } from 'luxon';

i18next
    .use(I18nextBrowserLanguageDetector)
    .use(initReactI18next)
    .init({
        debug: true,
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false, // not needed for react as it escapes by default
        },
        resources: {
            en: {
                translation: {
                    sidenav: {
                        header: 'Unified profiles for Travelers and Guests on AWS',
                        introduction: 'Introduction',
                        crudExample: 'CRUD Example / Item table',
                        backendIntegration: 'Backend Integration',
                        internationalization: 'Internationalization (i18n)',
                    },
                    i18nPage: {
                        i18nAndl10n: 'Internationalization and localization',
                        whatIsI18n: 'What is internationalization? What is localization?',
                        languageSwitching: 'Language switching',
                        dateFormats: 'Date formats',
                        explanation1:
                            'Internationalization (i18n) is the process of extracting hardcoded texts, data formats, number formats etc from an application and leave placeholders there. Localization (l10n) is the process of plugging localized content like translated texts back into these placeholders. For brevity, developers tend to use the term i18n (pronounced `i` - `eighteen` - `n`) to describe both.',
                        explanation2:
                            'If you need to publish your solution in English only, you can ignore this guide and hardcode texts in your UI components instead, to avoid the overhead and added complexity of i18n.',
                        todayIs: 'Today is {{date, DATE_MED}}',
                        nowIs: 'Now is {{date, DATETIME_SHORT}}',
                        localeText:
                            'In software, a language and related settings are identified by a code called a Locale. A locale typically consists of a language code only (like "en" for English), or a combination of language and country (like "en-US"). More specific locales (like regional variants of a language within a country) are defined in the IETF BCP 47 standard as well, but not as commonly used.',
                        languageSwitcherText:
                            'By default, your app using i18next will identify what language your browser is set to, and use that locale. Optionally, your app can implement a language switcher widget that allows the user to select a language and override the setting i18next read from the browser.',
                        fallback:
                            'i18next will look up each text in the i18n.ts file for each element if it is available in the set locale. If it is not available, i18next will fall back from the more specific locale (like "de-AT") to the more general locale ("de") and try to use that. If the text is also not available in that local, i18next will fall back on the fallback language that was set in i18n.ts ("en")',
                    },
                },
            },
            de: {
                translation: {
                    sidenav: {
                        header: 'Beispielseiten',
                        introduction: 'Einführung',
                        crudExample: 'CRUD-Beispiel / Tabelle',
                        backendIntegration: 'Backend-Integration',
                        internationalization: 'Internationalisierung (i18n)',
                    },
                    i18nPage: {
                        i18nAndl10n: 'Internationalisierung und Lokalisierung',
                        whatIsI18n: 'Was ist Internationalisierung? Was ist Lokalisierung?',
                        languageSwitching: 'Sprachauswahl',
                        dateFormats: 'Datenformate',
                        explanation1:
                            'Internationalisierung (i18n) bedeutet hardgecodete Texte, Datenformate, Zahlenformate etc aus einer Anwendung zu extrahieren und durch Platzhalter zu ersetzen. Lokalisierung (l10n) ist der Prozess lokalisierte Inhalte wie Texte in diese Platzhalter einzufügen. In Kurzform fassen Softwareentwickler beide Prozesse oft unter dem Begriff "i18n" (sprich: "i-eighteen-n") zusammen.',
                        explanation2:
                            'Wenn du deine Solution nur auf English veröffentlichen musst, ignoriere diese Anleitung und nutze stattdessen hartgecodete Texte.',
                        todayIs: 'Heute ist {{date, DATE_MED}}',
                        nowIs: 'Jetzt ist {{date, DATETIME_SHORT}}',
                        localeText:
                            'In software, a language and related settings are identified by a code called a Locale. A locale typically consists of a language code only (like "en" for English), or a combination of language and country (like "en-US"). More specific locales (like regional variants of a language within a country) are defined in the IETF BCP 47 standard as well, but not as commonly used.',
                        languageSwitcherText:
                            'By default, your app using i18next will identify what language your browser is set to, and use that locale. Optionally, your app can implement a language switcher widget that allows the user to select a language and override the setting i18next read from the browser.',
                        fallback:
                            'i18next will look up each text in the i18n.ts file for each element if it is available in the set locale. If it is not available, i18next will fall back from the more specific locale (like "de-AT") to the more general locale ("de") and try to use that. If the text is also not available in that local, i18next will fall back on the fallback language that was set in i18n.ts ("en")',
                    },
                },
            },
        },
    });

i18next.services.formatter?.add('DATE_MED', (value, language, options) => {
    return DateTime.fromJSDate(value)
        .setLocale(language ?? 'en-US')
        .toLocaleString(DateTime.DATE_MED);
});
i18next.services.formatter?.add('DATETIME_SHORT', (value, language, options) => {
    return DateTime.fromJSDate(value)
        .setLocale(language ?? 'en-US')
        .toLocaleString(DateTime.DATETIME_SHORT);
});
