# CDK Sleep Audio Pipeline (Go)

A production-ready, event-driven serverless AWS CDK application built in Go for processing sleep audio files. This pipeline leverages S3, EventBridge, Lambda, Step Functions, DynamoDB, and SNS to create a scalable, reliable audio processing system following AWS Well-Architected principles and strict Test-Driven Development (TDD) practices.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Project Philosophy](#project-philosophy)
- [Getting Started](#getting-started)
- [Deployment](#deployment)
- [Usage](#usage)
- [Testing](#testing)
- [Documentation](#documentation)
- [Contributing](#contributing)

## Overview

The Sleep Audio Pipeline is a fully serverless system designed to automatically process audio files uploaded to S3. The pipeline:

1. **Receives** audio files or text prompts via S3 upload
2. **Validates** input format and structure
3. **Processes** audio using AWS Polly for text-to-speech synthesis
4. **Tracks** processing state in DynamoDB
5. **Notifies** on completion or failure via SNS
6. **Stores** processed output in S3 with full metadata

### Key Capabilities

- ✅ **Event-Driven Architecture**: Fully asynchronous, decoupled components
- ✅ **Automatic Validation**: Input validation with clear error messages
- ✅ **Error Handling**: Comprehensive error handling with retry policies
- ✅ **Observability**: X-Ray tracing, CloudWatch Logs, and CloudWatch Alarms
- ✅ **Security**: Encryption at rest/transit, least-privilege IAM, public access blocks
- ✅ **Multi-Environment**: Support for dev, stage, and prod environments
- ✅ **Cost-Optimized**: Pay-per-use pricing with intelligent resource sizing

## Architecture

The system implements a complete event-driven architecture with the following flow:

```
User Upload → S3 Input Bucket → EventBridge → Step Functions State Machine
                                                       ↓
                        ┌──────────────────────────────┴───────────────────────────┐
                        ↓                                                          ↓
              Success Path:                                            Failure Path:
        1. Write Initial Metadata (DynamoDB)                   1. Update Metadata (FAILED)
        2. Invoke Lambda (Validation)                          2. Publish SNS (Failure Topic)
        3. Polly Task (Text-to-Speech)
        4. Update Metadata (COMPLETED)
        5. Publish SNS (Success Topic)
```

### Components

- **S3 Input/Output Buckets**: Encrypted, versioned storage with EventBridge notifications
- **EventBridge**: Event filtering and routing for S3 Object Created events
- **Step Functions State Machine**: Orchestrates the complete processing workflow
- **Lambda Function**: Python-based audio processor with input validation
- **DynamoDB Table**: Metadata tracking with status (PROCESSING/COMPLETED/FAILED)
- **SNS Topics**: Encrypted notifications for success and failure scenarios
- **CloudWatch**: Comprehensive logging, metrics, and alarms
- **X-Ray**: Distributed tracing for end-to-end visibility

For detailed architecture information including Mermaid diagrams, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Features

### Implemented (Issues #1-12)

- ✅ **Issue #1-2**: S3 buckets with encryption, versioning, and EventBridge notifications
- ✅ **Issue #3**: EventBridge rule for S3 Object Created events
- ✅ **Issue #4**: Step Functions state machine with Polly integration
- ✅ **Issue #5**: DynamoDB table for metadata tracking
- ✅ **Issue #6**: SNS topics for success/failure notifications with error handling
- ✅ **Issue #7**: Lambda function for audio processing and validation
- ✅ **Issue #8**: Complete pipeline integration with input validation
- ✅ **Issue #9**: Multi-environment support and deployment pipeline
- ✅ **Issue #10**: Advanced error handling, retry policies, and observability
- ✅ **Issue #11**: Audio processing logic and output handling
- ✅ **Issue #12**: End-to-end validation and documentation polish

### Future Enhancements

- 🔮 Real-time audio streaming with Amazon Kinesis
- 🔮 ML-powered personalization with SageMaker
- 🔮 Multi-region deployment with Route 53
- 🔮 RESTful API with API Gateway
- 🔮 GraphQL API with AWS AppSync

## Project Philosophy

This project is built with **TDD-first, issue-driven development**. Every feature follows the strict TDD cycle:

### Strict TDD Rules

1. **🔴 RED**: Write a failing test first - never write production code without a failing test
2. **🟢 GREEN**: Write the minimal code to make the test pass
3. **🔵 REFACTOR**: Improve code quality while keeping all tests green
4. **📋 ARCHITECTURE**: Keep `ARCHITECTURE.md` and its Mermaid diagram in sync with every change
5. **✅ VERIFY**: All tests must pass and `cdk synth` must succeed before committing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed TDD guidelines and development workflow.

## Getting Started

### Prerequisites

- **Go**: 1.21 or later
- **Node.js**: 18.x or later (for AWS CDK)
- **AWS CDK**: 2.x installed globally (`npm install -g aws-cdk`)
- **AWS CLI**: Configured with appropriate credentials
- **Python**: 3.12 or later (for Lambda function)

### Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd cdk-sleep-go-qdev
   ```

2. **Install Go dependencies**:
   ```bash
   go mod download
   ```

3. **Install Python dependencies** (for Lambda):
   ```bash
   cd lambda/audio-processor
   pip install -r requirements.txt
   cd ../..
   ```

4. **Bootstrap CDK** (first time only):
   ```bash
   cdk bootstrap aws://<account-id>/<region>
   ```

### Configuration

Environment-specific configuration is managed via CDK context in `cdk.json`:

```json
{
  "dev": {
    "account": "111111111111",
    "region": "us-east-1",
    "logRetentionDays": 7
  },
  "prod": {
    "account": "222222222222",
    "region": "us-east-1",
    "logRetentionDays": 90
  }
}
```

## Deployment

### Deploy to Development

```bash
# Synthesize CloudFormation template
cdk synth --context environment=dev

# View changes before deployment
cdk diff --context environment=dev

# Deploy to dev environment
cdk deploy --context environment=dev
```

### Deploy to Production

```bash
# Deploy to production
cdk deploy --context environment=prod
```

### Deploy CI/CD Pipeline

```bash
# Deploy the CDK Pipeline (automated deployments)
cdk deploy PipelineStack
```

Once deployed, the pipeline automatically:
1. Synthesizes on code commit to `main` branch
2. Runs all tests
3. Deploys to dev environment
4. Deploys to stage environment
5. Waits for manual approval
6. Deploys to production (after approval)

## Usage

### Uploading Audio Files

Upload audio files or text files to the S3 input bucket:

```bash
# Get the input bucket name from stack outputs
aws cloudformation describe-stacks --stack-name SleepAudioStack-dev \
  --query "Stacks[0].Outputs[?OutputKey=='InputBucketName'].OutputValue" --output text

# Upload an audio file
aws s3 cp my-audio.mp3 s3://<input-bucket-name>/uploads/

# Upload a text file (for Polly synthesis)
aws s3 cp meditation-script.txt s3://<input-bucket-name>/uploads/
```

### Monitoring Processing

Monitor state machine executions:

```bash
# List recent executions
aws stepfunctions list-executions \
  --state-machine-arn <state-machine-arn> \
  --max-results 10

# Describe execution details
aws stepfunctions describe-execution \
  --execution-arn <execution-arn>
```

### Checking Metadata

Query DynamoDB for processing status:

```bash
# Get item from DynamoDB
aws dynamodb get-item \
  --table-name <metadata-table-name> \
  --key '{"audioId": {"S": "uploads/my-audio.mp3"}}'
```

### Subscribing to Notifications

Subscribe to SNS topics for notifications:

```bash
# Subscribe email to success topic
aws sns subscribe \
  --topic-arn <success-topic-arn> \
  --protocol email \
  --notification-endpoint <your-email>

# Subscribe email to failure topic
aws sns subscribe \
  --topic-arn <failure-topic-arn> \
  --protocol email \
  --notification-endpoint <ops-team-email>
```

## Testing

### Run All Tests

```bash
# Run all unit tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
```

### Test Structure

The project includes comprehensive tests organized by issue:

- **Issue #1-4**: Foundation (S3, EventBridge, Step Functions, Polly)
- **Issue #5**: DynamoDB integration
- **Issue #6**: SNS notifications and error handling
- **Issue #7**: Lambda function integration
- **Issue #8**: Complete pipeline integration
- **Issue #9**: Multi-environment and deployment
- **Issue #10**: Retry policies and observability
- **Issue #11**: Audio processing and output handling
- **Issue #12**: End-to-end validation

### Useful CDK Commands

```bash
# Synthesize CloudFormation template
cdk synth

# View differences with deployed stack
cdk diff

# Deploy to AWS
cdk deploy

# Destroy deployed stack
cdk destroy

# List all stacks
cdk list
```

## Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)**: Detailed architecture documentation with Mermaid diagrams
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Development guidelines and TDD workflow
- **[SUMMARY.md](SUMMARY.md)**: Project summary, key decisions, and lessons learned
- **[.github/AGENT_GUIDELINES.md](.github/AGENT_GUIDELINES.md)**: AI agent development guidelines

## Contributing

This project follows strict TDD principles. All contributions must:

1. Include failing tests first
2. Implement minimal code to pass tests
3. Maintain or improve test coverage
4. Update architecture documentation
5. Follow existing code patterns

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.

## License

See [LICENSE](LICENSE) file for details.

## Project Status

**Status**: ✅ Complete (Issue #12)

This project has completed core development through Issue #12. All major features are implemented, tested, and documented. The pipeline is ready for deployment and further experimentation.

---

**Built with strict TDD principles • AWS CDK • Go • Event-Driven Architecture**
