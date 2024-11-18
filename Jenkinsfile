/* groovylint-disable CompileStatic */
pipeline {
    agent any

    tools {
        go 'go1.23.3'
        git 'Default'
    }

    environment {
        CGO_ENABLED = 0
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE = "${WORKSPACE}"
        MAIN_BRANCH = 'main'
    }

    options {
        // prevent dual pushes at PR merge from blowing us up
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '2', artifactNumToKeepStr: '2'))
    }

    stages {
        stage('Checkout') {
            when {
                not { branch "${ env.MAIN_BRANCH }" }
            }
            steps {
                script {
                    if (fileExists("${env.WORKSPACE}/")) {
                        echo 'Workspace is empty. Checking out from Git. '
                        checkout scm
                    } else {
                        echo 'Workspace is not empty. Skipping Git checkout.'
                    }
                    branchName = scm.branches[0].name
                    echo "Current branch: ${branchName}"
                }
            }
        }

        stage('Lint/Test') {
            when {
                not { branch "${ env.MAIN_BRANCH }" }
            }
            steps {
                withEnv(["PATH+GO=${GOPATH}/bin"]) {
                    echo 'installing golangci-lint'
                    sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                        sh -s -- -b $(go env GOPATH)/bin v1.62.0'
                    echo 'Running test'
                    sh 'make test'
                    echo 'Running lint'
                    sh 'make lint'
                    echo 'linting complete'
                }
            }
        }

        stage('Build, Push, Add Lambda Version') {
            when {
                not { branch "${ env.MAIN_BRANCH }" }
            }
            steps {
                script {
                    echo 'Pushing to STAGE'
                    RETURN_CODE = sh(
                        script: 'make push',
                        returnStatus: true
                    )
                    echo "Push to STAGE Return Status: ${RETURN_CODE}"
                    if (RETURN_CODE != 0) {
                        currentBuild.result = 'ABORTED'
                        error('Stopping due to error. Check log messages.')
                    }
                }
            }
        }

        stage('Deploy to Stage') {
            when {
                not { branch "${ env.MAIN_BRANCH }" }
            }
            steps {
                echo 'Deploying to stage'
                sh 'make updatestagealias'
            }
        }

        stage('Deploy to Prod') {
            when {
                branch "${ env.MAIN_BRANCH }"
            }
            steps {
                // This will have to more specific if we have more than one branch
                // under development at once.
                echo 'Promoting from STAGE to PROD alias'
                sh 'make promote'
            }
        }
    }

    post {
        always {
            // cleanWs()
            script {
                echo 'Running Docker prune'
                RETURN_CODE = sh(
                    script: 'docker system prune -a -f',
                    returnStatus: true
                )
                echo "returnStatus: ${RETURN_CODE}"
            }
        }
    }
}
