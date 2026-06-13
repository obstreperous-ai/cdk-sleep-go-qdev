# Meta-Prompts and Reusable Patterns

This document extracts reusable patterns, templates, and meta-prompts from the CDK Sleep Audio Pipeline project. These patterns can be applied to future agentic Test-Driven Development (TDD) Infrastructure-as-Code (IaC) projects.

## Purpose

This project demonstrates **pure issue-driven, TDD-first development** using AI agents. The patterns documented here serve as:

- **Meta-prompts** for AI agents working on similar projects
- **Templates** for structuring TDD IaC projects
- **Guidelines** for maintaining architectural consistency
- **Best practices** for serverless, event-driven AWS architectures
- **Reusable instructions** for future CDK projects

## Core Meta-Prompt: TDD IaC Agent Persona

```
You are a Senior AWS CDK [Language] TDD Specialist. You build infrastructure using 
strict Test-Driven Development principles. Every change follows the Red-Green-Refactor cycle.

CRITICAL RULES:
1. ARCHITECTURE.md is the single source of truth for all design decisions
2. Never write production code without a failing test first
3. Write minimal code to make tests pass
4. Keep ARCHITECTURE.md and diagrams in sync with every change
5. Use conventional commits for all changes
6. All tests must pass and CDK synth must succeed before committing

Your workflow:
1. Read ARCHITECTURE.md to understand the system
2. Write failing test (RED)
3. Verify test fails for the right reason
4. Write minimal code to pass test (GREEN)
5. Refactor while keeping tests green (REFACTOR)
6. Update ARCHITECTURE.md and diagrams
7. Verify: go test -v ./... && cdk synth
8. Commit with conventional format
```

## Pattern 1: Issue-Driven Development Template

### Issue Template

```markdown
## Issue #[N]: [Feature Name]

**Goal**: [Clear, concise goal statement]

**Context**: [Why this feature is needed, how it fits in the architecture]

**Acceptance Criteria**:
- [ ] [Specific, testable criterion 1]
- [ ] [Specific, testable criterion 2]
- [ ] [Specific, testable criterion 3]
- [ ] Architecture documentation updated
- [ ] All tests pass
- [ ] CDK synth succeeds

**Test Scenarios**:
1. **Scenario**: [Description]
   - Given: [Initial state]
   - When: [Action]
   - Then: [Expected result]

2. **Scenario**: [Edge case]
   - Given: [Initial state]
   - When: [Action]
   - Then: [Expected result]

**Architecture Impact**:
- [ ] New components: [List]
- [ ] Modified components: [List]
- [ ] Mermaid diagram updates: [Description]
- [ ] ADR needed: [Yes/No]

**Implementation Order**:
1. Write failing tests
2. Implement [Component A]
3. Implement [Component B]
4. Integration
5. Update documentation

**Success Criteria**:
- [ ] All tests pass (go test -v ./...)
- [ ] CDK synth succeeds
- [ ] ARCHITECTURE.md updated
- [ ] Mermaid diagram accurate
- [ ] Conventional commits used
```

## Pattern 2: TDD Cycle Template

### Red Phase Template

```go
// Test filename: [component]_test.go

func Test[FeatureName]_[Scenario](t *testing.T) {
    defer jsii.Close()

    // GIVEN - Setup test environment
    app := awscdk.NewApp(nil)
    
    // WHEN - Execute the code under test
    stack := New[Stack](app, "TestStack", nil)

    // THEN - Verify the expected behavior
    template := assertions.Template_FromStack(stack, nil)
    
    // Assert expected resource exists with properties
    template.HasResourceProperties(jsii.String("AWS::[Service]::[Resource]"), 
        map[string]interface{}{
            "Property": expectedValue,
        })
}
```

### Commit Message Template (Red Phase)

```
test: add failing test for [feature description]

Add test to verify [specific behavior being tested].

Test currently fails because [reason - feature not implemented].

Related to Issue #[N]
```

### Green Phase Template

Implementation approach:
1. Write **minimal** code to pass the test
2. Don't add features not covered by tests
3. Resist the urge to over-engineer
4. Get to green as quickly as possible

### Commit Message Template (Green Phase)

```
feat: implement [feature] to pass test

Implement [specific functionality] to satisfy test requirements.

- [Key implementation detail 1]
- [Key implementation detail 2]

All tests now pass.

Closes #[N]
```

### Refactor Phase Template

Refactoring checklist:
- [ ] Extract duplicated code
- [ ] Improve naming
- [ ] Simplify complex logic
- [ ] Add comments for clarity
- [ ] All tests still pass

```
refactor: improve [component] implementation

Refactor [component] to [improvement description]:
- [Change 1]
- [Change 2]

All tests remain green. No behavioral changes.

Related to Issue #[N]
```

## Pattern 3: Architecture Synchronization Workflow

### Pre-Implementation Checklist

```markdown
Before starting implementation:
- [ ] Read ARCHITECTURE.md completely
- [ ] Understand component interactions from Mermaid diagram
- [ ] Review relevant ADRs (Architectural Decision Records)
- [ ] Identify which components will be affected
- [ ] Plan test scenarios based on architecture
- [ ] Verify understanding with current implementation
```

### Post-Implementation Checklist

```markdown
After completing implementation:
- [ ] Update ARCHITECTURE.md with new/changed components
- [ ] Update Mermaid diagram to reflect changes
- [ ] Add ADR if architectural decision was made
- [ ] Update component descriptions to match implementation
- [ ] Update data flow diagrams if flows changed
- [ ] Verify all cross-references are accurate
- [ ] Update version number in ARCHITECTURE.md
- [ ] Document changes in changelog section
```

### Mermaid Diagram Update Template

When updating Mermaid diagrams:
1. Keep consistent styling (colors, shapes)
2. Maintain clear component groupings (subgraphs)
3. Show data flow with labeled arrows
4. Include retry/error paths with dotted lines
5. Add annotations for important behaviors
6. Keep diagram readable (not too crowded)

## Pattern 4: CDK Testing Patterns

### S3 Bucket Test Pattern

```go
func TestS3BucketEncryption(t *testing.T) {
    defer jsii.Close()
    app := awscdk.NewApp(nil)
    stack := NewMyStack(app, "TestStack", nil)
    template := assertions.Template_FromStack(stack, nil)
    
    template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), 
        map[string]interface{}{
            "BucketEncryption": map[string]interface{}{
                "ServerSideEncryptionConfiguration": []interface{}{
                    map[string]interface{}{
                        "ServerSideEncryptionByDefault": map[string]interface{}{
                            "SSEAlgorithm": "aws:kms",
                        },
                    },
                },
            },
        })
}
```

### Lambda Function Test Pattern

```go
func TestLambdaFunctionConfiguration(t *testing.T) {
    defer jsii.Close()
    app := awscdk.NewApp(nil)
    stack := NewMyStack(app, "TestStack", nil)
    template := assertions.Template_FromStack(stack, nil)
    
    template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), 
        map[string]interface{}{
            "Runtime": "python3.12",
            "MemorySize": 256,
            "Timeout": 30,
            "TracingConfig": map[string]interface{}{
                "Mode": "Active",
            },
        })
}
```

### State Machine Test Pattern

```go
func TestStateMachineDefinition(t *testing.T) {
    defer jsii.Close()
    app := awscdk.NewApp(nil)
    stack := NewMyStack(app, "TestStack", nil)
    template := assertions.Template_FromStack(stack, nil)
    
    // Verify state machine exists
    template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), 
        map[string]interface{}{
            "TracingConfiguration": map[string]interface{}{
                "Enabled": true,
            },
        })
    
    // Extract and verify definition
    template.HasResourceProperties(jsii.String("AWS::StepFunctions::StateMachine"), 
        map[string]interface{}{
            "DefinitionString": assertions.Match_StringLikeRegexp(".*SynthesizeSpeech.*"),
        })
}
```

## Pattern 5: Multi-Environment Configuration Pattern

### cdk.json Context Structure

```json
{
  "context": {
    "environments": {
      "dev": {
        "account": "111111111111",
        "region": "us-east-1",
        "logRetentionDays": 7,
        "enableXRay": true,
        "xraySamplingRate": 1.0
      },
      "prod": {
        "account": "222222222222",
        "region": "us-east-1",
        "logRetentionDays": 90,
        "enableXRay": true,
        "xraySamplingRate": 0.1
      }
    }
  }
}
```

### Environment-Aware Resource Naming

```go
func getResourceName(baseName string, environment string) string {
    return fmt.Sprintf("%s-%s", baseName, environment)
}

// Usage
bucketName := getResourceName("sleep-audio-input", environment)
```

## Pattern 6: Security Best Practices Checklist

```markdown
For every AWS resource:
- [ ] Encryption at rest enabled (KMS where applicable)
- [ ] Encryption in transit enforced (TLS 1.2+)
- [ ] IAM policies follow least privilege
- [ ] No hardcoded credentials or secrets
- [ ] Public access explicitly blocked (S3)
- [ ] Resource policies restrict access
- [ ] CloudTrail logging enabled for auditing
- [ ] Secrets stored in Secrets Manager
- [ ] Input validation implemented
- [ ] Error messages don't leak sensitive data
```

## Pattern 7: Observability Stack Pattern

Essential observability for every component:
- **Logging**: CloudWatch Logs with retention policies
- **Metrics**: CloudWatch metrics for key operations
- **Alarms**: Alarms on error rates and key metrics
- **Tracing**: X-Ray for distributed tracing
- **Dashboards**: CloudWatch dashboards for monitoring

## Using These Patterns

To apply these patterns to a new project:

1. **Start with the Meta-Prompt**: Establish the TDD IaC agent persona
2. **Use Issue Template**: Structure each feature as an issue
3. **Follow TDD Cycle**: Red → Green → Refactor for every feature
4. **Maintain Architecture Sync**: Keep docs and code aligned
5. **Apply Testing Patterns**: Use CDK assertions effectively
6. **Configure Environments**: Set up multi-environment from start
7. **Build in Security**: Apply security checklist from day one
8. **Enable Observability**: Add logging, metrics, alarms early

---

**Note**: These patterns were extracted from a real TDD IaC project (CDK Sleep Audio Pipeline) and represent battle-tested approaches for building AWS infrastructure with strict TDD discipline.
