# GitHub Actions CI/CD

This directory contains GitHub Actions workflows for continuous integration and deployment.

## Workflows

### Test Workflow (`test.yml`)
Runs on every push and pull request to master/main/develop branches.

**Jobs:**
- **Test**: Runs Go tests with PostgreSQL and Redis services
  - Sets up Go 1.25
  - Configures PostgreSQL and Redis services
  - Runs tests with race detection and coverage
  - Uploads coverage to Codecov
  - Generates HTML coverage report

- **Lint**: Runs golangci-lint to check code quality
  - Uses the project's `.golangci.yml` configuration
  - Reports linting issues

- **Build**: Ensures the application builds successfully
  - Builds the binary
  - Uploads build artifact

### Docker Build Workflow (`docker.yml`)
Builds and pushes Docker images on push to master/main or on tags.

**Jobs:**
- **Docker**: Builds multi-platform Docker images
  - Builds for linux/amd64 and linux/arm64
  - Tags images based on branch/tag/SHA
  - Pushes to Docker Hub (requires secrets)
  - Runs Trivy security scanner
  - Uploads security results to GitHub

## Status Badges

Add these badges to your README.md:

```markdown
[![Test](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/test.yml/badge.svg)](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/test.yml)
[![Docker Build](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/docker.yml/badge.svg)](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/docker.yml)
[![codecov](https://codecov.io/gh/YOUR_USERNAME/YOUR_REPO/branch/master/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/YOUR_REPO)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/YOUR_REPO)](https://goreportcard.com/report/github.com/YOUR_USERNAME/YOUR_REPO)
```

Replace `YOUR_USERNAME` and `YOUR_REPO` with your actual GitHub username and repository name.

## Required Secrets

For the Docker workflow to work, you need to configure these secrets in your repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add the following secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub password or access token

## Environment Variables for Tests

The test workflow uses these environment variables:

```yaml
DB_HOST: localhost
DB_PORT: 5432
DB_USER: postgres
DB_PASSWORD: postgres
DB_NAME: testdb
DB_SSLMODE: disable
REDIS_HOST: localhost
REDIS_PORT: 6379
JWT_SECRET: test-secret-key-minimum-32-characters-long-for-testing
BCRYPT_COST: 4  # Lower cost for faster tests
LOG_LEVEL: error
ENV: test
```

## Local Testing

To run tests locally with the same environment:

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=testdb
export JWT_SECRET=test-secret-key-minimum-32-characters-long-for-testing
export BCRYPT_COST=4
export LOG_LEVEL=error
export ENV=test

# Run tests
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Customization

### Test Workflow
- Modify Go version in `.github/workflows/test.yml`
- Add more test databases or services as needed
- Adjust timeout values for long-running tests
- Configure additional coverage tools

### Docker Workflow
- Modify image repository in `.github/workflows/docker.yml`
- Add/remove target platforms
- Configure different tagging strategies
- Add deployment steps after build

### Lint Workflow
- Customize linters in `.golangci.yml`
- Add pre-commit hooks
- Configure IDE integrations
