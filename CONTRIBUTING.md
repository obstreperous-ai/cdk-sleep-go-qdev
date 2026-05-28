# Contributing to CDK Sleep Audio Pipeline

Thank you for your interest in contributing to the CDK Sleep Audio Pipeline project! This document provides guidelines and instructions for contributing.

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

## Pre-Commit Checklist

Before committing your changes, ensure:

- [ ] All tests pass: `go test -v ./...`
- [ ] CDK synth succeeds: `cdk synth`
- [ ] Code follows Go best practices
- [ ] ARCHITECTURE.md is updated if architecture changed
- [ ] Commit message follows conventional commit format

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

## Questions?

If you have questions or need clarification, please open an issue for discussion.
