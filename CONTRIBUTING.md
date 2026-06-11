# Contributing to CDK Sleep Audio Pipeline

Thank you for your interest in contributing to the CDK Sleep Audio Pipeline project! This document provides guidelines and instructions for contributing.

## Project Status

**Current Status**: ✅ Core Development Complete (Issue #12)

The project has completed its core development phase with all major features implemented, tested, and documented. Future contributions are welcome for:
- Bug fixes and improvements
- Additional features from the roadmap
- Documentation enhancements
- Integration and performance testing
- Cost optimization

## Development Philosophy

This project follows **strict Test-Driven Development (TDD)** principles. All contributions must adhere to this approach.

## TDD Workflow

### The Red-Green-Refactor Cycle

1. **RED**: Write a failing test
   - Write a test that describes the desired functionality
   - Run the test and verify it fails (for the right reason)
   - Commit: `test: add failing test for <feature>`

2. **GREEN**: Make the test pass
   - Write the minimal code necessary to make the test pass
   - Run the test and verify it passes
   - Commit: `feat: implement <feature> to pass test`

3. **REFACTOR**: Improve the code
   - Refactor the code while keeping tests green
   - Ensure all tests still pass
   - Commit: `refactor: improve <component> implementation`

### Testing Requirements

- **Test First**: Never write production code without a failing test
- **Test Coverage**: Aim for high test coverage, but prioritize meaningful tests
- **Test Independence**: Each test should be independent and not rely on others
- **Test Naming**: Use descriptive test names that explain what is being tested
- **CDK Assertions**: Use the CDK assertions library to verify infrastructure

### Example Test Structure

```go
func TestFeatureName(t *testing.T) {
    defer jsii.Close()

    // GIVEN - Setup test environment
    app := awscdk.NewApp(nil)

    // WHEN - Execute the code under test
    stack := NewMyStack(app, "TestStack", nil)

    // THEN - Verify the results
    template := assertions.Template_FromStack(stack, nil)
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
    })
}
```

## Development Environment Setup

### Prerequisites

- Go 1.21+
- Node.js 18.x+
- AWS CDK 2.x (`npm install -g aws-cdk`)
- AWS CLI configured with credentials
- Python 3.12+ (for Lambda function)

### Setup Steps

1. Clone the repository
2. Run `go mod download` to install Go dependencies
3. Run `cd lambda/audio-processor && pip install -r requirements.txt` for Lambda dependencies
4. Bootstrap CDK: `cdk bootstrap aws://<account>/<region>` (first time only)
5. Run tests: `go test -v ./...`
6. Synthesize: `cdk synth`

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestFeatureName
```

## Pre-Commit Checklist

Before committing your changes, ensure:

- [ ] All tests pass: `go test -v ./...`
- [ ] CDK synth succeeds: `cdk synth`
- [ ] Code follows Go best practices
- [ ] ARCHITECTURE.md is updated if architecture changed
- [ ] Commit message follows conventional commit format
- [ ] Tests follow TDD principles (test-first)
- [ ] Documentation updated if public API changed
- [ ] Security best practices followed (no hardcoded secrets, least privilege IAM)

## Architecture Documentation

When making architectural changes:

1. Update `ARCHITECTURE.md` with clear descriptions
2. Update Mermaid diagrams to reflect changes
3. Add/update ADRs (Architectural Decision Records) for significant decisions
4. Keep diagrams in sync with implementation
5. Document rationale for design choices

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Examples:**
```
test: add failing test for S3 bucket encryption
feat: implement S3 bucket with KMS encryption
docs: update ARCHITECTURE.md with DynamoDB schema
refactor: improve error handling in Lambda processor
```

## Code Style Guidelines

### Go Code

- Follow standard Go formatting (`gofmt`)
- Use descriptive variable names
- Add comments for exported functions
- Keep functions focused and small
- Use CDK L2 constructs where available

### Python Code (Lambda)

- Follow PEP 8 style guide
- Use type hints where appropriate
- Add docstrings to functions
- Keep handlers focused and testable
- Use structured logging (JSON format)

## Issue-Driven Development

This project follows an issue-driven workflow:

1. **Create Issue**: Describe the feature or bug with clear acceptance criteria
2. **Write Tests**: Create failing tests based on acceptance criteria (RED)
3. **Implement**: Write minimal code to make tests pass (GREEN)
4. **Refactor**: Improve code while keeping tests green (REFACTOR)
5. **Document**: Update architecture docs and README
6. **Review**: Ensure all checks pass before merge

## Security Considerations

When contributing, ensure:

- No hardcoded credentials or secrets
- IAM policies follow least privilege principle
- All S3 buckets block public access
- Encryption enabled for data at rest and in transit
- Input validation implemented where applicable
- Error messages don't leak sensitive information

## Testing Best Practices

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **Infrastructure Tests**: Use CDK assertions to validate CloudFormation
- **Edge Cases**: Test failure scenarios and error handling
- **Naming**: Use descriptive test names that explain what's being validated

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`feat/my-feature` or `fix/my-bugfix`)
3. Follow TDD workflow (RED → GREEN → REFACTOR)
4. Ensure all tests pass and CDK synth succeeds
5. Update documentation as needed
6. Create pull request with clear description
7. Address review feedback
8. Await approval and merge

## Questions?

If you have questions or need clarification, please open an issue for discussion.

## Resources

- [AWS CDK Developer Guide](https://docs.aws.amazon.com/cdk/)
- [AWS CDK Go Examples](https://github.com/aws-samples/aws-cdk-examples/tree/master/go)
- [AWS Step Functions Best Practices](https://docs.aws.amazon.com/step-functions/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
