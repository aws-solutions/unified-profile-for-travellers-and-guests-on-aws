**This AWS Solution is no longer available. We encourage customers to explore using  [Amazon Connect Customer Profiles](https://aws.amazon.com/connect/customer-profiles/)) to deduplicate, unify, segment, and activate their traveler and guest profiles. You can find other AWS Solutions in the [AWS Solutions Library](https://aws.amazon.com/solutions/).**

# Unified Profiles For Travelers and Guests on AWS

Unified Profiles for Travelers and Guests on AWS (UPT) is a solution designed to bring together customer data from various internal systems into a single, comprehensive profile. By automatically sourcing, merging, de-duplicating, and centralizing guest information, UPT enables organizations to gain a 360-degree view of their customers. This unified profile serves as a valuable resource for marketing, sales, operations, and customer experience teams, facilitating personalized engagements, improved customer retention, and increased loyalty.

The solution provides a scalable means of transforming disparate booking, conversation, clickstream, etc ... data from source systems into profiles. It leverages rules-based and AI identity resolution, determining identity from collections of interactions. This approach allows for real time resolution and targeting in an event driven manner. It also reduces batch historical data transformation and integration time from months to weeks.

The solution provides downstream integration for marketing and customer service usecases. Dynamo and Amazon Connect Customer Profile integrations are provided out of the box. Profile data is in a format that can be sent to other platforms (eg. Salesforce) seamlessly. 

## Project Overview

- [UPT in Solutions Library](https://aws.amazon.com/solutions/implementations/unified-profiles-for-travelers-and-guests-on-aws/)
- [Implementation Guide](https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/solution-overview.html)
- [Introducing UPT](https://aws.amazon.com/blogs/industries/introducing-unified-profiles-for-travelers-and-guests-on-aws/)

 The solution diagram can be found in the UPT in Solutions Library link above. 

### CI/CD

- `deployment/developer-pipeline` scripts to run the developer pipeline
- `deployment/ucp-code-pipelines` CDK project with full solution CI/CD pipeline, used internally for development
- `deployment/build-s3-dist.sh` simplified build script that produces the artifacts necessary to publish the solution

### Development

- `source/storage` primary implementation of LCS (Low Cost Storage). This is the primary profile storage component of UPT, built on Aurora Postgres
- `source/tah_lib` python package for business object transformation
- `source/tah-common` schemas for Travel and Hospitality
- `source/tah-core` utility functions and AWS SDK wrappers for Travel and Hospitality
- `source/ucp-async` invoked by **ucp-backend** for longer running use cases that would otherwise hit the 30 second timeout for API Gateway
- `source/ucp-backend` primary API for the solution, sitting behind API Gateway
- `source/ucp-batch` batch processing for input data
- `source/ucp-change-processor` subscribes to the Amazon Connect Customer Profile (ACCP) stream for changes, writes to S3 and optionally EventBridge
- `source/ucp-common` shared code to use across UPT components
- `source/ucp-cp-indexer` dynamically indexes CP export streams to create ID indexing between UPT and CP, allowing usage of CP is a completely readless manner
- `source/ucp-cp-writer` dedicated lambda for writing to CP (Amazon Connect Customer Profiles)
- `source/ucp-error` process and store errors from the various DLQs in a centralized DynamoDB table
- `source/ucp-etls` Glue jobs
- `source/ucp-fargate` Fargate tasks for identity resolution
- `source/ucp-industry-connector-transfer` copies events from Industry Connector solution into an S3 bucket for UPT to process via Glue job (deprecated)
- `source/ucp-infra` CDK project for all solution infrastructure
- `source/ucp-match` process csv files that contain match scores between profiles, produced by ACCP or optionally a third party service that customers can configure
- `source/ucp-merger` handles instances of merging in solution (rules based IR and AI IR)
- `source/ucp-portal` Angular web app to use and manage UPT
- `source/ucp-real-time-transformer` Kinesis stream for real-time record ingestion -> Python Lambda that transforms the nested **business object** (tah-common schema) into flat **ACCP objects** (mapping stored in ucp-common's accp-mapping folder) -> Go Lambda that sends **ACCP objects** to ACCP
- `source/ucp-retry` received specific error types from **ucp-error** that should be retried
- `ucp-s3-excise-queue-processor` functionality for hard deletes of profile information from s3 buckets
- `source/ucp-sync` maintains partitions for business object S3 buckets and triggers Glue jobs based on the customer's configuration
- `source/z-coverage` test output produced by the solution (prefixed with z for convenience to have test files at the end of CRs)
- `vendor` contains all packages used by Go (this provides consistent builds without pipelines needing to go out and fetch packages, particularly important since we use internal packages that are not distributed)

## Branches

- `develop` contains the latest code for the upcoming release.
- `release/vX.X.X` release branches are created off develop at publish time to have a snapshot of the exact release code.
- `feature/upt-[type]-[number]` feature branches are created off develop and used for development. When feature branches are pushed to remote, pipelines are kicked off to test the code and run the AWS Solutions developer pipeline.

## Getting Started

### Prerequisites

- Go
  - Not to use proxy, `export GOPROXY=direct`
- Node
- Python
  - python3 -m pip install coverage
  - python3 -m pip install requests
- jq
- AWS CLI: make sure you have configured your AWS credential.

### Setup

1. Create two S3 buckets in your local AWS account
   1. One bucket for storing UPT artifacts, e.g. `<alias>-upt-artifacts`
   1. Another bucket for storing regional assets. This should match the bucket created above with region appended, e.g. `<alias>-upt-artifacts-<region>`
1. Clone the repository and navigate to the root directory _(note: there will likely be errors if opened in an IDE until all the necessary packages are imported)_
1. Locate `env.example.json` in `source/`. Create a copy named `env.json` and update _artifactBucket_ to be the bucket name you just created, without the region, e.g. `<alias>-upt-artifacts`.
   1. `env.json` is referenced in scripts to get local env settings
   1. The `artifactBucket` is where all the artifacts (Lambda function binaries, zipped code, etc) will be stored.
   1. Update the values for `ecrRepository` and `ecrPublishRoleArn` to match the repository and role in your account. These resources can be created by deploying the `UptInfra` stack.
   1. The unit tests can be configured to use a "test" cluster in AWS RDS, or a local Postgres Database running as a Container.
      1. To run unit and storage tests against an RDS cluster deployed to your AWS Account, run the `TestAuroraPosgres` test found in `source/tah-core/aurora/aurora_test.go`.  Alternatively, create the Aurora test cluster, with a secure connection, and input the parameters for that cluster in env.json.  A Secret in SecretsManager will also need to be created with keys: username and password.  Note that this is strictly for testing purposes.  The solution itself deploys its own resources including its own cluster. The E2E tests use this solution cluster.
      1. To run unit and storage tests against a locally deployed Postgres container, run the "run-local-aurora.sh" script found in the `source` folder.  This script will create the necessary certificates to enable SSL, as well as scaffold the Secret in SecretsManager based on values in the env.json file.
1. Pull in the latest code for the private tah packages
   1. Navigate to `source/` and run `update-private-packages.sh`

### Cloudformation Deployment

### Option 1 - Cloudformation Template (Recommended)

1. In the solutions library (UPT in Solutions Library) you will find the cloudformation template. Download the template, and in the AWS account you wish to deploy the solution go to Cloudformation, and deploy the solution. Reference the implementation guide for more information on deployment and parameters used to deploy the solution. 

### Local Deployment

#### Option 2 - Deploy Lambdas and infrastructure one by one

1. Deploy each Lambda function's code individually (`source/upt-` prefixes excluding `ucp-common`, `ucp-core`, `ucp-infra`, and `ucp-portal`) using the `deploy-local.sh` script in each microservice's directory
1. Deploy infrastructure in `source/ucp-infra` using its `deploy-local.sh` script
   1. If the solution is not deployed, optionally add the `--skip-tests` option, as some of the tests may fail.
1. Run the React web app locally using its `deploy-local.sh` script
   1. Optionally - deploy `ucp-portal`, the React web app for UPT, to CloudFront so it can be accessed via the CloudFront distribution's URL (can be found in CloudFormation output)
   1. Update and invalidate cache using `deploy-dev.sh`. This will sync the UI you can can access via the CloudFormation output url with the code you have locally.
   1. Note, front end is only accessible in this manner if the solution was deployed with the CFN Parameter on UI deployment set to CF (Cloudfront)

#### Option 3 - Deploy everything at once

1. Run `deploy-local-all.sh` to deploy everything at once (this will take a while to run)
   1. If the solution is not deployed, then some of the tests may fail, and there will be no lambda functions to update. When first setting up the solution, run: `sh deploy-local-all.sh --skip-errors`. If the solution fails to deploy, add the `--skip-tests` options: `sh deploy-local-all.sh --skip-errors --skip-tests`

### Configuration

1. Follow the instructions in the [Implementation Guide](https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/deploy-the-solution.html). If you choose to deploy locally, start in the section after **Launch the stack**. This will walk through setting up a user in Cognito, configuring a domain in Amazon Connect Customer Profiles, and configuring the domain. This also helps ensure our Implementation Guide is easy to follow for customers.

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the Apache-2.0 License.

## Collection of operational metrics

This solution collects anonymized operational metrics to help AWS improve the
quality of features of the solution. For more information, including how to disable
this capability, please see the [implementation guide](https://docs.aws.amazon.com/solutions/latest/unified-profiles-for-travelers-and-guests-on-aws/reference.html#anonymized-data-collection).
