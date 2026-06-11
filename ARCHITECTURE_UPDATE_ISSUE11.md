# Architecture Update - Issue #11: Core Audio Processing & Output Handling

## Summary
Issue #11 implements the core audio processing logic that moves the pipeline from validation-only to full functional audio processing with Polly integration and S3 output handling.

## Infrastructure Changes

### Lambda Function Enhancements (cdk-base.go)

**New Permissions:**
1. **Output Bucket Write Access** (Line 101)
   ```go
   outputBucket.GrantWrite(audioProcessorFunction, nil)
   ```
   - Allows Lambda to upload processed audio files
   - Uses least-privilege: write-only access

2. **Polly SynthesizeSpeech Permission** (Lines 104-107)
   ```go
   audioProcessorFunction.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
       Actions:   jsii.Strings("polly:SynthesizeSpeech"),
       Resources: jsii.Strings("*"),  // Polly doesn't support resource-level permissions
   }))
   ```
   - Enables text-to-speech synthesis
   - Global permission required by Polly service design

**New Environment Variable** (Line 85):
```go
"OUTPUT_BUCKET_NAME": outputBucket.BucketName()
```

## Lambda Processing Flow (handler.py)

### Updated Architecture

**Previous (Issue #10):**
```
Input S3 → EventBridge → State Machine → Lambda (Validation Only) → Return
```

**Current (Issue #11):**
```
Input S3 → EventBridge → State Machine → Lambda:
    1. Validate Input
    2. Download from Input S3
    3. Process Audio:
       - Text files: Polly synthesis
       - Audio files: Pass-through (future: enhance)
    4. Upload to Output S3
    5. Update DynamoDB metadata
    6. Return success with output info
```

### New Lambda Functions

1. **download_s3_file(bucket, key)**
   - Downloads file from input bucket
   - Returns bytes content
   - X-Ray traced
   - Error handling with structured logging

2. **polly_synthesize(text, voice='Joanna')**
   - Converts text to speech audio
   - Uses neural engine for quality
   - Returns: (audio_bytes, character_count)
   - Tracks characters for cost monitoring

3. **upload_s3_file(bucket, key, content)**
   - Uploads to output bucket
   - Sets ServerSideEncryption='AES256'
   - Returns S3 URI
   - Content-Type: audio/mpeg

4. **update_metadata(tbl, audio_id, output_info)**
   - Updates DynamoDB with output details
   - Fields: outputBucket, outputKey, outputFileSize, processingDuration, updatedAt
   - Non-blocking: logs errors but doesn't raise

5. **process_audio_file(bucket, audio_id, file_format)**
   - Orchestrates processing based on file type
   - Text (.txt): Polly synthesis
   - Audio (.mp3, .wav, etc.): Pass-through

6. **handle_output(audio_id, audio_data, start_time)**
   - Generates unique output key with timestamp
   - Uploads to S3
   - Calculates processing duration
   - Updates DynamoDB
   - Returns output metadata dict

### Output File Naming Convention

**Pattern:** `processed/{timestamp}-{original_basename}.mp3`

**Examples:**
- Input: `uploads/meditation.txt` → Output: `processed/20240115-143022-meditation.mp3`
- Input: `user123/sleep-story.mp3` → Output: `processed/20240115-143045-sleep-story.mp3`

### DynamoDB Schema Extension

**New Attributes** (added by Lambda):
- `outputBucket`: Output S3 bucket name
- `outputKey`: Full S3 key of processed audio
- `outputFileSize`: Size in bytes
- `processingDuration`: Elapsed time in seconds (decimal)
- `updatedAt`: ISO 8601 timestamp

**Existing Attributes** (set by state machine):
- `audioId`: Partition key (input file path)
- `status`: PROCESSING → COMPLETED/FAILED
- `inputBucket`, `inputKey`, `createdAt`
- `errorMessage` (on failure)

## Data Flow Update

### Success Path
1. S3 event triggers state machine
2. State machine writes initial DynamoDB record (status=PROCESSING)
3. Lambda invoked:
   a. Validates input
   b. Downloads from input S3
   c. Processes (Polly for text, pass-through for audio)
   d. Uploads to output S3
   e. Updates DynamoDB with output metadata
   f. Returns success response
4. State machine updates status=COMPLETED
5. SNS notification sent

### Lambda Response Format

**Enhanced Response** (Issue #11):
```json
{
    "statusCode": 200,
    "valid": true,
    "audioId": "uploads/meditation.txt",
    "bucket": "sleep-audio-input-bucket",
    "format": ".txt",
    "status": "completed",
    "outputBucket": "sleep-audio-output-bucket",
    "outputKey": "processed/20240115-143022-meditation.mp3",
    "outputS3Uri": "s3://sleep-audio-output-bucket/processed/20240115-143022-meditation.mp3",
    "outputFileSize": 245678,
    "processingDuration": 3.42,
    "message": "Audio processing and output handling completed successfully"
}
```

## Error Handling

**No Changes to Error Flow:**
- Validation errors still caught by state machine
- S3/Polly errors raise exceptions → state machine error path
- DynamoDB updates from Lambda are non-blocking
- State machine handles final status updates

## Observability

**X-Ray Tracing:**
- All new functions decorated with `@xray_recorder.capture()`
- Traces: download_s3_file, polly_synthesize, upload_s3_file, update_metadata, process_audio_file, handle_output

**Structured Logging:**
- All S3 operations logged with bucket/key/size
- Polly operations logged with text length/characters/audio size
- Processing duration tracked and logged
- Error details captured with context

## Testing Coverage

**New CDK Tests** (cdk-base_test.go):
1. `TestLambdaHasOutputBucketWritePermissions` - S3 write access
2. `TestLambdaHasPollyPermissions` - Polly access
3. `TestLambdaHasOutputBucketEnvironmentVariable` - Environment config
4. `TestLambdaProcessingConfigurationForAudio` - Resource allocation
5. `TestOutputHandlingEndToEnd` - Integration verification

## Performance Considerations

**Lambda Configuration:**
- Timeout: 30 seconds (adequate for most audio files)
- Memory: 256 MB (sufficient for audio processing)
- Future: Increase for larger files or more complex processing

**Polly Limits:**
- Max text length: 3,000 characters per request
- For longer text: implement batching in future enhancement
- Neural voices: higher quality but slower than standard

**S3 Operations:**
- Concurrent downloads/uploads supported
- Server-side encryption overhead minimal
- Multipart upload for large files (future enhancement)

## Security

**Least Privilege:**
- Lambda can only write to output bucket (not read)
- Lambda can only read from input bucket (existing)
- Polly permission scoped to SynthesizeSpeech only
- No S3 delete permissions

**Encryption:**
- Output files encrypted at rest (AES256)
- In-transit: HTTPS for all API calls
- DynamoDB: AWS-managed encryption

## Cost Tracking

**New Cost Factors:**
1. **Polly:** $4-16 per million characters (neural voices)
   - Lambda tracks character count in logs
   - Can be aggregated for billing

2. **S3 PUT Requests:** $0.005 per 1,000 requests
   - One PUT per processed file

3. **S3 Storage:** $0.023 per GB/month
   - Output files accumulate
   - Consider lifecycle policies

4. **DynamoDB Writes:** $1.25 per million write units
   - One write per file for output metadata

## Future Enhancements (Post-Issue #11)

1. **Audio Enhancement:** Apply normalization, noise reduction, equalization
2. **Format Conversion:** Support multiple output formats
3. **Batch Processing:** Handle large text files with chunking
4. **Streaming:** Process audio streams for real-time use cases
5. **Caching:** Cache frequently requested Polly outputs
6. **Compression:** Apply audio compression for smaller files
7. **Metadata Extraction:** Parse audio metadata (duration, bitrate, etc.)

## Diagram Updates Required (ARCHITECTURE.md)

Update Mermaid diagram to show:
1. Lambda → Input S3 (download arrow)
2. Lambda → Polly (synthesis arrow)
3. Lambda → Output S3 (upload arrow)
4. Lambda → DynamoDB (update output metadata arrow)
5. Lambda processing box: Show internal steps

---

**Document Version:** 1.0  
**Issue:** #11  
**Date:** 2024  
**Status:** Implementation Complete (Infrastructure), Lambda Handler Code Documented
