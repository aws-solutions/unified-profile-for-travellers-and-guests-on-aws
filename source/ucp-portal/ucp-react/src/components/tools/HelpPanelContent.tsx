// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Badge, Icon } from '@cloudscape-design/components';
import HelpPanel from '@cloudscape-design/components/help-panel';
import { IconName } from '../../models/iconName';

export const NoHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>There is no help content available for this page. </p>
    </HelpPanel>
);

export const AIMatchesHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <h4>Badge definitions</h4>
        <p>Badges highlight the similarities and differences in attribute values for source and target profiles.</p>
        <ul>
            <li>
                A <Badge color="green">green</Badge> badge indicates both profile have the same values
            </li>
            <li>
                A <Badge color="red">red</Badge> badge indicates both profile have different values
            </li>
            <li>
                A <Badge color="blue">blue</Badge> badge indicates one of the profile has a value for the attribute while the other profile
                has a missing value, indicated by a <Badge color="grey">grey</Badge> badge
            </li>
        </ul>
        <h4>Mark as False Positives</h4>
        <p>If you identify a match pairing to be incorrect, select the checkbox and mark them as false positive</p>
    </HelpPanel>
);

export const CacheManagementHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The Cache Management page contains the configurable rulesets to create a set of rules (skip conditions) to ensure only profiles
            that contain actionable PII are ingested into the profile caches. A ruleset is the sum of all rules. Rulesets can be saved and
            versioned to iterate on the deterministic logic. A ruleset can be activated at any time. Ruleset activation will rerun all
            profiles against the rules in the activated ruleset. There is an implied “OR” for every Skip condition within a rule.{' '}
        </p>
        <h4>Example</h4>
        <p>
            A skip condition of profile, FirstName (business object) equals_value (qualifier) of empty (add nothing to the field) will
            prevent any profile that does not have a first name from being pushed to your configured profile caches of either or both
            DynamoDB and Connect Customer Profiles.
        </p>
    </HelpPanel>
);

export const RulesHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The Rules page contains the configurable rules that provide the deterministic interaction stitching and profile merging logic. A
            ruleset is sum of all rules; the Skip and Match Conditions. Rulesets can be saved and versioned to iterate on the deterministic
            logic. A ruleset can be activated at any time. Ruleset activation will rerun all interactions against the rules in the activated
            ruleset. There is an implied "OR" between every rule within a ruleset, an implied "OR" between every skip condition in a rule
            and an implied "AND" between every match condition in a rule.{' '}
        </p>
        <h4>Skip Condition</h4>
        <p>
            A skip condition met within a rule informs the system that an ingested interaction should skip the entire rule and does not meet
            the requirements of the rule.
        </p>
        <h4>Match Condition</h4>{' '}
        <p>
            A match condition within a rule informs the system that an ingested interaction that meets the match condition should be
            stitched with other interactions and/or profile. Interaction data lineage and profile states are saved when an interaction is
            stitched with another interaction or profile to allow the merge to be unmerged in the <b>Profiles</b> page.
        </p>
        <h4>Example</h4>{' '}
        <p>
            Stitch and merge a logged in clickstream interaction with a loyaltyID to a profile with a loyaltyID. The rule would have a skip
            condition of clickstream (interaction type), customer_loyalty_id (business object) equals value (qualifier) of empty (add
            nothing to the field) to skip the match condition. The same rule would have a match condition of clickstream (interaction type),
            customer_loyalty_id (business object) equals pax_profile (profile) and id (loyalty ID).
        </p>
    </HelpPanel>
);

export const PrivacyHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Privacy</b> page contains the search and delete controls to comply with privacy regulations and laws. UPT ships with a
            data privacy compliant feature. Search will take a comma separated list of Ids, search all Solution data stores for every
            instance of PII associated with the profile IDs, and return the location of PII. Purging the profile will delete all profile
            instances with PII in compliance with Privacy regulations and laws.
        </p>
    </HelpPanel>
);

export const JobsHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The Jobs page lists all batch ingestion glue jobs that must be run to batch ingest interactions from the S3 buckets listed in{' '}
            <b>Domain Settings</b>. The Run Job link will run the batch ingestion jobs, pulling from the S3 buckets.
        </p>
    </HelpPanel>
);

export const DomainSettingsHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Domain Settings</b> page contains all per domain settings, actions, high-level stats, and defined Amazon S3 buckets to
            run interaction batch ingestion. At the top of the Domain Settings page there are two buttons, Rebuild Cache and Delete Domain.
        </p>
        <h4>Rebuild Cache</h4> The Rebuild Cache button will delete and rebuild the profile caches in data stores of DynamoDB or Connect
        Customer Profiles. This operation will impact profile cache data store availability during the rebuild process.
        <h4>Delete Domain</h4>
        <p>
            Delete Domain will delete the profile caches in data stores of DynamoDB or Connect Customer Profiles. All UPT profile data will
            remain in the Aurora cluster. <b>Please note:</b> After deleting a domain you should wait a couple of minutes before creating a
            new domain with the same domain name of the deleted domain. will remain in the Aurora cluster.
        </p>
        <h4>Data Repositories</h4>
        <p>
            UPT batch ingestion requires interaction records be uploaded to the appropriate S3 buckets. Glue Jobs run to transform the
            interaction records with results uploaded to the Traveler Profile Records bucket. These records are read and ingested by UPT.
        </p>
        <h4>Generative AI Prompt Configuration</h4>
        <p>
            The Generative AI profile summary feature is controlled with this configuration, and is dependent on Anthropic Claude 3.0 Sonnet
            being enabled in your AWS Account (see Implementation Guide for more information). When enabled (“On”) the Generative AI Summary
            button will display on the profile summary prompt. The prompt can be changed and saved at any time to change how Generative AI
            builds the profile summary.
        </p>
        <h4>Auto Merging Threshold</h4>
        <p>
            The Auto Merging feature is controlled with this configuration. Ai-based identity resolution profile matches generated by
            Connect Customer Profiles are sent to an Amazon S3 bucket. When the feature is enabled (“On”) all profile matches that are equal
            to or greater than the threshold will be automatically merged.
        </p>
    </HelpPanel>
);

export const ProfileSearchHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Profiles</b> page is where you can search for and view traveler and guest profiles created by the Solution. Search fields
            are exact match (case sensitive) with an “and” qualifier between fields. Each search result has a view profile link that opens
            the profile details page.
        </p>
    </HelpPanel>
);

export const ProfileDisplayHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Profiles Detail</b> page contains all business object (interaction) records that have been associated to a traveler &
            guest profile. At the top of the profile details page there are two buttons, Generate Summary and Search for Data Locations.
        </p>
        <h4>Generate Summary</h4>
        <p>
            The Generate Summary feature must be enabled in the <b>Domain Settings</b> page for the button to be present. The feature sends
            all profile data and the prompt, also configured in the <b>Domain Settings</b> page, to Amazon Bedrock. The profile summary will
            be returned, displayed in the UI, and saved against the profile.
        </p>
        <h4>Search for Data Locations</h4>
        <p>The Search for Data Locations feature will return all data stores where all profile histories are stored.</p>
        <h4>Business Object Records</h4>
        <p>
            Every business object (interaction) sent to UPT is saved to a UPT profile with unification lineage and can be viewed by opening
            the record. Profile data lineage is a record of every profile an interaction was associated with before and after every
            interaction stitching or identity resolution merge activity. This allows customers to identify the cause and effect of all
            interaction and profile merge.
        </p>
        <p>
            Every interaction has a confidence factor that identifies how the interaction became associated with the current profile
            (AI-identity resolution, Rule based, or Manual merges). The confidence score is "100% ("1") for all interactions where a rule
            associated the interaction and merged profiles. The confidence score is &lt; 1 for all interactions where AI-based identity
            resolution was used to merge profiles. Clicking the confidence score will open the interaction history details pane with
            additional information and the capability to unmerge the profiles back to their previous profile state while preserving full
            interaction lineage for both profiles.
        </p>
    </HelpPanel>
);

export const ConfigHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Dashboard Configuration</b> page is where you can use the value of any object in a url to link out to third party
            systems. For example, if you wanted loyalty_id to be a navigable link to a third party system, then you would select air_loyalty
            from the first dropdown, id from the second dropdown, and add &#123;id&#125; to the final url
            https://admin.brand.com/member/&#123;id&#125;
        </p>
    </HelpPanel>
);

export const ErrorsHelp = () => (
    <HelpPanel header={<h2>Help panel</h2>}>
        <p>
            The <b>Errors</b> page is where you can view any error collected by one of the Solution error queues. The solution generates
            trace IDs (called transaction ID or tx_id) allowing you to easily troubleshoot ingestion and identity resolution errors. For
            more detailed error discovery and troubleshooting see{' '}
            <a href="https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/activate-cloudwatch-application-insights.html">
                Searching for Logs
            </a>{' '}
            in the Solution Implementation Guide to perform CloudWatch log search.
        </p>

        <h4>Common Error Types</h4>

        <b>
            <i>Unknown object type</i>
        </b>
        <p>
            This error occurs when the name of the object type is invalid. For the supported list of object types, see{' '}
            <a href="https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/sending-data-to-the-real-time-stream.html#kinesis-wrapper-schema">
                Kinesis wrapper schema
            </a>{' '}
            in the Solution Implementation Guide to perform CloudWatch log search.
        </p>

        <b>
            <i>DomainName not found</i>
        </b>
        <p>
            This error occurs when the name of the Amazon Connect Customer Proﬁles domain cannot be found in the account. Verify that there
            is an existing Domain Name. Use the <Icon name={IconName.STATUS_INFO} /> icon to open the interaction and verify “domain” is
            correct in the ingested record.
        </p>

        <b>
            <i>Transformer_exception 'stringValue'</i>
        </b>
        <p>
            This error occurs when a Clickstream attribute has a "string" type and the StringValue ﬁeld is not populated. Validate that your
            Clickstream attribute type and ﬁelds are consistent. For details, see{' '}
            <a href="https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/sending-data-to-the-real-time-stream.html#web-and-mobile-events">
                web and mobile events schema
            </a>
        </p>

        <b>
            <i>'NoneType' object is not iterable</i>
        </b>
        <p>
            This error occurs when the solution ingestion expects an “array” type and receives a null type. Use the{' '}
            <Icon name={IconName.STATUS_INFO} /> icon to open the interaction and find the field set to null in the ingested record. For the
            supported list of object types, see{' '}
            <a href="https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/sending-data-to-the-real-time-stream.html#kinesis-wrapper-schema">
                Kinesis wrapper schema
            </a>{' '}
            in the Solution Implementation Guide
        </p>

        <b>
            <i>Unconverted data remains</i>
        </b>
        <p>
            This error occurs when an ingested record has an incorrectly formatted timestamp. All timestamps must follow ISO 8601 format
            (including milliseconds).
        </p>

        <b>
            <i>A database-level error occurred while ….</i>
        </b>
        <p>
            This error occurs when an ingested record fails to be written to the interaction store. To troubleshoot the issue, Use the{' '}
            <Icon name={IconName.STATUS_INFO} /> icon to open the interaction, find the transaction ID (“tx_id”), and perform a CloudWatch
            log insight search as described in{' '}
            <a href="https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/activate-cloudwatch-application-insights.html">
                Searching for logs
            </a>{' '}
            in the Solution Implementation Guide
        </p>

        <b>
            <i>Unknown SQS record type with Body '&#123;&#125;' and attributes map[]</i>
        </b>
        <p>
            This error occurs at domain creation when Amazon Connect Customer Profiles sends a test message in the error queue to validate
            access. You can safely disregard this error.
        </p>
    </HelpPanel>
);

export const HomeHelp = () => (
    <HelpPanel header={<h2>Dashboard</h2>}>
        <p>
            The Solution Console is used to administer and configure UPT. The left navigation pane has two administration sections of
            “Overview” and “DomainName” (selectable from the top right of the Solution Console after a Domain has been created). All pages
            beneath “DomainName” are Domain specific.
        </p>
    </HelpPanel>
);
