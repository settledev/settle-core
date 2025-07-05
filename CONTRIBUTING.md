# Contributing to Settle Core

Thank you for your interest in contributing to Settle Core! This document provides guidelines and information for contributors.

## 🤝 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to see if your problem has already been reported
2. **Check the documentation** to ensure you're using the tool correctly
3. **Provide detailed information** including:
   - Operating system and version
   - Go version
   - Steps to reproduce the issue
   - Expected vs actual behavior
   - Any error messages or logs

### Suggesting Features

We welcome feature suggestions! When proposing a new feature:

1. **Check the roadmap** to see if it's already planned
2. **Explain the use case** and why it would be valuable
3. **Consider the impact** on existing functionality
4. **Think about implementation** - is it feasible within our constraints?

## 🤝 Development Setup

### Prerequisites

- Go 1.23.0 or later
- Git
- SSH access to test servers (for testing)

### Local Development

1. **Fork and clone** the repository:

   ```bash
   git clone https://github.com/yourusername/settle-core.git
   cd settle-core
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Build the project**:

   ```bash
   go build -o settlectl
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

### Testing Your Changes

1. **Unit tests**: Run `go test ./...`
2. **Integration tests**: Test with real SSH connections
3. **Manual testing**: Try your changes with example configurations

## 📝 Code Style Guidelines

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Add comments for complex logic
- Use meaningful variable names

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

type(scope): description
[optional body]
[optional footer]

Examples:

- `feat(cmd): add new ping command`
- `fix(core): resolve circular dependency issue`
- `docs(readme): update installation instructions`

### Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** following the style guidelines
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Run tests** and ensure they pass
6. **Submit a pull request** with a clear description

### Pull Request Guidelines

- **Title**: Clear, concise description
- **Description**: Explain what and why (not how)
- **Related issues**: Link to any related issues
- **Testing**: Describe how you tested your changes
- **Breaking changes**: Clearly mark any breaking changes

## 🏗️ Project Architecture

### Key Components

- **`cmd/`**: CLI commands and user interface
- **`core/`**: Core engine, graph logic, and state management
- **`drivers/`**: Package managers and service implementations
- **`inventory/`**: Host management and SSH connectivity
- **`common/`**: Shared types and constants

### Adding New Resources

To add a new resource type:

1. **Define the resource** in `core/resources.go`
2. **Implement the driver** in `drivers/`
3. **Add parsing logic** in `core/parser.go`
4. **Update tests** and documentation
5. **Add examples** in `examples/`

### Adding New Commands

To add a new CLI command:

1. **Create the command** in `cmd/`
2. **Follow the existing pattern** using Cobra
3. **Add help text** and usage examples
4. **Update the root command** if needed
5. **Add tests** for the new command

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./cmd -run TestPingCommand
```

### Test Guidelines

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test with real SSH connections
- **Mock external dependencies**: Use interfaces for testability
- **Test error conditions**: Don't just test happy paths
- **Keep tests fast**: Avoid slow network calls in unit tests

## 📄 Documentation

### Code Documentation

- **Add comments** for exported functions and types
- **Use godoc** format for comments
- **Include examples** for complex functions
- **Document interfaces** and their contracts

### User Documentation

- **Update README.md** for user-facing changes
- **Add examples** for new features
- **Update help text** for CLI commands
- **Document breaking changes** clearly

## 🚀 Release Process

### For Contributors

1. **Ensure tests pass** on your branch
2. **Update version** if needed
3. **Update changelog** with your changes
4. **Create a release** if it's a significant feature

### For Maintainers

1. **Review pull requests** thoroughly
2. **Run integration tests** before merging
3. **Update version** and create tags
4. **Generate release notes** from commits
5. **Publish binaries** for supported platforms

## 🆘 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code reviews**: Ask questions in pull request comments
- **Documentation**: Check the wiki and examples

## 📋 Checklist for Contributors

Before submitting your contribution, ensure:

- [ ] Code follows style guidelines
- [ ] Tests are added and passing
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Pull request description is clear
- [ ] No breaking changes without discussion
- [ ] Examples are provided for new features

## 🙏 Recognition

Contributors will be recognized in:

- **README.md** contributors section
- **Release notes** for significant contributions
- **GitHub contributors** page

Thank you for contributing to Settle!

---

**Questions?** Open an issue or start a discussion on GitHub.
