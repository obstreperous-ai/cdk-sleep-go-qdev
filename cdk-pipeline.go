package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/pipelines"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type PipelineStackProps struct {
	awscdk.StackProps
}

// PipelineStage represents a deployment stage (dev, stage, prod)
type PipelineStage struct {
	awscdk.Stage
}

type PipelineStageProp struct {
	awscdk.StageProps
	Environment string
}

// NewPipelineStage creates a new pipeline stage containing the application stack
func NewPipelineStage(scope constructs.Construct, id string, props *PipelineStageProp) awscdk.Stage {
	var sprops awscdk.StageProps
	if props != nil {
		sprops = props.StageProps
	}

	stage := awscdk.NewStage(scope, &id, &sprops)

	// Create the application stack in this stage
	environment := "dev"
	if props != nil && props.Environment != "" {
		environment = props.Environment
	}

	NewCdkBaseStack(stage, "SleepAudioPipelineStack", &CdkBaseStackProps{
		StackProps: awscdk.StackProps{
			Env: props.Env,
		},
		Environment: jsii.String(environment),
	})

	return stage
}

// NewPipelineStack creates a CDK Pipeline for automated deployment
func NewPipelineStack(scope constructs.Construct, id string, props *PipelineStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create the pipeline
	// Note: This is a skeleton implementation. In production, you would configure:
	// - Source: GitHub, CodeCommit, etc.
	// - Build: CodeBuild project
	// - Self-mutation: Pipeline updates itself
	pipeline := pipelines.NewCodePipeline(stack, jsii.String("Pipeline"), &pipelines.CodePipelineProps{
		PipelineName: jsii.String("SleepAudioPipeline"),
		
		// Synth step - builds the CDK application
		Synth: pipelines.NewShellStep(jsii.String("Synth"), &pipelines.ShellStepProps{
			// Input: GitHub source (placeholder - configure with actual source)
			// For now, this is a skeleton that demonstrates the structure
			Commands: jsii.Strings(
				"npm install -g aws-cdk",
				"go mod download",
				"go test -v",
				"cdk synth",
			),
		}),

		// Enable Docker support for Lambda functions
		DockerEnabledForSynth: jsii.Bool(true),
		
		// Enable self-mutation so the pipeline can update itself
		SelfMutation: jsii.Bool(true),
	})

	// Add deployment stages
	
	// Dev stage - deploys automatically
	devStage := NewPipelineStage(stack, "Dev", &PipelineStageProp{
		StageProps: awscdk.StageProps{
			Env: &awscdk.Environment{
				Account: jsii.String("111111111111"), // Replace with actual dev account
				Region:  jsii.String("us-east-1"),
			},
		},
		Environment: "dev",
	})
	pipeline.AddStage(devStage, nil)

	// Stage/Staging environment - deploys automatically after dev
	stageStage := NewPipelineStage(stack, "Stage", &PipelineStageProp{
		StageProps: awscdk.StageProps{
			Env: &awscdk.Environment{
				Account: jsii.String("111111111111"), // Replace with actual staging account
				Region:  jsii.String("us-east-1"),
			},
		},
		Environment: "stage",
	})
	pipeline.AddStage(stageStage, nil)

	// Production stage - requires manual approval
	prodStage := NewPipelineStage(stack, "Prod", &PipelineStageProp{
		StageProps: awscdk.StageProps{
			Env: &awscdk.Environment{
				Account: jsii.String("222222222222"), // Replace with actual prod account
				Region:  jsii.String("us-east-1"),
			},
		},
		Environment: "prod",
	})
	pipeline.AddStage(prodStage, &pipelines.AddStageOpts{
		Pre: &[]pipelines.Step{
			pipelines.NewManualApprovalStep(jsii.String("PromoteToProd"), &pipelines.ManualApprovalStepProps{
				Comment: jsii.String("Approve deployment to production"),
			}),
		},
	})

	// Add pipeline outputs
	awscdk.NewCfnOutput(stack, jsii.String("PipelineName"), &awscdk.CfnOutputProps{
		Value:       pipeline.Pipeline().PipelineName(),
		Description: jsii.String("Name of the CI/CD pipeline"),
	})

	// Tag the pipeline stack
	awscdk.Tags_Of(stack).Add(jsii.String("Project"), jsii.String("SleepAudioPipeline"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("ManagedBy"), jsii.String("CDK"), nil)

	return stack
}
