# CDK Sleep Audio Pipeline (Go)

An event-driven, serverless AWS CDK application built in Go for processing sleep audio files. This pipeline leverages S3, EventBridge, Lambda/Step Functions, DynamoDB, and SNS to create a scalable, reliable audio processing system following AWS Well-Architected principles and strict Test-Driven Development (TDD) practices.

## Project Philosophy

This project is built with **TDD-first, issue-driven development**. Every feature and change must follow the strict TDD cycle:

### Strict TDD Rules

1. **🔴 RED**: Write a failing test first - never write production code without a failing test
2. **🟢 GREEN**: Write the minimal code to make the test pass
3. **🔵 REFACTOR**: Improve code quality while keeping all tests green
4. **📋 ARCHITECTURE**: Keep `ARCHITECTURE.md` and its Mermaid diagram in sync with every change
5. **✅ VERIFY**: All tests must pass and `cdk synth` must succeed before committing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Architecture

The system implements an event-driven architecture:

- **S3** → **EventBridge** → **Processing (Lambda/Step Functions)** → **S3/DynamoDB/SNS**

For detailed architecture information, see [ARCHITECTURE.md](ARCHITECTURE.md).

The `cdk.json` file tells the CDK toolkit how to execute your app.

## Useful commands

- `go test -v ./...`     run all unit tests (run this first!)
- `cdk synth`            emits the synthesized CloudFormation template
- `cdk diff`             compare deployed stack with current state
- `cdk deploy`           deploy this stack to your default AWS account/region
- `cdk destroy`          remove the deployed stack

## Getting Started

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
