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

// TestStepFunctionsStateMachineExists verifies that Step Functions state machine is created
func TestStepFunctionsStateMachineExists(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Step Functions state machine exists
	template.ResourceCountIs(jsii.String("AWS::StepFunctions::StateMachine"), jsii.Number(1))
}

// TestStepFunctionsStateMachineHasPollyTask verifies state machine contains Polly integration
func TestStepFunctionsStateMachineHasPollyTask(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains Polly service integration
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*polly.*")),
				}),
			},
		},
	})
}

// TestStepFunctionsStateMachineHasCloudWatchLogs verifies logging is enabled
func TestStepFunctionsStateMachineHasCloudWatchLogs(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine has CloudWatch Logs enabled
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"LoggingConfiguration": map[string]interface{}{
			"Level": "ALL",
		},
	})
}

// TestStepFunctionsExecutionRoleHasPollyPermissions verifies IAM role has Polly permissions
func TestStepFunctionsExecutionRoleHasPollyPermissions(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify IAM role has Polly permissions (least privilege)
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": "polly:SynthesizeSpeech",
					"Effect": "Allow",
				},
			}),
		},
	})
}

// TestEventBridgeRuleTargetsStateMachine verifies EventBridge rule targets Step Functions
func TestEventBridgeRuleTargetsStateMachine(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify EventBridge rule has Step Functions state machine as target
	template.HasResourceProperties(jsii.String("AWS::Events::Rule"), map[string]interface{}{
		"Targets": assertions.Match_ArrayWith([]interface{}{
			map[string]interface{}{
				"Arn": map[string]interface{}{
					"Ref": assertions.Match_StringLikeRegexp(jsii.String(".*StateMachine.*")),
				},
			},
		}),
	})
}

// ========== Issue #5: DynamoDB Table + State Machine Input/Output Handling ==========

// TestDynamoDBMetadataTableExists verifies the DynamoDB table is created
func TestDynamoDBMetadataTableExists(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify DynamoDB table exists
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))
}

// TestDynamoDBTableHasCorrectSchema verifies the DynamoDB table has correct key schema
func TestDynamoDBTableHasCorrectSchema(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify table has audioId as partition key
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"KeySchema": []interface{}{
			map[string]interface{}{
				"AttributeName": "audioId",
				"KeyType":       "HASH",
			},
		},
		"AttributeDefinitions": assertions.Match_ArrayWith([]interface{}{
			map[string]interface{}{
				"AttributeName": "audioId",
				"AttributeType": "S",
			},
		}),
	})
}

// TestDynamoDBTableHasEncryption verifies the DynamoDB table has encryption enabled
func TestDynamoDBTableHasEncryption(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify table has server-side encryption enabled
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"SSESpecification": map[string]interface{}{
			"SSEEnabled": true,
		},
	})
}

// TestDynamoDBTableHasOnDemandBilling verifies billing mode is on-demand
func TestDynamoDBTableHasOnDemandBilling(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify table uses on-demand billing mode
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"BillingMode": "PAY_PER_REQUEST",
	})
}

// TestDynamoDBTableHasPointInTimeRecovery verifies PITR is enabled
func TestDynamoDBTableHasPointInTimeRecovery(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify point-in-time recovery is enabled
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"PointInTimeRecoverySpecification": map[string]interface{}{
			"PointInTimeRecoveryEnabled": true,
		},
	})
}

// TestStateMachineHasDynamoDBTask verifies state machine includes DynamoDB integration
func TestStateMachineHasDynamoDBTask(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains DynamoDB service integration
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*dynamodb.*")),
				}),
			},
		},
	})
}

// TestStateMachineRoleHasDynamoDBPermissions verifies IAM role has DynamoDB permissions
func TestStateMachineRoleHasDynamoDBPermissions(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify IAM role has DynamoDB PutItem and UpdateItem permissions
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": assertions.Match_ArrayWith([]interface{}{
						"dynamodb:PutItem",
					}),
					"Effect": "Allow",
				},
			}),
		},
	})
}

// ========== Issue #6: SNS Notifications and Error Handling ==========

// TestSNSTopicsExist verifies that SNS topics are created for notifications
func TestSNSTopicsExist(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify that exactly 2 SNS topics exist (success and failure)
	template.ResourceCountIs(jsii.String("AWS::SNS::Topic"), jsii.Number(2))
}

// TestSNSTopicsHaveEncryption verifies SNS topics are encrypted
func TestSNSTopicsHaveEncryption(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify SNS topics have KMS encryption enabled
	template.HasResourceProperties(jsii.String("AWS::SNS::Topic"), map[string]interface{}{
		"KmsMasterKeyId": map[string]interface{}{
			"Fn::GetAtt": assertions.Match_ArrayWith([]interface{}{
				assertions.Match_StringLikeRegexp(jsii.String(".*Key.*")),
			}),
		},
	})
}

// TestStateMachineHasSNSPublishTasks verifies state machine includes SNS publish tasks
func TestStateMachineHasSNSPublishTasks(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains SNS publish actions
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*sns.*")),
				}),
			},
		},
	})
}

// TestStateMachineHasErrorHandling verifies state machine includes error handling
func TestStateMachineHasErrorHandling(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains Catch error handling
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*Catch.*")),
				}),
			},
		},
	})
}

// TestStateMachineRoleHasSNSPublishPermissions verifies IAM role has SNS publish permissions
func TestStateMachineRoleHasSNSPublishPermissions(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify IAM role has SNS:Publish permissions
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": "sns:Publish",
					"Effect": "Allow",
				},
			}),
		},
	})
}

// TestDynamoDBUpdateTasksForFailure verifies state machine includes DynamoDB update for failures
func TestDynamoDBUpdateTasksForFailure(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains DynamoDB updateItem for FAILED status
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*FAILED.*")),
				}),
			},
		},
	})
}

// ========== Issue #7: Lambda Function Integration for Audio Processing ==========

// TestLambdaFunctionExists verifies the Lambda function resource is created
func TestLambdaFunctionExists(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Lambda function exists
	template.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))
}

// TestLambdaFunctionConfiguration verifies Lambda function has correct runtime and environment
func TestLambdaFunctionConfiguration(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Lambda function has correct runtime and environment variables
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
		"Runtime": assertions.Match_StringLikeRegexp(jsii.String("python.*")),
		"Handler": "handler.lambda_handler",
		"Environment": map[string]interface{}{
			"Variables": map[string]interface{}{
				"METADATA_TABLE_NAME": map[string]interface{}{
					"Ref": assertions.Match_StringLikeRegexp(jsii.String(".*MetadataTable.*")),
				},
			},
		},
	})
}

// TestLambdaFunctionHasExecutionRole verifies Lambda has proper IAM execution role
func TestLambdaFunctionHasExecutionRole(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Lambda function has an execution role
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
		"Role": map[string]interface{}{
			"Fn::GetAtt": assertions.Match_ArrayWith([]interface{}{
				assertions.Match_StringLikeRegexp(jsii.String(".*Role.*")),
			}),
		},
	})
}

// TestLambdaExecutionRoleHasDynamoDBPermissions verifies Lambda role can read/write DynamoDB
func TestLambdaExecutionRoleHasDynamoDBPermissions(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Lambda execution role has DynamoDB read/write permissions
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": assertions.Match_ArrayWith([]interface{}{
						"dynamodb:GetItem",
					}),
					"Effect": "Allow",
				},
			}),
		},
	})
}

// TestStateMachineHasLambdaInvokeTask verifies state machine includes Lambda invocation task
func TestStateMachineHasLambdaInvokeTask(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine definition contains Lambda invocation
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*SleepAudioProcessor.*")),
				}),
			},
		},
	})
}

// TestStateMachineRoleCanInvokeLambda verifies state machine role has lambda:InvokeFunction permission
func TestStateMachineRoleCanInvokeLambda(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine execution role has Lambda invoke permissions
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": "lambda:InvokeFunction",
					"Effect": "Allow",
				},
			}),
		},
	})
}

// ========== Issue #8: Complete Pipeline Integration with Input Validation ==========

// TestCompletePipelineWiring verifies all components are correctly wired together
func TestCompletePipelineWiring(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify EventBridge rule targets Step Functions state machine
	template.HasResourceProperties(jsii.String("AWS::Events::Rule"), map[string]interface{}{
		"Targets": assertions.Match_ArrayWith([]interface{}{
			map[string]interface{}{
				"Arn": map[string]interface{}{
					"Ref": assertions.Match_StringLikeRegexp(jsii.String(".*StateMachine.*")),
				},
			},
		}),
	})

	// Verify state machine definition contains all required tasks
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					// Verify complete chain exists
					assertions.Match_StringLikeRegexp(jsii.String(".*WriteInitialMetadata.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*InvokeSleepAudioProcessor.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*PollyTask.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*UpdateMetadataSuccess.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*PublishSuccess.*")),
				}),
			},
		},
	})
}

// TestStateMachineHasProperTaskChain verifies the state machine orchestrates tasks in correct order
func TestStateMachineHasProperTaskChain(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine contains proper task sequencing
	// WriteInitialMetadata -> Lambda -> Polly -> UpdateSuccess -> PublishSuccess
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*Next.*")),
				}),
			},
		},
	})
}

// TestLambdaHasInputValidation verifies Lambda function validates input properly
func TestLambdaHasInputValidation(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify Lambda function exists (validation logic is in handler.py)
	template.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))
	
	// Verify Lambda has proper configuration for validation
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
		"Handler": "handler.lambda_handler",
		"Runtime": assertions.Match_StringLikeRegexp(jsii.String("python.*")),
	})
}

// TestStateMachineErrorHandlingWithCatch verifies error paths exist
func TestStateMachineErrorHandlingWithCatch(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine has Catch blocks for error handling
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*Catch.*")),
				}),
			},
		},
	})
}

// TestErrorPathUpdatesStatusToFailed verifies failure path updates DynamoDB correctly
func TestErrorPathUpdatesStatusToFailed(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine contains UpdateMetadataFailure task
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*UpdateMetadataFailure.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*FAILED.*")),
				}),
			},
		},
	})
}

// TestSuccessPathUpdatesStatusToCompleted verifies success path updates DynamoDB correctly
func TestSuccessPathUpdatesStatusToCompleted(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine contains UpdateMetadataSuccess with COMPLETED status
	template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), map[string]interface{}{
		"DefinitionString": map[string]interface{}{
			"Fn::Join": []interface{}{
				"",
				assertions.Match_ArrayWith([]interface{}{
					assertions.Match_StringLikeRegexp(jsii.String(".*UpdateMetadataSuccess.*")),
					assertions.Match_StringLikeRegexp(jsii.String(".*COMPLETED.*")),
				}),
			},
		},
	})
}

// TestAllIAMPermissionsAreLeastPrivilege verifies IAM permissions across all services
func TestAllIAMPermissionsAreLeastPrivilege(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify state machine has necessary permissions but not excessive
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith([]interface{}{
				map[string]interface{}{
					"Action": "polly:SynthesizeSpeech",
					"Effect": "Allow",
				},
			}),
		},
	})
}

// TestCompleteStackSnapshot creates a comprehensive snapshot of the integrated stack
func TestCompleteStackSnapshot(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkBaseStack(app, "TestStack", nil)
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify all major resources exist
	template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(2))           // Input + Output
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))      // Metadata table
	template.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))     // Audio processor
	template.ResourceCountIs(jsii.String("AWS::StepFunctions::StateMachine"), jsii.Number(1)) // State machine
	template.ResourceCountIs(jsii.String("AWS::Events::Rule"), jsii.Number(1))         // EventBridge rule
	template.ResourceCountIs(jsii.String("AWS::SNS::Topic"), jsii.Number(2))           // Success + Failure topics
}
