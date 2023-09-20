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
                    echo 'installing golangci-lint'
                    sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2'
                    echo 'Running vetting'
                    sh 'cd mfl-scoring; go vet .'
                    //echo 'Running linting'
                    //sh 'cd mfl-scoring; $GOPATH/bin/golangci-lint run -v; ls -l'
                    echo 'Running test'
                    sh 'cd mfl-scoring; go test -v'
                }
            }
        }
        
        stage('Deploy') {       
            steps {
                sh '/var/jenkins_home/sam/venv/bin/sam build'
                sh '/var/jenkins_home/sam/venv/bin/sam deploy --no-confirm-changeset --no-fail-on-empty-changeset'
            }
        }
    }
}