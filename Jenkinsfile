pipeline {
    options {
        timestamps()
    }
    environment {
        GOCACHE = "/tmp"
    }
    agent {
    docker {
      label 'jenkins-agent-cloud'
      image 'golang:1.17.6-buster'
    }
  }
    stages {
        stage("test") {
            steps {
                dir("${WORKSPACE}/src/glog") {
                    // update env.PATH
                    withEnv(["PATH+GO=${GOPATH}/bin:/usr/local/go/bin"]) {
                        git url: 'https://gerrit.corp.arista.io/glog'
                        // Fetch the changeset to a local branch using the build parameters provided to the
                        // build by the Gerrit plugin
                        script {
                            env.CHANGE = "change-${GERRIT_CHANGE_NUMBER}-${GERRIT_PATCHSET_NUMBER}"
                        }
                        sh "git fetch origin ${GERRIT_REFSPEC}:${env.CHANGE}"
                        sh "git checkout ${env.CHANGE}"
                        // in order for the updated path to be visible to the make script,
                        // we need to explicitly pass it
                        sh 'go get ./...'
                        sh 'go get -u golang.org/x/lint/golint'
                        sh 'PATH=' + env.PATH + ' make check'
                    }
                }
            }
        }
    }
}
