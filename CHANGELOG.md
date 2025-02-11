# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.1] - 2025-02-10

### Changed

- Merge profiles was rewritten for significant optimization, profiles can be processed much faster when rule set merging is on
- Added indexes and moved more workload to db reader instances to improve database performance
- Bucket Keys were enabled on all buckets (cost optimization)
- Internal changes and code cleanup, enabling first open sourced release of solution

### Removed

- Ability to combine profile and object level conditions for rules within a rule set was deprecated

### Fixed

- golang.org/x/crypto updated to remove package issue

## [2.1.0] - 2024-12-09

### Added

- A second CloudFormation template to deploy UPT in an existing VPC
- New CloudFormation parameters for better control of the solution configuration
  - Deploy the web app via CloudFormation, ECS image, or headless
  - Import an existing S3 Bucket for CloudFormation template assets
  - Disable the fine-grained permission system
  - Add Permission Boundary to IAM Roles created during deployment
- Several additional fields for UPT schemas
- New "extended_data" field to store unstructured JSON data
- New object type, Alternate ID, to support a wide range of traveler identifier use cases

### Changed

- Improved indexing of traveler records in downstream caches (Amazon DynamoDB and Amazon Connect Customer Profiles)

### Removed

- References to the deprecated "Travel and Hospitality Application Connectors Catalog on AWS" solution

## [2.0.2] - 2024-10-02

### Fixed

- ListTags permissions to async lambda role
- Added merge queue reference to merger lambda

### Security

- Upgrade dependencies - micromatch, rollup, path-to-regexp, send, serve-static

## [2.0.1] - 2024-08-21

### Changed

- AWS Resources created dynamically after deployment are correctly tagged with customer-provided tags
- Update CloudFormation dependencies to avoid intermittent deployment failures
- Fix minor UI bugs
- Upgrade dependencies

## [2.0.0] - 2024-08-07

### Added

- Aurora-based Interaction Store
- GDPR-compliant purge
- Rule-based interaction stitching
- Generative AI summarization
- Profile display
- Profile search
- Interaction history
- Unmerge
- Rebuild cache

### Changed

- Migrate from Angular UI to React and Cloudscape

## [1.1.0] - 2023-12-07

### Added

- Performance improvements
- Additional CloudFormation template parameters
- Schema updates

## [1.0.1] - 2023-10-20

### Added

- Upgrade packages for vulerability fix

## [1.0.0] - 2020-06-05

### Added

- All files, initial version
