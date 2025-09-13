# GitHub Actions Migration Guide

This guide walks you through migrating from Jenkins to GitHub Actions for the MFL Scoring project.

## Overview

The migration replaces Jenkins with three GitHub Actions workflows:
- **CI Workflow**: Runs on all pushes and Pull Requests for code quality checks
- **Stage Deployment**: Deploys feature branches to STAGE environment
- **Production Deployment**: Promotes STAGE to PROD when merging to main

## Prerequisites

- AWS CLI configured with appropriate permissions
- Access to AWS CloudFormation console
- GitHub repository admin access

## Step 1: Deploy Bootstrap Infrastructure

1. **Update the bootstrap CloudFormation stack:**
   ```bash
   make updatebootstrapstack
   ```

2. **Get the GitHub Actions credentials from CloudFormation outputs:**
   - Go to AWS CloudFormation console
   - Find the "bootstrap" stack
   - Go to "Outputs" tab
   - Copy the following values:
     - `GitHubActionsCicdAccessKeyId`
     - `GitHubActionsCicdSecretAccessKey`

## Step 2: Configure GitHub Repository Secrets

1. **Go to your GitHub repository settings:**
   - Navigate to Settings → Secrets and variables → Actions

2. **Add the following repository secrets:**
   ```
   AWS_ACCESS_KEY_ID = [GitHubActionsCicdAccessKeyId from CloudFormation]
   AWS_SECRET_ACCESS_KEY = [GitHubActionsCicdSecretAccessKey from CloudFormation]
   AWS_ACCOUNT_ID = 287140326780
   AWS_REGION = us-east-1
   ```

## Step 3: Test the Migration

1. **Create a test feature branch:**
   ```bash
   git checkout -b test-github-actions
   git push origin test-github-actions
   ```

2. **Verify the workflows run:**
   - Check GitHub Actions tab in your repository
   - CI workflow should run and pass
   - Deploy to Stage workflow should run and deploy to STAGE

3. **Test the STAGE environment:**
   - Verify your application works in the STAGE environment
   - Check Lambda function and aliases in AWS console

4. **Test production deployment:**
   ```bash
   git checkout main
   git merge test-github-actions
   git push origin main
   ```
   - Deploy to Production workflow should run
   - Verify PROD environment is updated

## Step 4: Clean Up

1. **Disable/remove Jenkins infrastructure:**
   - Stop Jenkins EC2 instance
   - Remove Jenkins security groups and other resources

2. **Update documentation:**
   - Update README.md deployment instructions
   - Remove Jenkins references

## Workflow Details

### CI Workflow (`.github/workflows/ci.yml`)
- **Triggers**: All pushes and Pull Requests
- **Jobs**:
  - Lint and test Go code
  - Validate Docker build
- **Duration**: ~3-5 minutes

### Stage Deployment (`.github/workflows/deploy-stage.yml`)
- **Triggers**: Push to non-main branches
- **Jobs**:
  - Build and push Docker image to ECR
  - Update Lambda function with new image
  - Update STAGE alias to point to latest version
- **Duration**: ~5-8 minutes

### Production Deployment (`.github/workflows/deploy-prod.yml`)
- **Triggers**: Push to main branch
- **Jobs**:
  - Get current STAGE version
  - Update PROD alias to point to STAGE version
- **Duration**: ~2-3 minutes

## Troubleshooting

### Common Issues

1. **AWS Credentials Error:**
   - Verify secrets are correctly set in GitHub
   - Check IAM user has required permissions
   - Ensure AWS_ACCOUNT_ID matches your account
   - Ensure AWS_REGION is set correctly (us-east-1)

2. **ECR Push Failures:**
   - Verify ECR repository exists: `mfl-score`
   - Check ECR permissions in IAM policy
   - Ensure AWS_REGION secret matches your ECR region

3. **Lambda Update Failures:**
   - Verify Lambda function exists and starts with "mfl-scoring"
   - Check Lambda permissions in IAM policy
   - Ensure function architecture matches (arm64)

4. **Version Resolution Issues:**
   - Check ECR repository has existing images
   - Verify image tags follow semantic versioning (x.y format)

### Rollback Procedure

If you need to rollback to Jenkins:

1. **Re-enable Jenkins:**
   - Start Jenkins EC2 instance
   - Restore Jenkinsfile from `archive/Jenkinsfile.backup`

2. **Disable GitHub Actions:**
   - Rename workflow files to `.disabled` extension
   - Or delete the `.github/workflows/` directory

3. **Update deployment process:**
   - Use Jenkins for deployments
   - Update documentation accordingly

## IAM Permissions

The GitHub Actions user has the following permissions:

### ECR Access
- GetAuthorizationToken, BatchCheckLayerAvailability
- GetDownloadUrlForLayer, BatchGetImage
- PutImage, InitiateLayerUpload, UploadLayerPart, CompleteLayerUpload
- DescribeImages, DescribeRepositories

### Lambda Access
- UpdateFunctionCode, PublishVersion
- UpdateAlias, GetAlias, ListVersionsByFunction
- GetFunction, ListFunctions

### General Access
- sts:GetCallerIdentity

## Support

For issues with the migration:
1. Check GitHub Actions logs for detailed error messages
2. Verify AWS CloudFormation stack deployed successfully
3. Ensure all secrets are correctly configured
4. Test AWS CLI access with the provided credentials

## End-to-End Flow Example

### Feature Development Flow:
1. Create feature branch: `git checkout -b feature/new-scoring`
2. Make changes and push: `git push origin feature/new-scoring`
3. Create PR: GitHub Actions runs CI workflow (lint, test, build validation)
4. GitHub Actions runs Stage deployment workflow
5. Test changes in STAGE environment
6. Create PR and merge to main
7. GitHub Actions runs Production deployment workflow
8. Changes are live in PROD

### Infrastructure Changes Flow:
Infrastructure changes (CloudFormation templates) remain manual:
1. Update CloudFormation templates
2. Deploy using Makefile: `make updatestack`, `make updatebootstrapstack`, etc.
3. No automation in GitHub Actions for infrastructure changes