/* groovylint-disable CompileStatic */
pipeline {
    agent any
    options {
        buildDiscarder(logRotator(numToKeepStr: '1', artifactNumToKeepStr: '1'))
    }

    tools {
        go 'go1.21'
        git 'Default'
    }

    environment {
        CGO_ENABLED = 0
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE = "${WORKSPACE}"
    }

    stages {
        stage('Clean Workspace') {
            steps {
                cleanWs()
            }
        }

        stage('Checkout') {
            steps {
                script {
                    checkout scm
                    sh 'ls -la'  // Print contents of the workspace
                }
            }
        }
        
        stage('Lint/Test') {
            steps {
                withEnv(["PATH+GO=${GOPATH}/bin"]) {
                    echo 'installing golangci-lint'
                    sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                        sh -s -- -b $(go env GOPATH)/bin v1.54.2'
                    echo 'Running test'
                    sh 'make test'
                    echo 'Running lint'
                    sh 'make lint'
                }
            }
        }

        stage('Build, Push, Add Lambda Version') {
            steps {
                script {
                    RETURN_CODE = sh(
                        script: 'make push',
                        returnStatus: true
                    )
                    echo "returnStatus: ${RETURN_CODE}"
                    if (RETURN_CODE != 0) {
                        currentBuild.result = 'ABORTED'
                        error('Stopping due to error. Check log messages.')
                    }
                }
            }
        }

        stage('Deploy to Stage') {
            steps {
                echo 'Deploying to stage'
                sh 'make updatestagealias'
            }
        }
    }

    post {
        always {
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
