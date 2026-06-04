package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
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
	// Create CloudWatch Log Group for Step Functions state machine
	stateMachineLogGroup := awslogs.NewLogGroup(stack, jsii.String("StateMachineLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/vendedlogs/states/SleepAudioPipeline"),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,

	// Grant read permissions to placeholder Lambda
	// Create a minimal Polly task using CallAwsService for Amazon Polly integration
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

	// Define the state machine with the Polly task
	definition := pollyTask

	// Create the Step Functions state machine for audio processing pipeline
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
	rule.AddTarget(awseventstargets.NewLambdaFunction(placeholderLambda, &awseventstargets.LambdaFunctionProps{}))
	// Add Step Functions state machine as target for EventBridge rule
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
	return stack
}
	awscdk.NewCfnOutput(stack, jsii.String("StateMachineArn"), &awscdk.CfnOutputProps{
		Value:       stateMachine.StateMachineArn(),
		Description: jsii.String("ARN of the Step Functions state machine"),
	})

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
