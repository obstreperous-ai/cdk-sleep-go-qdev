# CDK Sleep Go Pipeline

An event-driven, serverless AWS CDK application built with Go that processes, analyzes, and delivers personalized sleep audio content. The pipeline leverages S3, EventBridge, Lambda, DynamoDB, and SNS to create a scalable, resilient architecture following AWS Well-Architected Framework principles and strict Test-Driven Development (TDD) practices.

## Project Philosophy

This project is **TDD-first and issue-driven**. Every feature and fix follows the Red-Green-Refactor cycle.

### Strict TDD Rules

1. **RED**: Write a failing test first - no exceptions
2. **GREEN**: Write minimal code to make the test pass
3. **REFACTOR**: Improve code while keeping tests green
4. **VERIFY**: Ensure `cdk synth` succeeds before committing
5. **NEVER DEPLOY**: Until all tests pass and synthesis succeeds

These rules are non-negotiable. Every commit must demonstrate this workflow.

## Quick Start

```bash
# Install dependencies
go mod download

# Run tests (must always pass)
go test -v ./...

# Synthesize CloudFormation template
cdk synth

# Deploy (only after tests pass)
cdk deploy
```

## Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)**: Complete architecture design and Mermaid diagrams
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Development workflow, TDD guidelines, and code standards
- **[.github/AGENT_GUIDELINES.md](.github/AGENT_GUIDELINES.md)**: AI agent persona and specialized development guidelines
