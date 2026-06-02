package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"

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
	// This will be replaced with Step Functions in a future issue
	placeholderLambda := awslambda.NewFunction(stack, jsii.String("PlaceholderProcessorFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromInline(jsii.String("# Placeholder - will be replaced with Step Functions")),
		Description: jsii.String("Placeholder Lambda for EventBridge target - will be replaced with Step Functions state machine"),
		Timeout:     awscdk.Duration_Seconds(jsii.Number(30)),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"OUTPUT_BUCKET": outputBucket.BucketName(),
		},
	})

	// Grant read permissions to placeholder Lambda
	inputBucket.GrantRead(placeholderLambda, nil)
	outputBucket.GrantWrite(placeholderLambda, nil)

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

	// Output bucket names for reference
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
