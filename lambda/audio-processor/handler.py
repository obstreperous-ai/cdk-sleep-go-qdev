"""
Sleep Audio Processor Lambda Function
Validates input, enriches metadata, and processes audio files in the pipeline.

Issue #10: Enhanced with structured logging and X-Ray tracing support.
Issue #11: Core audio processing logic implementation.
"""

import json
import logging
import os
import boto3
from datetime import datetime
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all

# Patch AWS SDK clients for X-Ray tracing
patch_all()

# Setup structured logger with JSON formatting
log = logging.getLogger()
log.setLevel(logging.INFO)

# DynamoDB client
ddb = boto3.resource('dynamodb')

# Environment variables
table_name = os.environ.get('METADATA_TABLE_NAME', '')
output_bucket_name = os.environ.get('OUTPUT_BUCKET_NAME', '')


def log_structured(level, message, **kwargs):
    """
    Structured logging helper that outputs JSON format.
    
    Args:
        level: Log level (INFO, ERROR, WARNING, etc.)
        message: Log message
        **kwargs: Additional context fields to include in log
    """
    log_entry = {
        'message': message,
        'level': level,
        **kwargs
    }
    
    if level == 'ERROR':
        log.error(json.dumps(log_entry))
    else:
        log.info(json.dumps(log_entry))


# Supported audio file extensions for validation
SUPPORTED_AUDIO_FORMATS = ['.mp3', '.wav', '.m4a', '.flac', '.ogg', '.txt']


class ValidationError(Exception):
    """Custom exception for input validation errors"""
    pass


@xray_recorder.capture('download_from_s3')
def download_audio_from_s3(bkt, obj_key):
    """Download file from S3"""
    s3_svc = boto3.client('s3')
    try:
        obj_response = s3_svc.get_object(Bucket=bkt, Key=obj_key)
        file_bytes = obj_response['Body'].read()
        log_structured('INFO', 'S3 download success', bucket=bkt, key=obj_key, bytes=len(file_bytes))
        return file_bytes
    except Exception as download_err:
        log_structured('ERROR', 'S3 download error', bucket=bkt, key=obj_key, err=str(download_err))
        raise


@xray_recorder.capture('polly_synthesis')
def synthesize_audio_with_polly(input_text, voice='Joanna', eng='neural'):
    """Use Polly to create audio"""
    polly_svc = boto3.client('polly')
    try:
        polly_response = polly_svc.synthesize_speech(
            Text=input_text,
            VoiceId=voice,
            OutputFormat='mp3',
            Engine=eng
        )
        audio_bytes = polly_response['AudioStream'].read()
        mime_type = polly_response.get('ContentType', 'audio/mpeg')
        calc_duration = int((len(input_text) / 5) / 150 * 60)
        log_structured('INFO', 'Polly success', voice=voice, bytes=len(audio_bytes))
        return {
            'audio_data': audio_bytes,
            'content_type': mime_type,
            'duration': calc_duration
        }
    except Exception as polly_err:
        log_structured('ERROR', 'Polly error', voice=voice, err=str(polly_err))
        raise


@xray_recorder.capture('s3_upload')
def upload_to_output_bucket(data_bytes, s3_key, mime='audio/mpeg'):
    """Upload to S3 output bucket"""
    if not output_bucket_name:
        raise Exception("No OUTPUT_BUCKET_NAME configured")
    s3_svc = boto3.client('s3')
    try:
        s3_svc.put_object(
            Bucket=output_bucket_name,
            Key=s3_key,
            Body=data_bytes,
            ContentType=mime
        )
        s3_uri = f's3://{output_bucket_name}/{s3_key}'
        log_structured('INFO', 'S3 upload success', bucket=output_bucket_name, key=s3_key)
        return {'output_uri': s3_uri, 'output_key': s3_key}
    except Exception as upload_err:
        log_structured('ERROR', 'S3 upload error', key=s3_key, err=str(upload_err))
        raise


@xray_recorder.capture('ddb_metadata_update')
def update_metadata_with_output(item_id, metadata_dict):
    """Update DynamoDB with output info"""
    ddb_table = ddb.Table(table_name)
    ddb_table.update_item(
        Key={'audioId': item_id},
        UpdateExpression='SET outputUri = :uri_val, outputKey = :key_val, fileSize = :size_val, duration = :dur_val',
        ExpressionAttributeValues={
            ':uri_val': metadata_dict.get('output_uri', ''),
            ':key_val': metadata_dict.get('output_key', ''),
            ':size_val': metadata_dict.get('file_size', 0),
            ':dur_val': metadata_dict.get('duration', 0)
        }
    )


def generate_output_key(input_path):
    """Create output key with timestamp"""
    base_name = input_path.split('/')[-1]
    without_ext = base_name.rsplit('.', 1)[0]
    file_ext = base_name.rsplit('.', 1)[1] if '.' in base_name else 'mp3'
    time_stamp = datetime.utcnow().strftime('%Y%m%d-%H%M%S')
    return f'processed/{without_ext}-{time_stamp}.{file_ext}'


def validate_input(bucket, audio_id):
@xray_recorder.capture('validate_input')
    """
    Validates the input parameters and file metadata.
    
    Args:
        bucket: S3 bucket name
        audio_id: S3 object key (audio file path)
        
    Raises:
        ValidationError: If validation fails
        
    Returns:
        dict: Validation result with metadata
    """
    # Validate required fields are present
    if not bucket or not audio_id:
        raise ValidationError("Missing required fields: bucket and audioId are required")
    
    # Validate bucket name format (basic sanity check)
    if len(bucket) < 3 or len(bucket) > 63:
        raise ValidationError(f"Invalid bucket name: {bucket}")
    
    # Validate file extension
    file_extension = None
    for ext in SUPPORTED_AUDIO_FORMATS:
        if audio_id.lower().endswith(ext):
            file_extension = ext
            break
    
    if not file_extension:
        raise ValidationError(f"Unsupported file format. audioId must end with one of: {', '.join(SUPPORTED_AUDIO_FORMATS)}")
    
    return {'valid': True, 'format': file_extension, 'audioId': audio_id}

@xray_recorder.capture('lambda_handler')
def lambda_handler(event, context):
    """
    Main Lambda handler for audio processing.
    
    Receives:
        - S3 event details (bucket, object key)
        - audioId from state machine
        
    Returns:
        - Processing status and enriched metadata (on success)
        - Error details (on validation failure)
        
    Raises:
        Exception: For validation errors (caught by Step Functions error handling)
    """
    
    # Get request ID from context for tracing
    request_id = context.request_id if context else 'no-request-id'
    
    # Structured logging for incoming request
    log_structured('INFO', 'Lambda invocation started', 
                   request_id=request_id,
                   function_name=context.function_name if context else 'unknown',
                   function_version=context.function_version if context else 'unknown')
    
    # Add custom X-Ray annotations for easier searching
    xray_recorder.put_annotation('request_id', request_id)
    
    try:
        # Extract data from event
        # Handle Step Functions format with 'detail' key
        if 'detail' in event:
            bucket = event['detail'].get('bucket', {}).get('name', '')
            audio_id = event['detail'].get('object', {}).get('key', '')
        else:
            bucket = event.get('bucket', '')
            audio_id = event.get('audioId', '')
        
        # Add X-Ray metadata for this request
        xray_recorder.put_metadata('bucket', bucket, 'audio_processing')
        xray_recorder.put_metadata('audio_id', audio_id, 'audio_processing')
        
        log_structured('INFO', 'Processing audio file',
                      request_id=request_id,
                      bucket=bucket,
                      audio_id=audio_id,
                      status='validating')
        
        # Perform input validation
        # This will raise ValidationError if input is invalid
        validation_result = validate_input(bucket, audio_id)
        
        log_structured('INFO', 'Validation completed successfully',
                      request_id=request_id,
                      audio_id=audio_id,
                      format=validation_result.get('format', 'unknown'),
                      status='validated')
        
        # Additional validation: Check if file exists in S3 (optional but recommended)
        # For now, we'll skip the actual S3 head object call to avoid additional API calls
        # In production, you would add:
        # s3_client = boto3.client('s3')
        # try:
        #     s3_client.head_object(Bucket=bucket, Key=audio_id)
        # except ClientError as e:
        #     if e.response['Error']['Code'] == '404':
        #         raise ValidationError(f"File not found: {audio_id}")
        #     raise
        
    except ValidationError as e:
        # ========== Issue #11: Audio Processing ==========
        
        file_format = validation_result.get('format', '')
        
        if file_format == '.txt':
            log_structured('INFO', 'Text input detected',
                          request_id=request_id,
                          audio_id=audio_id)
            text_bytes = download_audio_from_s3(bucket, audio_id)
            tts_input = text_bytes.decode('utf-8')
        else:
            log_structured('INFO', 'Audio input detected',
                          request_id=request_id,
                          audio_id=audio_id)
            download_audio_from_s3(bucket, audio_id)
            tts_input = "Welcome to your peaceful sleep session. Let the calming sounds guide you into deep, restful sleep. Breathe slowly and deeply. Release all tension. You are safe, relaxed, and at peace."
        
        log_structured('INFO', 'Starting synthesis',
                      request_id=request_id,
                      audio_id=audio_id)
        
        synth_output = synthesize_audio_with_polly(
            input_text=tts_input[:3000],
            voice='Joanna',
            eng='neural'
        )
        
        output_path = generate_output_key(audio_id)
        
        log_structured('INFO', 'Uploading result',
                      request_id=request_id,
                      audio_id=audio_id,
                      output_key=output_path)
        
        upload_info = upload_to_output_bucket(
            data_bytes=synth_output['audio_data'],
            s3_key=output_path,
            mime=synth_output['content_type']
        )
        
        metadata_update = {
            **upload_info,
            'file_size': len(synth_output['audio_data']),
            'duration': synth_output['duration']
        }
        
        update_metadata_with_output(audio_id, metadata_update)
        
        # Log validation error and raise exception for Step Functions to catch
        log_structured('ERROR', 'Validation failed',
                      request_id=request_id,
                      error_type='ValidationError',
                      error_message=str(e),
                      status='failed')
        # Raise exception so Step Functions can catch it and route to failure path
        raise Exception(f"ValidationError: {str(e)}")
    
    except Exception as e:
        # Log unexpected errors
        log_structured('ERROR', 'Unexpected processing error',
                      request_id=request_id,
                      error_type='ProcessingError',
                      error_message=str(e),
                      status='failed')
        # Re-raise to trigger Step Functions error handling
        raise Exception(f"ProcessingError: {str(e)}")
    
    # Placeholder: Future processing logic goes here
    # - Validate audio format
    # - Extract metadata
    # - Enrich with additional data
    
    # Return success with metadata
    result = {
        'statusCode': 200,
        'valid': True,
        'audioId': audio_id,
        'bucket': bucket,
        'format': validation_result.get('format', 'unknown'),
        'status': 'completed',
        'message': 'Audio processing completed successfully',
        'outputUri': upload_info['output_uri'],
        'outputKey': upload_info['output_key'],
        'duration': synth_output['duration'],
        'fileSize': len(synth_output['audio_data'])
    }
    
    log_structured('INFO', 'Processing completed successfully',
                  request_id=request_id,
                  audio_id=audio_id,
                  output_uri=upload_info['output_uri'],
                  status='completed')
    
    return result
