pipeline {
    agent any
    
    tools { go 'go1.21' }

    environment {
        CGO_ENABLED=0 
        GOPATH="${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE="${WORKSPACE}"
        SAM_CLI_TELEMETRY=0
    }

    stages {        
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
        
        stage('Build') {
            steps {
                echo 'Compiling and building'
                sh 'make build'
            }
        }

        stage('Push'){
            steps{
                echo 'Compiling and building'
                sh 'make push'
            }
        }

        stage('Deploy'){
            steps{
                sh "make deploy"
            }        
        }
    }
}