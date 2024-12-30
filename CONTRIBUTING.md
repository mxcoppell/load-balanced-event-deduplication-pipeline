# Contributing to Key Expiration Test Workgroup

Thank you for your interest in contributing to this project! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/your-username/k8s-key-expiration-workgroup.git
cd k8s-key-expiration-workgroup
```
3. Add the upstream repository:
```bash
git remote add upstream https://github.com/mxcoppell/k8s-key-expiration-workgroup.git
```

## Development Workflow

1. Create a new branch for your feature or fix:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following our coding standards:
   - Use meaningful variable and function names
   - Add comments for complex logic
   - Follow Go style guidelines (run `go fmt` and `go vet`)
   - Write tests for new functionality

3. Test your changes:
   - Run unit tests: `go test ./...`
   - Build and deploy locally following README.md instructions
   - Verify functionality in the WebUI
   - Check consumer logs for expected behavior

4. Commit your changes:
   - Write clear, concise commit messages
   - Use present tense ("Add feature" not "Added feature")
   - Reference relevant issues

5. Keep your branch updated:
```bash
git fetch upstream
git rebase upstream/main
```

## Pull Request Process

1. Push your changes to your fork:
```bash
git push origin feature/your-feature-name
```

2. Create a Pull Request (PR) from your fork to our main branch

3. In your PR description:
   - Clearly describe the changes
   - Explain the motivation
   - List any breaking changes
   - Include steps to test the changes

4. Update documentation if needed:
   - README.md for user-facing changes
   - Code comments for implementation details
   - design.md for architectural changes

5. Address review feedback:
   - Make requested changes
   - Push updates to your branch
   - Respond to comments

## Testing Guidelines

1. Unit Tests
   - Write tests for new functions
   - Maintain existing test coverage
   - Use table-driven tests where appropriate

2. Integration Tests
   - Test interactions between services
   - Verify Redis and NATS integration
   - Check metrics collection

3. Performance Testing
   - Test with various key generation rates
   - Verify consumer scalability
   - Monitor resource usage

## Areas for Contribution

- Performance improvements
- Additional metrics and monitoring
- UI enhancements
- Documentation improvements
- Bug fixes
- Test coverage

## Getting Help

- Create an issue for bugs or feature requests
- Ask questions in pull requests
- Contact maintainers for guidance

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License. 