# Schema Generators

UPT heavily relies on schemas. We build generators to build schemas for validation and testing.

## Real-Time Generator

Send profile objects in as continuous traffic to the configured Kinesis Data Stream for ingestion.

#### Usage
1. From `schemas/generators/real-time`, copy `env.example.json` and save as `env.json`
1. Update each parameter based on the UPT stack you want to target and the amount of traffic you want to send
    - **awsProfile** select the configured AWS profile to use (must match the account of the target UPT stack)
    - **stream** Kinesis Data Stream name for the ingestor stream (`kinesisStreamNameRealTime` in the stack output)
    - **region** AWS region for the target UPT stack
    - **numProfiles** Number of profiles to generate and be used in profile object generation (low number for rich profiles, high number for more profiles)
    - **businessObject** Business object to generate records for
    - **traffic** Approximate number of requests per second
    - **domain** Target UPT domain
1. Run `generate-local.sh`
1. Kill the terminal process once desired traffic has been sent
