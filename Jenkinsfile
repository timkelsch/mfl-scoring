pipeline {
    agent any
    tools { go 'go1.21' }
    environment {
        CGO_ENABLED = 0 
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE="${WORKSPACE}"
        SAM_CLI_TELEMETRY=0
    }
    stages {        
        stage('Pre Test') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
            }
        }
        
        stage('Build') {
            steps {
                echo 'Compiling and building'
                sh 'echo $GOCACHE'
                sh 'make test'
            }
        }

        stage('Test') {
            steps {
                withEnv(["PATH+GO=${GOPATH}/bin"]){
                    echo 'installing golangci-lint'
                    sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                        sh -s -- -b $(go env GOPATH)/bin v1.54.2'
                    echo 'Running lint'
                    sh 'make lint'
                    echo 'Running test'
                    sh 'make test'
                }
            }
        }
        
        stage('Deploy') {       
            steps {
                sh 'make jsambuild'
                sh 'make jsamdeploy'
            }
        }
    }
}