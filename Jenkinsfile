pipeline {
    agent any
    tools { go 'go1.21' }
    environment {
        CGO_ENABLED = 0 
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE="${WORKSPACE}"
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
                sh 'cd mfl-scoring; go build'
            }
        }

        stage('Test') {
            steps {
                withEnv(["PATH+GO=${GOPATH}/bin"]){
                    echo 'Running vetting'
                    sh 'cd mfl-scoring; go vet .'
                    //echo 'Running linting'
                    echo 'Running test'
                    sh 'go test -v'
                }
            }
        }
        
    }
}