# Release Process

This document describes how to create a new release for KIRA Interoperability Microservices.

## Overview

The release process is fully automated via GitHub Actions. When a feature branch is merged to `master`, the CI/CD pipeline automatically:

1. Creates a new version tag (semantic versioning)
2. Builds and pushes all 7 Docker images to GitHub Container Registry
3. Signs all images with cosign for security verification
4. Generates an automated changelog from commit messages
5. Calculates and publishes SHA256 digests for all images
6. Creates a GitHub release with comprehensive release notes

## Triggering a Release

### Step 1: Create and Push Your Feature Branch

Work on your changes in a feature branch:

```bash
git checkout master
git pull origin master
git checkout -b feature/your-feature-name

# Make your changes
git add .
git commit -m "feature: Add new functionality"
git push origin feature/your-feature-name
```

### Step 2: Create a Pull Request

1. Go to https://github.com/KiraCore/interx
2. Click "Pull requests" → "New pull request"
3. Select `feature/your-feature-name` → `master`
4. Fill in PR details and create the PR

### Step 3: Merge to Master

Once the PR is reviewed and approved:

1. Merge the PR to `master` (use "Squash and merge" or "Merge commit")
2. The release workflow **automatically triggers** on PR merge

The workflow triggers for these branch patterns:
- `feature/*`
- `hotfix/*`
- `bugfix/*`
- `release/*`
- `major/*`

### Step 4: Monitor the Release

1. Go to **Actions** tab in GitHub
2. Watch the "Build, Release, and Sign Docker Images" workflow
3. The workflow will:
   - Create a new version tag (e.g., `v1.2.3`)
   - Build all 7 microservices
   - Push images to `ghcr.io/kiracore/interx/*`
   - Sign images with cosign
   - Generate changelog and release notes
   - Create GitHub release

### Step 5: Verify the Release

1. Go to **Releases** in GitHub
2. Check the newly created release
3. Verify:
   - ✅ Changelog is generated
   - ✅ Docker image list is present
   - ✅ SHA256 digests are included
   - ✅ All 7 images are listed

## Release Workflow Details

### Automatic Versioning

The workflow uses [github-tag-action](https://github.com/mathieudutour/github-tag-action) for automatic semantic versioning:

- **Default**: Minor version bump (e.g., `v1.2.0` → `v1.3.0`)
- Analyzes commit messages to determine version bump
- Creates annotated git tags

To control version bumps, use conventional commit messages:
- `fix:` → Patch version bump (v1.2.0 → v1.2.1)
- `feature:` → Minor version bump (v1.2.0 → v1.3.0)
- `BREAKING CHANGE:` → Major version bump (v1.2.0 → v2.0.0)

### Docker Images Built

The pipeline builds and publishes 7 Docker images:

| Service | Image | Description |
|---------|-------|-------------|
| Manager | `ghcr.io/kiracore/interx/manager` | P2P load balancer and HTTP server |
| Proxy | `ghcr.io/kiracore/interx/proxy` | Legacy HTTP request converter |
| Storage | `ghcr.io/kiracore/interx/storage` | MongoDB storage service |
| Cosmos Indexer | `ghcr.io/kiracore/interx/cosmos-indexer` | Indexes Cosmos blockchain data |
| Cosmos Interaction | `ghcr.io/kiracore/interx/cosmos-interaction` | Creates and publishes Cosmos transactions |
| Ethereum Indexer | `ghcr.io/kiracore/interx/ethereum-indexer` | Indexes Ethereum blockchain data |
| Ethereum Interaction | `ghcr.io/kiracore/interx/ethereum-interaction` | Creates and publishes Ethereum transactions |

### Image Signing with Cosign

All images are signed using [cosign](https://github.com/sigstore/cosign) for supply chain security.

**Verification** (requires cosign public key):
```bash
cosign verify --key cosign.pub ghcr.io/kiracore/interx/manager@sha256:<digest>
```

### Automated Changelog

The changelog is automatically generated from git commit history between the previous and current tag.

**Commit Message Format** (for best changelog results):

```
<type>: <description>

Examples:
- feature: Add Ethereum transaction batching
- fix: Resolve cosmos indexer timeout
- cicd: Update release workflow
- update: Enhance P2P load balancing
```

The changelog groups commits by type:
- ✨ **Features** - `feature:` prefix
- 🐛 **Bug Fixes** - `fix:` prefix
- 🔄 **Updates** - `update:` prefix
- 🚀 **CI/CD** - `cicd:` prefix
- 📝 **Other Changes** - all other commits

### Release Notes Structure

Generated release notes include:

1. **What's Changed** - Grouped changelog from commits
2. **Docker Images** - Table of all published images with tags
3. **Image Digests & Verification** - SHA256 digests and cosign verification instructions
4. **Installation** - Instructions for pulling and using the new version

## Manual Release (If Needed)

If you need to manually trigger a release or create a specific version:

### Option 1: Create a Release Branch

```bash
git checkout master
git pull origin master
git checkout -b release/v1.5.0

# Make any last-minute changes
git push origin release/v1.5.0
```

Then create a PR from `release/v1.5.0` → `master` and merge it.

### Option 2: Manual Tag (Not Recommended)

```bash
git checkout master
git pull origin master
git tag -a v1.5.0 -m "Release v1.5.0"
git push origin v1.5.0
```

**Note**: Manual tags won't trigger the automated release workflow. Use PR merge method instead.

## Hotfix Release

For urgent production fixes:

```bash
git checkout master
git pull origin master
git checkout -b hotfix/critical-bug-fix

# Fix the bug
git add .
git commit -m "fix: Critical security patch"
git push origin hotfix/critical-bug-fix
```

Create PR, merge to `master`, and the release will trigger automatically.

## Required GitHub Secrets

The release workflow requires these secrets (already configured by maintainers):

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `COSIGN_PRIVATE_KEY` - Private key for signing Docker images
- `COSIGN_PASSWORD` - Password for the cosign private key
- `REPO_ACCESS` - Token for triggering downstream repository events

## Troubleshooting

### Release workflow didn't trigger

**Check**:
1. Was the PR merged (not just closed)?
2. Is the branch name in the correct format (`feature/*`, `hotfix/*`, etc.)?
3. Check the Actions tab for any failed workflows

### Build failed

**Check**:
1. Review the workflow logs in Actions tab
2. Ensure all Dockerfiles are valid
3. Check if all dependencies are available

### Images not pushed

**Check**:
1. GitHub Container Registry permissions
2. `GITHUB_TOKEN` has write access
3. Repository settings → Actions → General → Workflow permissions

### Changelog is empty

**Check**:
1. Ensure commits follow conventional commit format
2. There are actual commits between tags
3. Commits are not merge commits (they're excluded)

## Best Practices

1. **Use Conventional Commits**: Prefix commits with `feature:`, `fix:`, etc.
2. **Descriptive Commit Messages**: Write clear, concise commit descriptions
3. **Test Before Merge**: Ensure all tests pass before merging to master
4. **Review Release Notes**: Check generated release notes after release
5. **Update Docker Compose**: Update `docker-compose.yml` to use the new version in downstream projects

## Post-Release Checklist

After a successful release:

- [ ] Verify all Docker images are available on `ghcr.io`
- [ ] Check release notes are complete and accurate
- [ ] Test pulling and running the new images
- [ ] Update documentation if needed
- [ ] Notify team members of the new release
- [ ] Update any dependent projects or infrastructure

## Version History

To see all releases:
```bash
git tag -l
```

To see changelog between versions:
```bash
git log v1.2.0..v1.3.0 --oneline --no-merges
```

---

For questions or issues with the release process, contact the maintainers or create an issue in the repository.
