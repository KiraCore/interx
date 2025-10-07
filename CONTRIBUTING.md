# Contributing to KIRA Interoperability Microservices

Thank you for your interest in contributing to the KIRA Interoperability Microservices project! This document outlines the git workflow and contribution process.

## Git Workflow

This project follows a feature branch workflow with `master` as the main branch.

### Branch Structure

- **master**: The main production-ready branch. All features are merged here.
- **feature/\***: Feature branches for new functionality or improvements.
- **fix/\***: Branches for bug fixes.
- **hotfix/\***: Branches for urgent production fixes.

### Development Workflow

#### 1. Fork and Clone (External Contributors)

```bash
# Fork the repository on GitHub first, then:
git clone git@github.com:YOUR_USERNAME/interx.git
cd interx
git remote add upstream git@github.com:KiraCore/interx.git
```

For team members with direct access:

```bash
git clone git@github.com:KiraCore/interx.git
cd interx
```

#### 2. Create a Feature Branch

Always create a new branch from the latest `master`:

```bash
# Update your local master branch
git checkout master
git pull origin master

# Create and switch to a new feature branch
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/description` - For new features
- `fix/description` - For bug fixes
- `hotfix/description` - For urgent production fixes

#### 3. Make Your Changes

- Write clean, maintainable code
- Follow existing code style and conventions
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

#### 4. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "brief: Description of changes"
```

Commit message format:
- Use imperative mood ("Add feature" not "Added feature")
- First line should be concise (50 characters or less)
- Add detailed description in the body if needed

Examples:
```
cicd: Add new pipeline
fix: Resolve cosmos indexer timeout issue
feature: Implement Ethereum transaction signing
update: Enhance P2P load balancing logic
```

#### 5. Keep Your Branch Updated

Regularly sync your branch with master to avoid conflicts:

```bash
git checkout master
git pull origin master
git checkout feature/your-feature-name
git rebase master
```

If conflicts occur, resolve them and continue:

```bash
git add .
git rebase --continue
```

#### 6. Push Your Branch

```bash
git push origin feature/your-feature-name
```

If you've rebased, you may need to force push:

```bash
git push -f origin feature/your-feature-name
```

#### 7. Create a Pull Request

1. Go to the [repository on GitHub](https://github.com/KiraCore/interx)
2. Click "Pull requests" → "New pull request"
3. Select your branch to merge into `master`
4. Fill in the PR template with:
   - **Title**: Clear, concise description of changes
   - **Description**: What changes were made and why
   - **Testing**: How the changes were tested
   - **Related Issues**: Link any related issues

#### 8. Code Review Process

- Wait for code review from maintainers
- Address any feedback or requested changes
- Push additional commits to your branch as needed
- Once approved, a maintainer will merge your PR

#### 9. After Merge

Clean up your local branches:

```bash
git checkout master
git pull origin master
git branch -d feature/your-feature-name
```

## CI/CD Pipeline

The project uses automated CI/CD pipelines. When you push to your branch or create a PR:

- Automated tests will run
- Code quality checks will be performed
- Build verification will occur

Ensure all checks pass before requesting review.

## Best Practices

### Code Quality

- Write self-documenting code with clear variable and function names
- Add comments for complex logic
- Follow Go best practices and idioms
- Handle errors appropriately
- Avoid code duplication

### Testing

- Write unit tests for new functionality
- Ensure existing tests still pass
- Test edge cases and error conditions
- Run tests locally before pushing:

```bash
go test ./...
```

### Docker and Microservices

- Test your changes with Docker Compose locally
- Ensure all microservices work together correctly
- Update configuration files if needed
- Document any new environment variables or config options

### Documentation

- Update README.md if adding new features
- Document API changes
- Add inline code documentation
- Update configuration examples

## Release Process

For information about creating releases, see [RELEASE.md](RELEASE.md).

Releases are automatically created when feature branches are merged to master. The CI/CD pipeline handles:
- Version tagging
- Docker image building and publishing
- Automated changelog generation
- Image signing with cosign
- Release notes with SHA256 digests

## Getting Help

- Check existing issues and discussions
- Read the [README.md](README.md) for architecture overview
- Read the [RELEASE.md](RELEASE.md) for release process
- Ask questions in pull request comments
- Contact maintainers if needed

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to KIRA Interoperability Microservices! 🚀
