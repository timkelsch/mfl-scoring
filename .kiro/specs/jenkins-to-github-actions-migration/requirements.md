# Requirements Document

## Introduction

This feature migrates the existing Jenkins CI/CD pipeline to GitHub Actions while maintaining identical functionality. The migration addresses the pain point of managing a hibernating EC2 instance for infrequent deployments by leveraging GitHub's hosted runners. The solution maintains the current deployment workflow: feature branches deploy to STAGE, and merging to main automatically promotes to PROD.

## Requirements

### Requirement 1

**User Story:** As a developer, I want to push code changes to feature branches and have them automatically deployed to the STAGE environment, so that I can test changes without managing Jenkins infrastructure.

#### Acceptance Criteria

1. WHEN a developer pushes to any non-main branch THEN the system SHALL run linting and tests
2. WHEN linting and tests pass THEN the system SHALL build a Docker container and push to ECR with auto-versioning
3. WHEN the container is pushed THEN the system SHALL update the Lambda function and STAGE alias
4. WHEN deployment completes THEN the system SHALL provide the same STAGE environment functionality as the current Jenkins pipeline

### Requirement 2

**User Story:** As a developer, I want to merge PRs to main and have them automatically promote to production, so that the manual approval step is the PR merge itself.

#### Acceptance Criteria

1. WHEN a PR is merged to main branch THEN the system SHALL automatically promote the current STAGE version to PROD
2. WHEN promotion occurs THEN the system SHALL use the same promotion logic as the existing Jenkins pipeline
3. WHEN promotion completes THEN the system SHALL update the PROD alias to point to the promoted version
4. IF promotion fails THEN the system SHALL fail the workflow and provide clear error messages

### Requirement 3

**User Story:** As a developer, I want AWS credentials managed through the existing CloudFormation infrastructure, so that the CI/CD user is defined alongside other pipeline resources.

#### Acceptance Criteria

1. WHEN setting up the migration THEN the system SHALL add IAM user and policies to the existing storage.yaml CloudFormation template (to be renamed to bootstrap.yaml)
2. WHEN the IAM user is created THEN it SHALL have minimum required permissions for ECR, Lambda, and existing AWS operations
3. WHEN credentials are generated THEN they SHALL be manually copied to GitHub repository secrets
4. WHEN credentials are used THEN they SHALL work with the existing AWS account and region (us-east-1)

### Requirement 4

**User Story:** As a developer, I want the same build and deployment process as Jenkins, so that there are no behavioral changes or new functionality added.

#### Acceptance Criteria

1. WHEN the migration is complete THEN the system SHALL use the same Docker build process as the current Dockerfile
2. WHEN versioning occurs THEN the system SHALL use the same ECR auto-versioning logic as push.sh
3. WHEN Lambda updates occur THEN the system SHALL use the same update commands as the current Makefile
4. WHEN promotion occurs THEN the system SHALL use the same promotion script logic
5. IF any existing functionality is changed THEN it SHALL only be to fix serious issues, not add features

### Requirement 5

**User Story:** As a developer, I want to maintain the existing manual CloudFormation deployment process, so that infrastructure changes remain controlled and match current workflow.

#### Acceptance Criteria

1. WHEN infrastructure changes are needed THEN the system SHALL continue using manual Makefile commands (createstack, updatestack, etc.)
2. WHEN CloudFormation templates are updated THEN they SHALL be deployed manually using existing processes, not automated in GitHub Actions
3. WHEN the migration is complete THEN the system SHALL not change the current manual CloudFormation workflow
4. WHEN the bootstrap.yaml template is updated with the CI/CD user THEN it SHALL be deployed manually using make updatebootstrapstack (updated Makefile target)

### Requirement 6

**User Story:** As a developer, I want the migration to be simple and maintainable, so that it doesn't add complexity to an already over-engineered personal project.

#### Acceptance Criteria

1. WHEN creating workflows THEN the system SHALL use straightforward GitHub Actions without complex orchestration
2. WHEN defining jobs THEN the system SHALL mirror the existing Jenkins stages as closely as possible
3. WHEN handling errors THEN the system SHALL provide clear, actionable error messages
4. WHEN the migration is complete THEN it SHALL require minimal ongoing maintenance