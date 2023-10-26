pipeline {
    agent any
    options {
        buildDiscarder(logRotator(numToKeepStr: '5', artifactNumToKeepStr: '5'))
    }    

    tools { go 'go1.21' } 

    environment {
        CGO_ENABLED=0 
        GOPATH="${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE="${WORKSPACE}"
        SAM_CLI_TELEMETRY=0
    }

    stages {        
        stage('Check for Modified Files') {
            steps {
                script {
                    def changes = currentBuild.changeSets.poll()
                    if (changes.isEmpty()) {
                        currentBuild.result = 'ABORTED'
                        error("No changes detected. Pipeline aborted.")
                    }

                    // Define the list of files you want to check for changes
                    def filesToCheck = ['Dockerfile', 'mfl-scoring/*']

                    for (entry in changes) {
                        for (file in filesToCheck) {
                            if (entry.affectedPaths.contains(file)) {
                                echo "Found changes in ${file}. Proceeding with the pipeline."
                            }
                        }
                    }
                }
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
        
        stage('Build, Push, Update Lambda') {
            steps {
                sh 'make push'
            }
        }

        stage('Deploy'){
            steps{
                sh 'make updatestagealias'
            }        
        }
    }
}