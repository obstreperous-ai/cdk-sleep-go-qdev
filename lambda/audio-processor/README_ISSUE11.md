# Lambda Handler Implementation Notes - Issue #11

## Current State
The handler.py file has input validation implemented but needs core audio processing.

## Syntax Error to Fix (Lines 59-60)
WRONG:
```python
def validate_input(bucket, audio_id):
@xray_recorder.capture('validate_input')
```

CORRECT:
```python
@xray_recorder.capture('validate_input')
def validate_input(bucket, audio_id):
```

## Required Additions

### 1. After line 11 (`import boto3`), add:
```python
import time
from datetime import datetime
from botocore.exceptions import ClientError
```

### 2. After line 23 (`ddb = boto3.resource('dynamodb')`), add:
```python
s3_client = boto3.client('s3')
polly_client = boto3.client('polly')
```

### 3. After line 26 (`table_name = os.environ.get('METADATA_TABLE_NAME', '')`), add:
```python
output_bucket_name = os.environ.get('OUTPUT_BUCKET_NAME', '')
```

### 4. After line 92 (end of `validate_input` function), add 5 new functions:
- `download_s3_file(bucket, key)` - Downloads from S3
- `polly_synthesize(text, voice='Joanna')` - Polly TTS
- `upload_s3_file(bucket, key, content)` - Uploads to S3
- `update_metadata(tbl, audio_id, output_info)` - Updates DynamoDB
- `process_audio_file(bucket, audio_id, file_format)` - Core processing pipeline
- `handle_output(audio_id, audio_data, start_time)` - Output handling

See IMPLEMENTATION_GUIDE_ISSUE11.md in project root for complete function implementations.
