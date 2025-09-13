# Design Document

## Overview

This design migrates the existing Jenkins CI/CD pipeline to GitHub Actions while maintaining identical functionality and deployment patterns. The solution eliminates the need to manage a hibernating EC2 instance by leveraging GitHub's hosted runners, while preserving the current two-stage deployment model (STAGE for feature branches, PROD for main branch).

The migration maintains the existing build processes, versioning logic, and deployment commands to ensure zero behavioral changes. Infrastructure credentials are managed through the renamed bootstrap.yaml CloudFormation template.

## Architecture

### Current Jenkins Pipeline Flow
```
Feature Branch Push → Jenkins → Lint/Test → Build/Push → Deploy to STAGE
Main Branch Push → Jenkins → Promote STAGE to PROD
```

### New GitHub Actions Flow
```
Feature Branch Push → GitHub Actions → Lint/Test → Build/Push → Deploy to STAGE
Main Branch Push → GitHub Actions → Promote STAGE to PROD
```

### Workflow Structure

**1. CI Workflow (`.github/workflows/ci.yml`)**
- **Trigger**: Push to any branch, Pull Requests
- **Purpose**: Fast feedback for code quality
- **Jobs**: Lint, Test, Build validation (no deployment)
- **Runtime**: ~3-5 minutes

**2. Deploy to Stage Workflow (`.github/workflows/deploy-stage.yml`)**
- **Trigger**: Push to non-main branches
- **Purpose**: Deploy feature branches to STAGE environment
- **Jobs**: Build → Push to ECR → Update Lambda STAGE alias
- **Runtime**: ~5-8 minutes

**3. Deploy to Production Workflow (`.github/workflows/deploy-prod.yml`)**
- **Trigger**: Push to main branch (after PR merge)
- **Purpose**: Promote STAGE version to PROD automatically
- **Jobs**: Execute promotion script logic
- **Runtime**: ~2-3 minutes

## Components and Interfaces

### GitHub Actions Workflows

#### CI Workflow Components
- **Go Setup**: Uses actions/setup-go@v5 with Go 1.23.3
- **Linting**: Integrates with existing .golangci.yml configuration
- **Testing**: Runs `go test -cover` in mfl-scoring directory
- **Build Validation**: Ensures Docker build succeeds without pushing

#### Stage Deployment Components
- **AWS Authentication**: Uses configured IAM user credentials
- **ECR Login**: Authenticates to existing ECR repository
- **Version Management**: Replicates push.sh auto-versioning logic
- **Lambda Update**: Uses existing AWS CLI commands from Makefile
- **Alias Management**: Updates STAGE alias to $LATEST

#### Production Deployment Components
- **Promotion Logic**: Replicates scripts/promote_stage_to_prod.sh
- **Version Resolution**: Gets current STAGE version
- **Alias Update**: Points PROD alias to STAGE version

### Infrastructure Components

#### Bootstrap CloudFormation Template (bootstrap.yaml)
- **Existing Resources**: S3 buckets, KMS keys (unchanged)
- **New Resources**:
  - IAM User for GitHub Actions
  - IAM Policy with minimum required permissions
  - Access Keys for programmatic access

#### IAM Permissions Required
- **ECR**: GetAuthorizationToken, BatchCheckLayerAvailability, GetDownloadUrlForLayer, BatchGetImage, PutImage, InitiateLayerUpload, UploadLayerPart, CompleteLayerUpload
- **Lambda**: UpdateFunctionCode, PublishVersion, UpdateAlias, GetAlias, ListVersionsByFunction
- **General**: sts:GetCallerIdentity for account operations

### Integration Points

#### ECR Integration
- **Repository**: 287140326780.dkr.ecr.us-east-1.amazonaws.com/mfl-score
- **Versioning**: Maintains existing semantic versioning (v1.v2 format)
- **Image Comparison**: Replicates push.sh logic to avoid duplicate pushes

#### Lambda Integration
- **Function**: Uses existing FUNCTION_NAME environment variable resolution
- **Aliases**: Maintains STAGE and PROD alias pattern
- **Architecture**: Preserves arm64 architecture specification

#### Docker Integration
- **Build Process**: Uses existing Dockerfile without modifications
- **Platform**: Maintains linux/amd64 build target
- **Multi-stage**: Preserves Alpine-based runtime image

## Data Models

### Workflow Environment Variables
```yaml
AWS_REGION: us-east-1
AWS_ACCOUNT: ${{ secrets.AWS_ACCOUNT_ID }}
FUNCTION_NAME: mfl-scoring-[generated-suffix]
```

### GitHub Secrets Required
```
AWS_ACCESS_KEY_ID: [from bootstrap.yaml IAM user]
AWS_SECRET_ACCESS_KEY: [from bootstrap.yaml IAM user]
AWS_ACCOUNT_ID: 287140326780
```

### Version Management Data Flow
```
Current ECR Version → Increment Logic → New Version → Tag → Push → Lambda Update
```

## Error Handling

### Build Failures
- **Lint Errors**: Fail fast with golangci-lint output
- **Test Failures**: Display test results and coverage
- **Docker Build Errors**: Show build context and error details

### Deployment Failures
- **ECR Push Failures**: Retry with exponential backoff
- **Lambda Update Failures**: Preserve existing version, fail workflow
- **Alias Update Failures**: Rollback to previous state if possible

### Version Conflicts
- **Duplicate Images**: Skip push if image content identical (existing logic)
- **Version Resolution**: Handle edge cases in promotion script
- **Concurrent Deployments**: Use GitHub's concurrency controls

## Testing Strategy

### Pre-Migration Testing
1. **Validate IAM Permissions**: Test all required AWS operations
2. **Verify ECR Access**: Confirm push/pull operations work
3. **Test Lambda Updates**: Ensure function updates succeed
4. **Validate Promotion**: Test STAGE to PROD promotion logic

### Migration Testing Approach
1. **Parallel Testing**: Run both Jenkins and GitHub Actions initially
2. **Feature Branch Testing**: Deploy test branches to validate STAGE deployment
3. **Production Validation**: Test promotion logic on non-critical changes
4. **Rollback Testing**: Ensure ability to revert to Jenkins if needed

### Post-Migration Validation
1. **End-to-End Testing**: Full feature branch to production flow
2. **Performance Comparison**: Ensure deployment times are acceptable
3. **Monitoring**: Verify all existing functionality works identically
4. **Documentation**: Update README with new deployment process

### Automated Testing Integration
- **Existing Tests**: Preserve current `go test` execution
- **Coverage Reporting**: Maintain existing coverage requirements
- **Lint Standards**: Use existing .golangci.yml configuration
- **Build Validation**: Ensure Docker builds succeed before deployment

## Migration Strategy

### Phase 1: Infrastructure Setup
1. Update storage.yaml → bootstrap.yaml with IAM resources
2. Deploy bootstrap stack with `make updatebootstrapstack`
3. Configure GitHub repository secrets
4. Update Makefile targets for new naming

### Phase 2: Workflow Creation
1. Create CI workflow for immediate feedback
2. Create Stage deployment workflow
3. Create Production deployment workflow
4. Test workflows on feature branches

### Phase 3: Cutover
1. Disable Jenkins pipeline
2. Update documentation
3. Remove Jenkins-specific files
4. Monitor first production deployment

### Rollback Plan
- Keep Jenkins configuration files until migration proven
- Maintain ability to re-enable Jenkins if critical issues arise
- Document rollback procedure for emergency situations