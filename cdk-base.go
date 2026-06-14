package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctionstasks"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// CdkBaseStackProps defines the properties for the CDK base stack.
// It extends the standard CDK StackProps with an optional Environment field
// to support environment-specific resource naming and configuration.
type CdkBaseStackProps struct {
	awscdk.StackProps
	Environment *string
}

// NewCdkBaseStack creates a new CDK base stack for the Sleep Audio Pipeline.
// This stack contains all the infrastructure resources needed for the audio processing pipeline:
// - S3 buckets for input and output storage
// - DynamoDB table for metadata tracking
// - Lambda function for audio processing
// - Step Functions state machine for orchestration
// - EventBridge rule for event routing
// - SNS topics for notifications
// - CloudWatch alarms for monitoring
func NewCdkBaseStack(scope constructs.Construct, id string, props *CdkBaseStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Determine environment name for resource naming and tagging
	environment := "dev"
	if props != nil && props.Environment != nil {
		environment = *props.Environment
	}

	// Apply environment tags to all resources in the stack
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(environment), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Project"), jsii.String("SleepAudioPipeline"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("ManagedBy"), jsii.String("CDK"), nil)

	// Create Input S3 Bucket for raw audio file uploads
	inputBucket := awss3.NewBucket(stack, jsii.String("SleepAudioInputBucket"), &awss3.BucketProps{
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EventBridgeEnabled: jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
		EnforceSSL:        jsii.Bool(true),
	})

	// Create Output S3 Bucket for processed audio files
	outputBucket := awss3.NewBucket(stack, jsii.String("SleepAudioOutputBucket"), &awss3.BucketProps{
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
		EnforceSSL:        jsii.Bool(true),
	})

	// Create a placeholder Lambda function for EventBridge target
	// ========== Issue #5: DynamoDB Table for Metadata Storage ==========
	// Create DynamoDB table for storing audio pipeline metadata
	metadataTable := awsdynamodb.NewTable(stack, jsii.String("SleepAudioMetadataTable"), &awsdynamodb.TableProps{
		TableName: jsii.String("SleepAudioMetadataTable"),
		TableName: jsii.String("SleepAudioMetadataTable-" + environment),
			Name: jsii.String("audioId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:         awsdynamodb.BillingMode_PAY_PER_REQUEST,
		Encryption:          awsdynamodb.TableEncryption_AWS_MANAGED,
		PointInTimeRecovery: jsii.Bool(true),
		RemovalPolicy:       awscdk.RemovalPolicy_RETAIN,
	})

	stateMachineLogGroup := awslogs.NewLogGroup(stack, jsii.String("StateMachineLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/vendedlogs/states/SleepAudioPipeline"),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,

	})

	// ========== Issue #5: State Machine with DynamoDB Integration ==========
	// ========== Issue #7: Lambda Function for Audio Processing ==========
	// ========== Issue #10: Enhanced observability with X-Ray tracing ==========
	// ========== Issue #11: Core audio processing logic with Polly integration ==========
	audioProcessorFunction := awslambda.NewFunction(stack, jsii.String("SleepAudioProcessor"), &awslambda.FunctionProps{
		FunctionName: jsii.String("SleepAudioProcessor"),
		FunctionName: jsii.String("SleepAudioProcessor-" + environment),
		Handler:      jsii.String("handler.lambda_handler"),
		Code:         awslambda.Code_FromAsset(jsii.String("lambda/audio-processor"), nil),
		Environment: &map[string]*string{
			"METADATA_TABLE_NAME": metadataTable.TableName(),
			"OUTPUT_BUCKET_NAME":  outputBucket.BucketName(),
		},
		Timeout:     awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:  jsii.Number(256),
		Description: jsii.String("Processes audio files, enriches metadata, and performs validation"),
		Tracing:     awslambda.Tracing_ACTIVE, // Enable X-Ray tracing
	})

	// Grant Lambda function read/write permissions to DynamoDB table
	metadataTable.GrantReadWriteData(audioProcessorFunction)

	// Grant Lambda function read access to input bucket (for future file validation)
	inputBucket.GrantRead(audioProcessorFunction, nil)

	// ========== Issue #11: Grant Lambda permissions for audio processing and output handling ==========
	// Grant Lambda write access to output bucket for processed audio files
	outputBucket.GrantWrite(audioProcessorFunction, nil)

	// Grant Lambda permission to use Amazon Polly for text-to-speech synthesis
	audioProcessorFunction.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("polly:SynthesizeSpeech"),
		Resources: jsii.Strings("*"), // Polly doesn't support resource-level permissions
	}))

	// Task 1: Write initial metadata record to DynamoDB when pipeline starts
	// ========== Issue #10: Add retry policy for DynamoDB operations ==========
	retryPolicyDynamoDB := &[]*awsstepfunctions.RetryProps{
		{
			Errors:       jsii.Strings("States.TaskFailed", "DynamoDB.ProvisionedThroughputExceededException", "DynamoDB.RequestLimitExceeded"),
			Interval:     awscdk.Duration_Seconds(jsii.Number(2)),
			MaxAttempts:  jsii.Number(3),
			BackoffRate:  jsii.Number(2.0), // Exponential backoff
		},
	}

	// Task 1: Write initial metadata with retry policy
	writeInitialMetadataTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("WriteInitialMetadata"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service: jsii.String("dynamodb"),
		Action:  jsii.String("putItem"),
		Parameters: &map[string]interface{}{
			"TableName": metadataTable.TableName(),
			"Item": map[string]interface{}{
				"audioId": map[string]interface{}{
					"S.$": "$.detail.object.key",
				},
				"status": map[string]interface{}{
					"S": "PROCESSING",
				},
				"inputBucket": map[string]interface{}{
					"S.$": "$.detail.bucket.name",
				},
				"inputKey": map[string]interface{}{
					"S.$": "$.detail.object.key",
				},
				"createdAt": map[string]interface{}{
					"S.$": "$$.State.EnteredTime",
				},
			},
		},
		IamResources: &[]*string{metadataTable.TableArn()},
		ResultPath:   jsii.String("$.dynamoDbResult"),
	})
	writeInitialMetadataTask.AddRetry((*retryPolicyDynamoDB)[0])

	// Task 2: Update metadata after successful processing
	updateMetadataSuccessTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("UpdateMetadataSuccess"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service: jsii.String("dynamodb"),
		Action:  jsii.String("updateItem"),
		Parameters: &map[string]interface{}{
			"TableName": metadataTable.TableName(),
			"Key": map[string]interface{}{
				"audioId": map[string]interface{}{
					"S.$": "$.detail.object.key",
				},
			},
			"UpdateExpression": "SET #status = :completed, updatedAt = :timestamp",
			"ExpressionAttributeNames": map[string]interface{}{
				"#status": "status",
			},
			"ExpressionAttributeValues": map[string]interface{}{
				":completed": map[string]interface{}{
					"S": "COMPLETED",
				},
				":timestamp": map[string]interface{}{
					"S.$": "$$.State.EnteredTime",
				},
			},
		},
		IamResources: &[]*string{metadataTable.TableArn()},
		ResultPath:   jsii.String("$.updateResult"),
	})
	updateMetadataSuccessTask.AddRetry((*retryPolicyDynamoDB)[0])

	// ========== Issue #6: SNS Topics for Notifications ==========
	// Create KMS key for SNS topic encryption
	snsEncryptionKey := awskms.NewKey(stack, jsii.String("SNSEncryptionKey"), &awskms.KeyProps{
		Description:       jsii.String("KMS key for encrypting SNS topics"),
		EnableKeyRotation: jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
	})

	// Create SNS topic for successful pipeline completion
	completionTopic := awssns.NewTopic(stack, jsii.String("SleepAudioPipelineCompleted"), &awssns.TopicProps{
		TopicName:   jsii.String("SleepAudioPipelineCompleted-" + environment),
		DisplayName: jsii.String("Sleep Audio Pipeline Completed Notifications"),
		MasterKey:   snsEncryptionKey,
	})

	// Create SNS topic for pipeline failures
	failureTopic := awssns.NewTopic(stack, jsii.String("SleepAudioPipelineFailed"), &awssns.TopicProps{
		TopicName:   jsii.String("SleepAudioPipelineFailed-" + environment),
		DisplayName: jsii.String("Sleep Audio Pipeline Failed Notifications"),
		MasterKey:   snsEncryptionKey,
	})

	// Task 3: Update metadata when pipeline fails
	updateMetadataFailureTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("UpdateMetadataFailure"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service: jsii.String("dynamodb"),
		Action:  jsii.String("updateItem"),
		Parameters: &map[string]interface{}{
			"TableName": metadataTable.TableName(),
			"Key": map[string]interface{}{
				"audioId": map[string]interface{}{
					"S.$": "$.detail.object.key",
				},
			},
			"UpdateExpression": "SET #status = :failed, updatedAt = :timestamp, errorMessage = :error",
			"ExpressionAttributeNames": map[string]interface{}{
				"#status": "status",
			},
			"ExpressionAttributeValues": map[string]interface{}{
				":failed": map[string]interface{}{
					"S": "FAILED",
				},
				":timestamp": map[string]interface{}{
					"S.$": "$$.State.EnteredTime",
				},
				":error": map[string]interface{}{
					"S.$": "$.Error",
				},
			},
		},
		IamResources: &[]*string{metadataTable.TableArn()},
		ResultPath:   jsii.String("$.updateResult"),
	})
	updateMetadataFailureTask.AddRetry((*retryPolicyDynamoDB)[0])

	// Task 4: Publish success notification to SNS
	publishSuccessTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("PublishSuccess"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service: jsii.String("sns"),
		Action:  jsii.String("publish"),
		Parameters: &map[string]interface{}{
			"TopicArn": completionTopic.TopicArn(),
			"Message.$": "States.Format('Pipeline completed successfully for audioId: {}', $.detail.object.key)",
			"Subject":  jsii.String("Sleep Audio Pipeline Completed"),
		},
		IamResources: &[]*string{completionTopic.TopicArn()},
		ResultPath:   jsii.String("$.snsResult"),
	})

	// Task 5: Publish failure notification to SNS
	publishFailureTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("PublishFailure"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service: jsii.String("sns"),
		Action:  jsii.String("publish"),
		Parameters: &map[string]interface{}{
			"TopicArn": failureTopic.TopicArn(),
			"Message.$": "States.Format('Pipeline failed for audioId: {} with error: {}', $.detail.object.key, $.Error)",
			"Subject":  jsii.String("Sleep Audio Pipeline Failed"),
		},
		IamResources: &[]*string{failureTopic.TopicArn()},
		ResultPath:   jsii.String("$.snsResult"),
	})

	// Grant read permissions to placeholder Lambda
	// ========== Issue #10: Lambda task with retry policy and enhanced error handling ==========
	// Task 6: Invoke Lambda function for audio processing
	audioProcessorTask := awsstepfunctionstasks.NewLambdaInvoke(stack, jsii.String("InvokeSleepAudioProcessor"), &awsstepfunctionstasks.LambdaInvokeProps{
		LambdaFunction: audioProcessorFunction,
		ResultPath:     jsii.String("$.processorResult"),
		Payload: awsstepfunctions.TaskInput_FromObject(&map[string]interface{}{
			"detail.$": "$.detail",
			"bucket.$": "$.detail.bucket.name",
			"audioId.$": "$.detail.object.key",
		}),
	})
	
	// Add retry policy for Lambda task (exponential backoff)
	audioProcessorTask.AddRetry(&awsstepfunctions.RetryProps{
		Errors:      jsii.Strings("Lambda.ServiceException", "Lambda.TooManyRequestsException", "States.TaskFailed"),
		Interval:    awscdk.Duration_Seconds(jsii.Number(2)),
		MaxAttempts: jsii.Number(3),
		BackoffRate: jsii.Number(2.0),
	})
	
	// Add enhanced error handling with specific error types
	audioProcessorTask.AddCatch(updateMetadataFailureTask.Next(publishFailureTask), &awsstepfunctions.CatchProps{
		Errors:     jsii.Strings("States.ALL"),
		ResultPath: jsii.String("$.Error"),
	})

	// ========== Issue #10: Polly task with retry policy and enhanced error handling ==========
	// Polly task - placeholder implementation for audio generation
	pollyTask := awsstepfunctionstasks.NewCallAwsService(stack, jsii.String("PollyTask"), &awsstepfunctionstasks.CallAwsServiceProps{
		Service:      jsii.String("polly"),
		Action:       jsii.String("synthesizeSpeech"),
		IamResources: &[]*string{jsii.String("*")},
		Parameters: &map[string]interface{}{
			"Text":         jsii.String("This is a placeholder for sleep audio generation"),
			"VoiceId":      jsii.String("Joanna"),
			"OutputFormat": jsii.String("mp3"),
			"Engine":       jsii.String("neural"),
		},
		ResultPath: jsii.String("$.pollyResult"),
	})
	
	// Add retry policy for Polly task
	pollyTask.AddRetry(&awsstepfunctions.RetryProps{
		Errors:      jsii.Strings("Polly.EngineNotSupportedException", "Polly.ServiceFailureException", "States.TaskFailed"),
		Interval:    awscdk.Duration_Seconds(jsii.Number(2)),
		MaxAttempts: jsii.Number(3),
		BackoffRate: jsii.Number(2.0),
	})
	
	// Add enhanced error handling with specific error types
	pollyTask.AddCatch(updateMetadataFailureTask.Next(publishFailureTask), &awsstepfunctions.CatchProps{
		Errors:     jsii.Strings("States.ALL"),
		ResultPath: jsii.String("$.Error"),
	})

	// ========== Issue #10: CloudWatch Alarms for Observability ==========
	// Create CloudWatch Alarms for critical failure monitoring
	// Alarm 1: State Machine Execution Failures

	// Define the state machine workflow with error handling:
	// Complete end-to-end flow:
	// 1. Write initial metadata (status=PROCESSING)
	// 2. Invoke Lambda for audio processing and validation (with error handling)
	// 3. Call Polly for text-to-speech synthesis (with error handling)
	// 4. Update metadata to COMPLETED
	// 5. Publish success notification
	// Error paths: Lambda/Polly failures -> Update status to FAILED -> Publish failure notification
	definition := writeInitialMetadataTask.
		Next(audioProcessorTask).
		Next(pollyTask).
		Next(updateMetadataSuccessTask).
		Next(publishSuccessTask)

	// Create the Step Functions state machine for audio processing pipeline
	// This state machine orchestrates the complete end-to-end pipeline
	stateMachine := awsstepfunctions.NewStateMachine(stack, jsii.String("SleepAudioPipelineStateMachine"), &awsstepfunctions.StateMachineProps{
		StateMachineName: jsii.String("SleepAudioPipelineStateMachine"),
		StateMachineName: jsii.String("SleepAudioPipelineStateMachine-" + environment),
		StateMachineType: awsstepfunctions.StateMachineType_STANDARD,
		Logs: &awsstepfunctions.LogOptions{
			Destination: stateMachineLogGroup,
			Level:       awsstepfunctions.LogLevel_ALL,
		},
		TracingEnabled: jsii.Bool(true),
		Timeout:        awscdk.Duration_Minutes(jsii.Number(5)),
	})

	// Grant the state machine read access to input bucket and write access to output bucket
	inputBucket.GrantRead(stateMachine, nil)
	outputBucket.GrantWrite(stateMachine, nil)
	// Create EventBridge Rule to trigger on S3 Object Created events

	// Grant the state machine permissions to read/write DynamoDB table
	metadataTable.GrantReadWriteData(stateMachine)

	// Create CloudWatch Alarm for state machine execution failures
	awscloudwatch.NewAlarm(stack, jsii.String("StateMachineExecutionFailedAlarm"), &awscloudwatch.AlarmProps{
		AlarmName:          jsii.String("SleepAudioPipeline-ExecutionsFailed"),
		AlarmName:          jsii.String("SleepAudioPipeline-ExecutionsFailed-" + environment),
		Metric: awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
			Namespace:          jsii.String("AWS/States"),
			MetricName:         jsii.String("ExecutionsFailed"),
			DimensionsMap: &map[string]*string{
				"StateMachineArn": stateMachine.StateMachineArn(),
			},
			Statistic: jsii.String("Sum"),
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		}),
		Threshold:       jsii.Number(1),
		EvaluationPeriods: jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData: awscloudwatch.TreatMissingData_NOT_BREACHING,
	})

	// Create CloudWatch Alarm for Lambda function errors
	awscloudwatch.NewAlarm(stack, jsii.String("LambdaErrorAlarm"), &awscloudwatch.AlarmProps{
		AlarmName:          jsii.String("SleepAudioProcessor-Errors"),
		AlarmName:          jsii.String("SleepAudioProcessor-Errors-" + environment),
		Metric: awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
			Namespace:          jsii.String("AWS/Lambda"),
			MetricName:         jsii.String("Errors"),
			DimensionsMap: &map[string]*string{
				"FunctionName": audioProcessorFunction.FunctionName(),
			},
			Statistic: jsii.String("Sum"),
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		}),
		Threshold:       jsii.Number(5),
		EvaluationPeriods: jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData: awscloudwatch.TreatMissingData_NOT_BREACHING,
	})

	// ========== EventBridge Rule for Pipeline Triggering ==========
	// Create EventBridge Rule to trigger on S3 Object Created events

	rule := awsevents.NewRule(stack, jsii.String("SleepAudioProcessingRule"), &awsevents.RuleProps{
		Description: jsii.String("Triggers audio processing pipeline when new files are uploaded to the input bucket"),
		EventPattern: &awsevents.EventPattern{
			Source:     jsii.Strings("aws.s3"),
			DetailType: jsii.Strings("Object Created"),
			Detail: &map[string]interface{}{
				"bucket": &map[string]interface{}{
					"name": []interface{}{inputBucket.BucketName()},
				},
			},
		},
		Enabled: jsii.Bool(true),
	})

	// Add placeholder Lambda as target
	// Pass S3 event data (bucket, key, etc.) as input to the state machine
	rule.AddTarget(awseventstargets.NewSfnStateMachine(stateMachine, &awseventstargets.SfnStateMachineProps{
		Input: awsevents.RuleTargetInput_FromEventPath(jsii.String("$")),
	}))
	awscdk.NewCfnOutput(stack, jsii.String("InputBucketName"), &awscdk.CfnOutputProps{
		Value:       inputBucket.BucketName(),
		Description: jsii.String("Name of the S3 bucket for raw audio file uploads"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("OutputBucketName"), &awscdk.CfnOutputProps{
		Value:       outputBucket.BucketName(),
		Description: jsii.String("Name of the S3 bucket for processed audio files"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("StateMachineArn"), &awscdk.CfnOutputProps{
		Value:       stateMachine.StateMachineArn(),
		Description: jsii.String("ARN of the Step Functions state machine"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("MetadataTableName"), &awscdk.CfnOutputProps{
		Value:       metadataTable.TableName(),
		Description: jsii.String("Name of the DynamoDB table for audio metadata"),
	})
	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkBaseStack(app, "CdkBaseStack", &CdkBaseStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
//
// By default, this returns nil, making the stack environment-agnostic.
// For production deployments, uncomment one of the options below.
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
