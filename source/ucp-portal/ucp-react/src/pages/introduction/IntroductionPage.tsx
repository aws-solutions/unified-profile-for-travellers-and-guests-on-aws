// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { Box, Button, ColumnLayout, Container, ContentLayout, Header } from '@cloudscape-design/components';
import { ROUTES } from '../../models/constants';

export default function IntroductionPage() {
    return (
        <ContentLayout
            header={
                <Header
                    data-id="Solution-header"
                    variant="h1"
                    description="The easiest way to centralize, unify, and operationalize Traveler and Guest Profiles."
                    actions={
                        <Button variant="primary" href={ROUTES.CREATE}>
                            Create Domain
                        </Button>
                    }
                >
                    Unified profiles for travelers and guests on AWS
                </Header>
            }
        >
            <br></br>
            <br></br>
            <br></br>
            <Box variant="h2" padding={{ bottom: 's' }}>
                Overview
            </Box>
            <Container>
                <b>Unified Profiles for Travelers and Guests</b> (UPT) is an AWS Solution that helps Travel & Hospitality brands address the
                challenge of centralizing, unifiying, and operationalizing traveler and guest profiles. UPT ingests and normalizes all
                traveler and guest interactions from T&H brand's disparate source systems of Clickstream, Customer Service, Reservations,
                Property Management, Passenger Safety, Booking, Loyalty, and more into a central interaction data store. Configurable
                rulesets run in real time to coalesce interactions into a single unified traveler and guest profile that is enriched with
                the context of their engagement and behavior extracted from the ingested online and offline source system interactions.
                Unified profile versions, and interaction and profile merge and unmerge data lineage are automatically saved in UPT. The
                most recent profile version is cached in configurable profile caches of Amazon Connect Customer Profiles and Amazon DyanmoDB
                to deliver AI-based identity resolution and low-latency personalization use cases. All UPT events are published to common
                data services of Amazon S3 and Amazon EventBridge and make it easy to integrate UPT with marketing automation, CDP, and any
                other activation platforms.
            </Container>
            <br></br>

            <Box variant="h3" padding={{ top: 's', bottom: 's' }}>
                Benefits
            </Box>
            <Container>
                <ColumnLayout columns={3} variant="text-grid">
                    <div>Accelerate and own your C360 implementation without building and managing big data infrastructure.</div>
                    <div>
                        Easily integrate with leading CDP, Marketing Automation, and other activation platforms to unbundle functionality
                        and choose the best tool for the job.
                    </div>
                    <div>
                        Use UPT with AWS tools and capabilities that you are familiar with to deploy scalable, reliable, and secure AI/ML
                        and generative AI applications.
                    </div>
                </ColumnLayout>
            </Container>
        </ContentLayout>
    );
}
