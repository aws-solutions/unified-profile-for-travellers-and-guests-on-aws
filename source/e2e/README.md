# End-to-End Testing Repository

## Overview

This repository centralizes the end-to-end (e2e) tests for our project. Previously, e2e tests were included within specific components, making it challenging to focus on testing use cases comprehensively.

## Best Practices

- Use SDK to simulate interacting with UPT in the same way a customer would (e.g. making API calls instead of directly interacting with underlying service)
- All e2e tests should be able to run in parallel
- Tests should be runnable directly from the IDE in `run` or `debug` mode, and should NOT rely on a shell script to set environment variables. That pattern makes it significantly harder to debug tests locally.
- Testing specific utility functions are stored in the utils.go file in this directory. Non-testing specific utility functions, for example, making an api call, are stored in the uptSdk directory. A good rule of thumb is that if the function takes testing.T as a parameter, it goes in this directory, otherwise in uptSdk.
- Repository is organized in 4 catagories: PutProfileObject testing, Merge Testing, Rules based merging testing, and cache testing. Each catagory should be reflect a feature being tested. Each test should test a specific usecase, and be isolated to that usecase. The idea is that if something breaks, we know what specific thing is broken by examining which test failed.
