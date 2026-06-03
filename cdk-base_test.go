package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

// TestCdkBaseStackCreation verifies that the stack can be created without errors
func TestCdkBaseStackCreation(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)

	// THEN
	if stack == nil {
		t.Fatal("Expected stack to be created, got nil")
	}
	// Stack creation succeeds - this is our baseline test
	// Future tests will verify specific resources using assertions.Template_FromStack
}

// TestSleepAudioInputBucket verifies the Input S3 bucket is created with correct properties
func TestSleepAudioInputBucket(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify S3 bucket exists with correct properties
	template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(2)) // Input + Output

	// Verify Input Bucket has encryption enabled
	template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
		"BucketEncryption": map[string]interface{}{
			"ServerSideEncryptionConfiguration": []interface{}{
				map[string]interface{}{
					"ServerSideEncryptionByDefault": map[string]interface{}{
						"SSEAlgorithm": "AES256",
					},
				},
			},
		},
		"VersioningConfiguration": map[string]interface{}{
			"Status": "Enabled",
		},
		"PublicAccessBlockConfiguration": map[string]interface{}{
			"BlockPublicAcls":       true,
			"BlockPublicPolicy":     true,
			"IgnorePublicAcls":      true,
			"RestrictPublicBuckets": true,
		},
		"NotificationConfiguration": map[string]interface{}{
			"EventBridgeConfiguration": map[string]interface{}{
				"EventBridgeEnabled": true,
			},
		},
	})
}

// TestSleepAudioOutputBucket verifies the Output S3 bucket is created with correct properties
func TestSleepAudioOutputBucket(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Output Bucket has encryption and versioning
	template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
		"BucketEncryption": map[string]interface{}{
			"ServerSideEncryptionConfiguration": []interface{}{
				map[string]interface{}{
					"ServerSideEncryptionByDefault": map[string]interface{}{
						"SSEAlgorithm": "AES256",
					},
				},
			},
		},
		"VersioningConfiguration": map[string]interface{}{
			"Status": "Enabled",
		},
		"PublicAccessBlockConfiguration": map[string]interface{}{
			"BlockPublicAcls":       true,
			"BlockPublicPolicy":     true,
			"IgnorePublicAcls":      true,
			"RestrictPublicBuckets": true,
		},
	})
}

// TestEventBridgeRule verifies EventBridge rule is created for S3 events
func TestEventBridgeRule(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify EventBridge rule exists
	template.ResourceCountIs(jsii.String("AWS::Events::Rule"), jsii.Number(1))

	// Verify rule has correct event pattern for S3 Object Created events
	template.HasResourceProperties(jsii.String("AWS::Events::Rule"), map[string]interface{}{
		"EventPattern": map[string]interface{}{
			"source":      []interface{}{"aws.s3"},
			"detail-type": []interface{}{"Object Created"},
		},
		"State": "ENABLED",
	})
}

// TestStackSnapshot creates a snapshot test for the entire synthesized stack
func TestStackSnapshot(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN - Basic sanity check that template is not empty
	templateJSON := template.ToJSON()
	if templateJSON == nil {
		t.Fatal("Expected template to be generated, got nil")
	}
}
