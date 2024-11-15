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
    }

    options {
        // prevent dual pushes at PR merge from blowing us up
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '2', artifactNumToKeepStr: '2'))
    }

    stages {
        stage('Checkout') {
            steps {
                script {
                    if (fileExists("${env.WORKSPACE}/")) {
                        echo 'Workspace is empty. Checking out from Git.'
                        checkout scm
                    } else {
                        echo 'Workspace is not empty. Skipping Git checkout.'
                    }
                }
            }
        }

        stage('Lint/Test') {
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
            steps {
                script {
                    echo 'KAPUSHHHH'
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
