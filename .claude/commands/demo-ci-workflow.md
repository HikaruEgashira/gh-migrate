---
description: Add a simple GitHub Actions CI workflow file
---
Add a simple GitHub Actions workflow file at .github/workflows/ci.yml that:

1. Triggers on:
   - Push to main branch
   - Pull requests to main branch

2. Contains a single job called "build" that:
   - Runs on ubuntu-latest
   - Uses actions/checkout@v3 to checkout the repository

Keep the workflow minimal and straightforward - just the basic structure for CI without any build or test steps yet.

## Summary
Add a reusable slash command for creating a simple GitHub Actions CI workflow with actions/checkout@v3 that triggers on push and pull requests to main branch.