# Contributing to CDK Sleep Go Pipeline

Thank you for your interest in contributing to the Sleep Audio Pipeline project! This document outlines our development process and guidelines.

## Development Philosophy

This project follows **strict Test-Driven Development (TDD)** principles. Every change must follow the Red-Green-Refactor cycle:

1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve code while keeping tests green

## Prerequisites

- Go 1.25.0 or later
- Node.js 20+ (for AWS CDK CLI)
- AWS CDK CLI: `npm install -g aws-cdk`
- AWS CLI configured with appropriate credentials
- Git for version control

## Getting Started

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd cdk-sleep-go-qdev
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests to verify setup**
   ```bash
   go test -v ./...
   ```

4. **Synthesize CDK to verify CDK setup**
   ```bash
   cdk synth
   ```

## TDD Workflow (MANDATORY)

### For New Features

1. **Write the test first** in `*_test.go` files
   - Test should fail initially (Red)
   - Use descriptive test names: `TestFeatureName_Scenario_ExpectedBehavior`
   - Follow Go testing conventions and CDK assertions library

2. **Implement minimal code** to pass the test
   - Write only enough code to make the test pass (Green)
   - Avoid over-engineering or implementing unrequested features

3. **Refactor** if needed
   - Improve code quality, readability, maintainability
   - Ensure all tests remain green
   - Update architecture documentation if structure changes

4. **Verify CDK synthesis**
   ```bash
   cdk synth
   ```
   Must succeed before committing

### For Bug Fixes

1. **Write a test that reproduces the bug** (should fail)
2. **Fix the bug** (test should pass)
3. **Verify no regression** (all other tests still pass)

## Code Standards

- **Go Idioms**: Follow standard Go conventions and idioms
- **CDK Best Practices**: Prefer L2/L3 constructs over L1 (CloudFormation)
- **AWS Well-Architected**: Follow AWS Well-Architected Framework principles
- **Formatting**: Use `go fmt` and `go vet` before committing
- **Naming**: Clear, descriptive names for functions, variables, and resources
- **Comments**: Document exported functions and complex logic

## Commit Message Convention

Use conventional commits for clear history:

- `feat:` New feature or enhancement
- `fix:` Bug fix
- `test:` Adding or updating tests
- `docs:` Documentation changes
- `refactor:` Code refactoring without behavior change
- `chore:` Tooling, dependencies, or configuration changes
- `perf:` Performance improvements

Example: `feat: add S3 event notification for audio uploads`

## Architecture Alignment

Every code change must align with `ARCHITECTURE.md`. If your change affects architecture:

1. Update `ARCHITECTURE.md` description
2. Update the Mermaid diagram to reflect changes
3. Ensure diagram and code remain in perfect sync

## Pull Request Process

1. Ensure all tests pass: `go test ./...`
2. Ensure CDK synth succeeds: `cdk synth`
3. Update documentation if needed
4. Create PR with clear description of changes
5. Reference any related issues
6. Wait for CI checks to pass
7. Address review feedback promptly

## Testing Guidelines

- **Unit Tests**: Test individual functions and constructs
- **Integration Tests**: Use CDK assertions to verify CloudFormation templates
- **Coverage**: Aim for high test coverage, especially for business logic
- **Naming**: `TestFunctionName_Scenario_ExpectedOutcome`
- **Assertions**: Use `github.com/aws/aws-cdk-go/awscdk/v2/assertions` for CDK tests

Example test structure:
```go
func TestCdkBaseStack_CreatesResources_Successfully(t *testing.T) {
    // GIVEN
    app := awscdk.NewApp(nil)
    
    // WHEN
    stack := NewCdkBaseStack(app, "TestStack", nil)
    
    // THEN
    template := assertions.Template_FromStack(stack, nil)
    template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(1))
}
```

## Never Deploy Before Tests Pass

**CRITICAL RULE**: Never deploy to AWS until:
- All tests pass locally: `go test ./...`
- CDK synthesis succeeds: `cdk synth`
- CI pipeline is green

## Agent Persona Reference

For AI agents contributing to this project, refer to `.github/AGENT_GUIDELINES.md` for the permanent agent persona and specialized guidelines.

## Questions?

- Review existing code for patterns and examples
- Check `ARCHITECTURE.md` for design decisions
- Refer to AWS CDK Go documentation for CDK-specific questions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
