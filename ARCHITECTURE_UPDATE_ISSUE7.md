# Architecture Update - Issue #7: Lambda Function Integration

## Overview
This document describes the Lambda function integration completed in Issue #7 following strict TDD principles.

## Changes Summary

### 1. Lambda Function - SleepAudioProcessor

**New Component Added**: Lambda function for audio processing, metadata enrichment, and validation.

**Configuration**:
- **Function Name**: SleepAudioProcessor
- **Runtime**: Python 3.12
- **Handler**: handler.lambda_handler
- **Timeout**: 30 seconds
- **Memory**: 256 MB
- **Code Location**: `lambda/audio-processor/`

**Environment Variables**:
- `METADATA_TABLE_NAME`: DynamoDB table name for metadata storage

**Current Capabilities** (Placeholder Implementation):
- Receives S3 event details (bucket, audioId) from Step Functions
- Logs incoming events for debugging and audit trail
- Basic input validation (bucket and audioId presence)
- Returns processing status and enriched metadata

**Future Enhancements** (Documented but not yet implemented):
- Audio format validation (MP3, WAV, FLAC, etc.)
- Metadata extraction (duration, bitrate, sample rate, codec)
- Audio quality analysis
- Content validation and integrity checks

**IAM Permissions** (Least Privilege):
- **DynamoDB**: `GetItem`, `PutItem`, `UpdateItem` on SleepAudioMetadataTable
- **S3**: `GetObject` on SleepAudioInputBucket
- **CloudWatch Logs**: `CreateLogGroup`, `CreateLogStream`, `PutLogEvents`

### 2. Step Functions Integration

**Updated Workflow**:
```
Write Initial Metadata (DynamoDB)
    ↓
Invoke SleepAudioProcessor (Lambda) ← NEW STEP
    ↓
Polly Task (synthesizeSpeech)
    ↓
Update Metadata Success (DynamoDB)
    ↓
Publish Success (SNS)

(Error Path from Polly)
Update Metadata Failure (DynamoDB) → Publish Failure (SNS)
```

**New Task**:
- Task Name: `InvokeSleepAudioProcessor`
- Type: `LambdaInvoke`
- Result Path: `$.processorResult`
- Input Payload: S3 event details (bucket, audioId) from state machine context

**Updated IAM Permissions**:
- State machine execution role granted `lambda:InvokeFunction` permission

### 3. Tests Added (TDD Approach)

Six new tests added to `cdk-base_test.go`:
- Lambda function existence, configuration, execution role
- Lambda IAM permissions for DynamoDB
- State machine Lambda invocation task
- State machine IAM permissions for Lambda invocation
