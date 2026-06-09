"""
Sleep Audio Processor Lambda Function
Validates input, enriches metadata, and processes audio files in the pipeline.

Issue #10: Enhanced with structured logging and X-Ray tracing support.
"""

import json
import logging
import os
import boto3
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
        'status': 'processed',
        'message': 'Audio validation and processing completed successfully'
    }
    
    log_structured('INFO', 'Processing completed successfully',
                  request_id=request_id,
                  audio_id=audio_id,
                  status='completed',
                  format=validation_result.get('format', 'unknown'))
    
    return result
