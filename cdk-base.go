package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctionstasks"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkBaseStackProps struct {
	awscdk.StackProps
}

func NewCdkBaseStack(scope constructs.Construct, id string, props *CdkBaseStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

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
		PartitionKey: &awsdynamodb.Attribute{
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
	// Create Lambda function for audio processing, metadata enrichment, or validation
	audioProcessorFunction := awslambda.NewFunction(stack, jsii.String("SleepAudioProcessor"), &awslambda.FunctionProps{
		FunctionName: jsii.String("SleepAudioProcessor"),
		Runtime:      awslambda.Runtime_PYTHON_3_12(),
		Handler:      jsii.String("handler.lambda_handler"),
		Code:         awslambda.Code_FromAsset(jsii.String("lambda/audio-processor"), nil),
		Environment: &map[string]*string{
			"METADATA_TABLE_NAME": metadataTable.TableName(),
		},
		Timeout:     awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:  jsii.Number(256),
		Description: jsii.String("Processes audio files, enriches metadata, and performs validation"),
	})

	// Grant Lambda function read/write permissions to DynamoDB table
	metadataTable.GrantReadWriteData(audioProcessorFunction)

	// Grant Lambda function read access to input bucket (for future file validation)
	inputBucket.GrantRead(audioProcessorFunction, nil)

	// Task 1: Write initial metadata record to DynamoDB when pipeline starts
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

	// ========== Issue #6: SNS Topics for Notifications ==========
	// Create KMS key for SNS topic encryption
	snsEncryptionKey := awskms.NewKey(stack, jsii.String("SNSEncryptionKey"), &awskms.KeyProps{
		Description:       jsii.String("KMS key for encrypting SNS topics"),
		EnableKeyRotation: jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
	})

	// Create SNS topic for successful pipeline completion
	completionTopic := awssns.NewTopic(stack, jsii.String("SleepAudioPipelineCompleted"), &awssns.TopicProps{
		TopicName:   jsii.String("SleepAudioPipelineCompleted"),
		DisplayName: jsii.String("Sleep Audio Pipeline Completed Notifications"),
		MasterKey:   snsEncryptionKey,
	})

	// Create SNS topic for pipeline failures
	failureTopic := awssns.NewTopic(stack, jsii.String("SleepAudioPipelineFailed"), &awssns.TopicProps{
		TopicName:   jsii.String("SleepAudioPipelineFailed"),
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

	// This is a placeholder implementation - real audio generation logic comes later
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

	// Add error handling to Polly task - if it fails, update DDB and publish to failure topic
	pollyTask.AddCatch(updateMetadataFailureTask.Next(publishFailureTask), &awsstepfunctions.CatchProps{
		ResultPath: jsii.String("$.Error"),
	})

	// Add error handling to Lambda task - if it fails or validation fails, update DDB and publish to failure topic
	audioProcessorTask.AddCatch(updateMetadataFailureTask.Next(publishFailureTask), &awsstepfunctions.CatchProps{
		ResultPath: jsii.String("$.Error"),
	})

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
		DefinitionBody:   awsstepfunctions.DefinitionBody_FromChainable(definition),
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
