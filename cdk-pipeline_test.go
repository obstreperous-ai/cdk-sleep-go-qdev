package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

// Helper functions for pipeline tests

// createTestPipelineStack creates a CDK pipeline stack for testing
func createTestPipelineStack() awscdk.Stack {
	app := awscdk.NewApp(nil)
	return NewPipelineStack(app, "TestPipelineStack", &PipelineStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("us-east-1"),
			},
		},
	})
}

// TestPipelineStackCreation verifies that the pipeline stack can be created
func TestPipelineStackCreation(t *testing.T) {
	defer jsii.Close()

	// WHEN
	stack := createTestPipelineStack()

	// THEN
	if stack == nil {
		t.Fatal("Expected pipeline stack to be created, got nil")
	}
}

// TestPipelineExists verifies that a CodePipeline is created
func TestPipelineExists(t *testing.T) {
	defer jsii.Close()

	// WHEN
	stack := createTestPipelineStack()
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify CodePipeline resource exists
	template.ResourceCountIs(jsii.String("AWS::CodePipeline::Pipeline"), jsii.Number(1))
}

// TestPipelineHasSelfMutation verifies pipeline can update itself
func TestPipelineHasSelfMutation(t *testing.T) {
	defer jsii.Close()

	// WHEN
	stack := createTestPipelineStack()
	template := assertions.Template_FromStack(stack, nil)

	// THEN
	// Verify pipeline has self-mutation capability (UpdatePipeline stage)
	template.HasResourceProperties(jsii.String("AWS::CodePipeline::Pipeline"), map[string]interface{}{
		"Stages": assertions.Match_ArrayWith([]interface{}{
			map[string]interface{}{
				"Name": "UpdatePipeline",
			},
		}),
	})
}

// TestPipelineStageCreation verifies pipeline stages can be created
func TestPipelineStageCreation(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stage := NewPipelineStage(app, "TestStage", &PipelineStageProp{
		Environment: "dev",
	})

	// THEN
	if stage == nil {
		t.Fatal("Expected pipeline stage to be created, got nil")
	}
}
