pipeline {
    agent any
    options {
        buildDiscarder(logRotator(numToKeepStr: '3', artifactNumToKeepStr: '3'))
    }    

    tools { go 'go1.21' } 

    environment {
        CGO_ENABLED=0 
        GOPATH="${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOCACHE="${WORKSPACE}"
    }

    stages {        
        // Too annoying to block on code / lambda changes here
        // Already blocking duplicates at push to ECR stage 

        // stage('Check for Modified Files') {
        //     steps {
        //         script {
        //             def changeLogSets = currentBuild.changeSets
        //             if (changeLogSets.isEmpty()) {
        //                 currentBuild.result = 'ABORTED'
        //                 error("ChangeLogSets is empty. Pipeline aborted.")
        //             }

        //             // Define the list of files you want to check for changes
        //             def filesToCheck = ['Dockerfile', 'mfl-scoring/main.go', 'mfl-scoring/main_test.go',
        //                 'mfl-scoring/go.mod', 'mfl-scoring/go.sum']
                    
        //             def numFilesToCheckChanged = 0
        //             for (changeLogSet in changeLogSets) {
        //                 for (entry in changeLogSet) {
        //                     for (file in filesToCheck) {
        //                         echo "File: " + file
        //                         echo "AffectedPaths: " + entry.getAffectedPaths()
        //                         if (entry.getAffectedPaths().contains(file)) {
        //                             echo "${file} was modified"
        //                             numFilesToCheckChanged++
        //                         }
        //                     }
        //                 }
        //             }

        //             if (numFilesToCheckChanged > 0) {
        //                 echo "Found changes. Proceeding with the pipeline."
        //             } else {
        //                 currentBuild.result = 'ABORTED'
        //                 error("No changes detected. Pipeline aborted.")
        //             }                
        //         }
        //     }
        // }

        stage('Test') {
            parallel {
                stage('lint') {
                    steps {
                        echo 'installing golangci-lint'
                        sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
                            sh -s -- -b $(go env GOPATH)/bin v1.54.2'
                        echo 'Running golangci-lint'
                        sh 'make lint'
                    }
                }
                stage('test') {
                    steps {
                        withEnv(["PATH+GO=${GOPATH}/bin"]){
                            echo 'Running unit tests'
                            sh 'make test'
                        }
                    }
                }
            }
        }

        stage('Build, Push, Update Lambda') {
            steps {
                script {
                    RETURN_CODE = sh (
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

        stage('Deploy'){
            steps{
                echo 'Deploying to stage'
                sh 'make updatestagealias'
            }
        }

        stage('Clean Up') {
            steps {
                script {
                    echo 'Running Docker prune'
                    RETURN_CODE = sh (
                        script: 'docker system prune -a -f',
                        returnStatus: true
                    )
                    echo "returnStatus: ${RETURN_CODE}"
                }
            }
        }
    }
}