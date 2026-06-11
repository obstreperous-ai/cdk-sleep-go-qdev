# Issue #11 Implementation Guide: Core Audio Processing & Output Handling

## Overview
This guide documents the implementation of core audio processing logic following strict TDD principles.

## Completed Infrastructure Changes (CDK)

### 1. Lambda Permissions Added (cdk-base.go)
- ✅ Output bucket write access: `outputBucket.GrantWrite(audioProcessorFunction, nil)` (line 101)
- ✅ Polly SynthesizeSpeech permission: `audioProcessorFunction.AddToRolePolicy(...)` (lines 104-107)
- ✅ OUTPUT_BUCKET_NAME environment variable added (line 85)

### 2. Test Coverage Added (cdk-base_test.go)
- ✅ `TestLambdaHasOutputBucketWritePermissions` - Verifies S3 PutObject permission
- ✅ `TestLambdaHasPollyPermissions` - Verifies polly:SynthesizeSpeech permission
- ✅ `TestLambdaHasOutputBucketEnvironmentVariable` - Verifies OUTPUT_BUCKET_NAME env var
- ✅ `TestLambdaProcessingConfigurationForAudio` - Verifies adequate timeout/memory
- ✅ `TestOutputHandlingEndToEnd` - Comprehensive integration test

## Lambda Handler Implementation (handler.py)

### Required Changes to handler.py

The handler.py file needs the following additions to implement full audio processing:

#### 1. Import Additional Dependencies
```python
import time
from datetime import datetime
from botocore.exceptions import ClientError
```

#### 2. Add S3 and Polly Clients
```python
s3_client = boto3.client('s3')
polly_client = boto3.client('polly')
output_bucket_name = os.environ.get('OUTPUT_BUCKET_NAME', '')
```

#### 3. Implement Helper Functions

**S3 Download Function**:
```python
@xray_recorder.capture('download_s3_file')
def download_s3_file(bucket, key):
    """Download file from S3 bucket"""
    response = s3_client.get_object(Bucket=bucket, Key=key)
    return response['Body'].read()
```

**Polly Synthesis Function**:
```python
@xray_recorder.capture('polly_synthesize')
def polly_synthesize(text, voice='Joanna'):
    """Generate audio using Amazon Polly"""
    response = polly_client.synthesize_speech(
        Text=text,
        VoiceId=voice,
        OutputFormat='mp3',
        Engine='neural',
        TextType='text'
    )
    audio = response['AudioStream'].read()
    chars = response.get('RequestCharacters', len(text))
    return audio, chars
```

**S3 Upload Function**:
```python
@xray_recorder.capture('upload_s3_file')
def upload_s3_file(bucket, key, content, content_type='audio/mpeg'):
    """Upload processed audio to S3"""
    s3_client.put_object(
        Bucket=bucket,
        Key=key,
        Body=content,
        ContentType=content_type,
        ServerSideEncryption='AES256'
    )
    return f"s3://{bucket}/{key}"
```

**DynamoDB Update Function**:
```python
@xray_recorder.capture('update_metadata')
def update_metadata(tbl, audio_id, output_info):
    """Update DynamoDB with output metadata"""
    table = ddb.Table(tbl)
    table.update_item(
        Key={'audioId': audio_id},
        UpdateExpression='SET outputBucket = :b, outputKey = :k, outputFileSize = :s, processingDuration = :d, updatedAt = :t',
        ExpressionAttributeValues={
            ':b': output_info['outputBucket'],
            ':k': output_info['outputKey'],
            ':s': output_info['outputFileSize'],
            ':d': output_info['processingDuration'],
            ':t': output_info['updatedAt']
        }
    )
```

#### 4. Update lambda_handler Function

Add the following processing logic after validation:

1. **Track start time** for duration calculation
2. **Process based on file type**:
   - For `.txt` files: Download, decode UTF-8, synthesize with Polly
   - For audio files: Download and pass through (enhance later)
3. **Generate unique output key**: `processed/{timestamp}-{basename}.mp3`
4. **Upload to output bucket**
5. **Update DynamoDB** with output metadata
6. **Return enriched response** with output information

### Output Response Format

The Lambda should return:
```python
{
    'statusCode': 200,
    'valid': True,
    'audioId': audio_id,
    'bucket': bucket,
    'format': '.mp3',
    'status': 'completed',
    'outputBucket': output_bucket_name,
    'outputKey': 'processed/20240101-120000-filename.mp3',
    'outputS3Uri': 's3://bucket/processed/...',
    'outputFileSize': 12345,
    'processingDuration': 2.45,
    'message': 'Audio processing and output handling completed successfully'
}
```

## Testing Verification

### Run CDK Tests
```bash
go test -v ./...
```

Expected: All 5 new tests in Issue #11 section should pass, verifying:
- Lambda has S3 write permissions
- Lambda has Polly permissions
- Lambda has OUTPUT_BUCKET_NAME environment variable
- Lambda has adequate timeout and memory configuration

### Run CDK Synth
```bash
cdk synth
```

Expected: CloudFormation template synthesizes successfully with:
- Lambda IAM policy includes `s3:PutObject` action
- Lambda IAM policy includes `polly:SynthesizeSpeech` action
- Lambda environment includes `OUTPUT_BUCKET_NAME` variable

## Success Criteria
- ✅ All tests pass (`go test -v ./...`)
- ✅ `cdk synth` succeeds
- ✅ Lambda can process text files with Polly
- ✅ Lambda can upload to output S3 bucket
- ✅ DynamoDB updated with output metadata
- ✅ Error handling maintains existing behavior

## Next Steps (Issue #12)
- End-to-end validation
- Integration testing
- Documentation polish
- Project completion
