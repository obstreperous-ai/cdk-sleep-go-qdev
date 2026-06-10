"""
Unit tests for Sleep Audio Processor Lambda Function (Issue #11)
Tests audio processing logic: download, process, upload, and metadata updates.
Following strict TDD: Tests written first, then implementation.
"""

import json
import unittest
from unittest.mock import Mock, patch, MagicMock
from io import BytesIO
import handler


class TestAudioProcessing(unittest.TestCase):
    """Test audio processing logic"""

    def setUp(self):
        """Set up test fixtures"""
        self.test_event = {
            'detail': {
                'bucket': {'name': 'test-input-bucket'},
                'object': {'key': 'test-audio.mp3'}
            },
            'bucket': 'test-input-bucket',
            'audioId': 'test-audio.mp3'
        }
        self.test_context = Mock()
        self.test_context.request_id = 'test-request-id'
        self.test_context.function_name = 'test-function'
        self.test_context.function_version = '1'

    @patch.dict('os.environ', {
        'METADATA_TABLE_NAME': 'test-table',
        'OUTPUT_BUCKET_NAME': 'test-output-bucket'
    })
    @patch('handler.boto3')
    def test_download_audio_from_s3(self, mock_boto3):
        """Test downloading audio file from S3 input bucket"""
        # Mock S3 client
        mock_s3 = Mock()
        mock_boto3.client.return_value = mock_s3
        
        # Mock S3 download
        test_audio_data = b'fake audio data'
        mock_s3.get_object.return_value = {
            'Body': BytesIO(test_audio_data),
            'ContentLength': len(test_audio_data)
        }
        
        # Call download function
        result = handler.download_audio_from_s3('test-bucket', 'test-key')
        
        # Verify S3 client was called correctly
        mock_s3.get_object.assert_called_once_with(
            Bucket='test-bucket',
            Key='test-key'
        )
        
        # Verify audio data was returned
        self.assertEqual(result, test_audio_data)

    @patch.dict('os.environ', {'OUTPUT_BUCKET_NAME': 'test-output-bucket'})
    @patch('handler.boto3')
    def test_synthesize_audio_with_polly(self, mock_boto3):
        """Test synthesizing audio using Amazon Polly"""
        # Mock Polly client
        mock_polly = Mock()
        mock_boto3.client.return_value = mock_polly
        
        # Mock Polly response
        test_audio_stream = b'synthesized audio data'
        mock_polly.synthesize_speech.return_value = {
            'AudioStream': BytesIO(test_audio_stream),
            'ContentType': 'audio/mpeg',
            'RequestCharacters': 50
        }
        
        # Call synthesis function
        result = handler.synthesize_audio_with_polly(
            text='Test sleep audio',
            voice_id='Joanna'
        )
        
        # Verify Polly was called correctly
        mock_polly.synthesize_speech.assert_called_once()
        call_args = mock_polly.synthesize_speech.call_args[1]
        self.assertEqual(call_args['Text'], 'Test sleep audio')
        self.assertEqual(call_args['VoiceId'], 'Joanna')
        self.assertEqual(call_args['OutputFormat'], 'mp3')
        self.assertEqual(call_args['Engine'], 'neural')
        
        # Verify audio data was returned
        self.assertEqual(result['audio_data'], test_audio_stream)
        self.assertEqual(result['content_type'], 'audio/mpeg')

    @patch.dict('os.environ', {'OUTPUT_BUCKET_NAME': 'test-output-bucket'})
    @patch('handler.boto3')
    def test_upload_to_output_bucket(self, mock_boto3):
        """Test uploading processed audio to S3 output bucket"""
        # Mock S3 client
        mock_s3 = Mock()
        mock_boto3.client.return_value = mock_s3
        
        # Mock S3 upload response
        mock_s3.put_object.return_value = {
            'ETag': '"test-etag"',
            'VersionId': 'test-version'
        }
        
        # Test data
        test_audio_data = b'processed audio data'
        output_key = 'processed/test-audio-processed.mp3'
        
        # Call upload function
        result = handler.upload_to_output_bucket(
            audio_data=test_audio_data,
            output_key=output_key,
            content_type='audio/mpeg'
        )
        
        # Verify S3 put_object was called correctly
        mock_s3.put_object.assert_called_once()
        call_args = mock_s3.put_object.call_args[1]
        self.assertEqual(call_args['Bucket'], 'test-output-bucket')
        self.assertEqual(call_args['Key'], output_key)
        self.assertEqual(call_args['Body'], test_audio_data)
        self.assertEqual(call_args['ContentType'], 'audio/mpeg')
        
        # Verify result includes output location
        self.assertIn('output_uri', result)
        self.assertIn('test-output-bucket', result['output_uri'])

    @patch.dict('os.environ', {'METADATA_TABLE_NAME': 'test-table'})
    @patch('handler.boto3')
    def test_update_metadata_with_output(self, mock_boto3):
        """Test updating DynamoDB metadata with output information"""
        # Mock DynamoDB resource
        mock_ddb = Mock()
        mock_table = Mock()
        mock_boto3.resource.return_value = mock_ddb
        mock_ddb.Table.return_value = mock_table
        
        # Test metadata
        audio_id = 'test-audio.mp3'
        output_info = {
            'output_uri': 's3://test-output-bucket/processed/test-audio.mp3',
            'output_key': 'processed/test-audio.mp3',
            'duration': 120,
            'file_size': 2048000
        }
        
        # Call update function
        handler.update_metadata_with_output(audio_id, output_info)
        
        # Verify DynamoDB update was called
        mock_table.update_item.assert_called_once()
        call_args = mock_table.update_item.call_args[1]
        self.assertEqual(call_args['Key']['audioId'], audio_id)
        self.assertIn('outputUri', call_args['UpdateExpression'])
        self.assertIn('outputKey', call_args['UpdateExpression'])

    @patch.dict('os.environ', {
        'METADATA_TABLE_NAME': 'test-table',
        'OUTPUT_BUCKET_NAME': 'test-output-bucket'
    })
    @patch('handler.download_audio_from_s3')
    @patch('handler.synthesize_audio_with_polly')
    @patch('handler.upload_to_output_bucket')
    @patch('handler.update_metadata_with_output')
    @patch('handler.validate_input')
    def test_end_to_end_audio_processing(
        self,
        mock_validate,
        mock_update_metadata,
        mock_upload,
        mock_synthesize,
        mock_download
    ):
        """Test complete end-to-end audio processing flow"""
        # Mock validation
        mock_validate.return_value = {
            'valid': True,
            'format': '.mp3',
            'audioId': 'test-audio.mp3'
        }
        
        # Mock download
        mock_download.return_value = b'input audio data'
        
        # Mock Polly synthesis
        mock_synthesize.return_value = {
            'audio_data': b'synthesized audio data',
            'content_type': 'audio/mpeg',
            'duration': 120
        }
        
        # Mock upload
        mock_upload.return_value = {
            'output_uri': 's3://test-output-bucket/processed/test-audio.mp3',
            'output_key': 'processed/test-audio.mp3'
        }
        
        # Call Lambda handler
        result = handler.lambda_handler(self.test_event, self.test_context)
        
        # Verify result
        self.assertEqual(result['statusCode'], 200)
        self.assertEqual(result['status'], 'completed')
        self.assertIn('outputUri', result)
        self.assertIn('outputKey', result)
        
        # Verify all functions were called
        mock_validate.assert_called_once()
        mock_download.assert_called_once()
        mock_synthesize.assert_called_once()
        mock_upload.assert_called_once()
        mock_update_metadata.assert_called_once()

    @patch.dict('os.environ', {'OUTPUT_BUCKET_NAME': 'test-output-bucket'})
    @patch('handler.boto3')
    def test_error_handling_s3_download_failure(self, mock_boto3):
        """Test error handling when S3 download fails"""
        # Mock S3 client to raise error
        mock_s3 = Mock()
        mock_boto3.client.return_value = mock_s3
        mock_s3.get_object.side_effect = Exception('S3 download failed')
        
        # Verify exception is raised
        with self.assertRaises(Exception) as context:
            handler.download_audio_from_s3('test-bucket', 'test-key')
        
        self.assertIn('S3 download failed', str(context.exception))

    @patch.dict('os.environ', {'OUTPUT_BUCKET_NAME': 'test-output-bucket'})
    @patch('handler.boto3')
    def test_error_handling_polly_failure(self, mock_boto3):
        """Test error handling when Polly synthesis fails"""
        # Mock Polly client to raise error
        mock_polly = Mock()
        mock_boto3.client.return_value = mock_polly
        mock_polly.synthesize_speech.side_effect = Exception('Polly synthesis failed')
        
        # Verify exception is raised
        with self.assertRaises(Exception) as context:
            handler.synthesize_audio_with_polly('Test text', 'Joanna')
        
        self.assertIn('Polly synthesis failed', str(context.exception))

    def test_output_key_generation(self):
        """Test generation of output S3 key with timestamp"""
        input_key = 'user-uploads/meditation.mp3'
        
        # Call key generation function
        output_key = handler.generate_output_key(input_key)
        
        # Verify output key format
        self.assertIn('processed/', output_key)
        self.assertIn('meditation', output_key)
        self.assertTrue(output_key.endswith('.mp3'))


if __name__ == '__main__':
    unittest.main()
