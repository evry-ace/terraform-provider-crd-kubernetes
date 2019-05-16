#!groovy
@Library("ace") _

properties([disableConcurrentBuilds()])

def isMaster = "${env.BRANCH_NAME}" == 'master'
def isPR = "${env.CHANGE_URL}".contains('/pull/')

opts = [
  buildAgent: 'jenkins-docker-3',
]

ace(opts) {
  def goVer = "1.12.1"

  def args = [
    "-v ${pwd()}:/src",
    "-w /src",
    "-e GOCACHE=/tmp/.GOCACHE",
    "-e GOPATH=/src/.GO",
    "-e CI=1"
  ]

  stage('Setup') {
    docker.image("golang:${goVer}").inside(args.join(' ')) {
      sh """
        make setup
      """
    }
  }

  stage("test") {
    docker.image("golang:${goVer}").inside(args.join(' ')) {
        sh """
        make test
        make cover
        """
    }
  }

  stage('Build') {
    docker.image("golang:${goVer}").inside(args.join(' ')) {
      sh """
        make
      """
    }

    // cache(maxCacheSize: 1000, caches: [
    //   [$class: 'ArbitraryFileCache', includes: '**/*', path: '.GOPATH/src/github.com'],
    //   [$class: 'ArbitraryFileCache', includes: '**/*', path: '.GOPATH/go/src/golang.org'],
    //   [$class: 'ArbitraryFileCache', includes: '**/*', path: '.GOPATH/go/src/gopkg.in'],
    //   [$class: 'ArbitraryFileCache', includes: '**/*', path: '.GOPATH/go/src/google.golang.org']
    // ])

    dockerBuild()
  }

  stage('Push') {
    dockerPush()
  }

  stage('Deploy') {
    if (isMaster) {
      deploy("test")
    }

    // slack.notifyDeploy('test')
  }
}
