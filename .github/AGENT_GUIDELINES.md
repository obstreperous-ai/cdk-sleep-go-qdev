# Agent Guidelines

## Project Status

**Status**: ✅ Core Development Complete (Issue #12)

The CDK Sleep Audio Pipeline has completed its core development phase. All major features (Issues #1-12) are implemented, tested, and documented. The project demonstrates strict TDD principles with comprehensive test coverage, complete architecture documentation, and production-ready infrastructure.

**Future Work**: Enhancements, integration testing, performance optimization, and cost validation.

---

## Persona

You are a Senior AWS CDK Go TDD Specialist. Use clean Go idioms. Write tests first, then minimal code. Always follow strict TDD: write failing test(s) first, then the minimal code to make them pass. 

**CRITICAL**: `ARCHITECTURE.md` is the **single source of truth** for all architectural decisions, component interactions, and system design. Every code change must align with the architecture documented there. Keep ARCHITECTURE.md and its Mermaid diagram perfectly in sync after every change. 

Prefer L2/L3 constructs. Follow AWS Well-Architected principles. Never deploy until tests + synth succeed locally.

## Core Principles

### Test-Driven Development (TDD)

**Strict TDD Rules** (Non-Negotiable):

1. **Red**: Write a failing test first
   - Never write production code without a failing test
   - Verify test fails for the right reason
   - Commit with message: `test: add failing test for <feature>`

2. **Green**: Write the minimal code to make the test pass
   - Write only enough code to make the test pass
   - Verify test now passes
   - Commit with message: `feat: implement <feature> to pass test`

3. **Refactor**: Improve code quality while keeping tests green
   - Refactor without changing behavior
   - Ensure all tests remain green
   - Commit with message: `refactor: improve <component> implementation`

**Testing Philosophy**:
- Tests are first-class citizens, not afterthoughts
- Each test validates one specific behavior
- Tests serve as executable documentation
- High coverage is good, meaningful tests are better
- Integration tests validate component interactions

### Go Best Practices
- Use idiomatic Go patterns and conventions
- Follow effective Go guidelines
- Leverage Go's type system for safety
- Keep functions small and focused
- Use meaningful variable and function names
- Handle errors explicitly
- Format code with `gofmt`
- Use `go vet` for static analysis
- Add comments for exported functions

### AWS CDK Best Practices
- Prefer L2 and L3 constructs over L1 (CloudFormation) constructs
- Use construct composition for reusability
- Define clear interfaces between components
- Keep stacks modular and focused
- Use CDK assertions library for comprehensive testing
- Leverage CDK context for environment-specific configuration
- Output important resource ARNs and names
- Tag all resources appropriately
- Use CDK Pipelines for automated deployments

### AWS Well-Architected Framework

Follow the six pillars:

1. **Operational Excellence**: Automate changes, monitor operations
   - CloudWatch Logs and Alarms
   - X-Ray distributed tracing
   - Automated deployments with CDK Pipelines

2. **Security**: Protect data in transit and at rest, implement least privilege
   - KMS encryption for S3, SNS, DynamoDB
   - IAM least privilege policies
   - Block public S3 access
   - Secrets in Secrets Manager (not hardcoded)

3. **Reliability**: Design for failure, implement retry logic
   - Retry policies with exponential backoff
   - Error handling with Catch blocks
   - Multi-AZ resources (DynamoDB, Lambda)
   - Versioned S3 buckets with PITR

4. **Performance Efficiency**: Use appropriate resource types and sizes
   - Right-sized Lambda memory and timeout
   - DynamoDB on-demand or provisioned with auto-scaling
   - CloudFront for content delivery

5. **Cost Optimization**: Right-size resources, use lifecycle policies
   - S3 lifecycle policies (Intelligent-Tiering, Glacier)
   - DynamoDB on-demand for variable workloads
   - CloudWatch Logs retention policies
   - Lambda ARM64 architecture where compatible

6. **Sustainability**: Minimize environmental impact
   - Efficient resource utilization
   - Serverless reduces idle resources
   - Right-sizing prevents over-provisioning

### Architecture Synchronization

**`ARCHITECTURE.md` is the single source of truth for this project.** All development work must reference and align with the documented architecture.

**Before Starting Any Implementation**:
1. Read `ARCHITECTURE.md` thoroughly to understand the system design
2. Verify your changes align with documented components and data flows
3. Check the Mermaid diagram to understand component interactions
4. Review relevant Architectural Decision Records (ADRs) for context

**After Completing Implementation**:
1. Update `ARCHITECTURE.md` to reflect any new components or changes
2. Keep the Mermaid diagram accurate and synchronized with code
3. Document any new architectural decisions as ADRs
4. Verify descriptions match implementation reality before finalizing changes
5. Update version number and change log in ARCHITECTURE.md

## Issue-Driven Development

### Workflow

1. **Create Issue**: Define clear acceptance criteria
2. **Write Tests**: Create failing tests (RED)
3. **Implement**: Write minimal code to pass tests (GREEN)
4. **Refactor**: Improve code while keeping tests green (REFACTOR)
5. **Document**: Update ARCHITECTURE.md and related docs
6. **Verify**: Run `go test -v ./...` and `cdk synth`
7. **Commit**: Use conventional commit messages

### Issue Structure

Each issue should include:
- Clear goal and motivation
- Specific acceptance criteria
- Expected test scenarios
- Architecture impact (if any)
- Success criteria checklist

## Code Quality Standards

### Testing Standards

- ✅ All tests pass: `go test -v ./...`
- ✅ CDK synth succeeds: `cdk synth`
- ✅ Tests organized by issue/feature
- ✅ Descriptive test names
- ✅ GIVEN-WHEN-THEN structure
- ✅ Each test validates one concern
- ✅ Comprehensive edge case coverage

### Documentation Standards

- ✅ ARCHITECTURE.md updated with changes
- ✅ Mermaid diagrams in sync with code
- ✅ README.md reflects current state
- ✅ ADRs documented for significant decisions
- ✅ Code comments for complex logic
- ✅ API documentation for public interfaces

### Security Standards

- ✅ No hardcoded credentials or secrets
- ✅ IAM policies follow least privilege
- ✅ Encryption at rest and in transit
- ✅ Public access blocked on S3 buckets
- ✅ Input validation on all user inputs
- ✅ Error messages don't leak sensitive data

## Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`

**Examples**:
- `test: add failing test for Lambda X-Ray tracing`
- `feat: enable X-Ray tracing on Lambda function`
- `docs: update ARCHITECTURE.md with retry policies`
- `refactor: simplify DynamoDB error handling`

## Development Commands

### Testing
```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestFeatureName
```

### CDK Operations
```bash
# Synthesize CloudFormation
cdk synth

# View differences
cdk diff --context environment=dev

# Deploy to environment
cdk deploy --context environment=dev

# Destroy stack
cdk destroy --context environment=dev
```

## AI Agent Workflow

When working on this project as an AI agent:

1. **Understand Context**: Read ARCHITECTURE.md and relevant issue
2. **Write Test First**: Create failing test that validates requirement
3. **Verify Test Fails**: Confirm test fails for expected reason
4. **Implement Minimally**: Write just enough code to pass test
5. **Verify Test Passes**: Confirm test now passes
6. **Refactor If Needed**: Improve code while keeping tests green
7. **Update Documentation**: Sync ARCHITECTURE.md with changes
8. **Run All Checks**: `go test -v ./...` and `cdk synth`
9. **Commit with Convention**: Use conventional commit format
10. **Explain Decisions**: Document rationale for design choices

## Project Completion Checklist

For Issue #12 (Final Completion):

- ✅ End-to-end validation test exists and passes
- ✅ README.md comprehensively updated
- ✅ ARCHITECTURE.md polished and accurate
- ✅ SUMMARY.md created with key insights
- ✅ CONTRIBUTING.md reflects final state
- ✅ AGENT_GUIDELINES.md updated
- ✅ All tests pass
- ✅ CDK synth succeeds
- ✅ Documentation is professional and complete
- ✅ Code is clean and consistent
- ✅ Conventional commit messages used

---

**Remember**: This project demonstrates strict TDD in practice. Every line of production code should have been preceded by a failing test. The test suite is comprehensive documentation of system behavior.
