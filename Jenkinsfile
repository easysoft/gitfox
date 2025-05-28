library 'z-common@feat-240314'

pipeline {
  agent {
    kubernetes {
      inheritFrom "dind2 xuanim"
      yaml '''
      spec:
        nodeSelector:
          ci-cpu-level: high
      '''
    }
  }

  environment {
      REGISTRY_HOST="hub.zentao.net"
      DOCKER_CREDENTIALS_ID = 'hub-qucheng-push'
      BUILD_IMAGE="${REGISTRY_HOST}/app/gitfox:dev"
      BUILD_FEAT_IMAGE="${REGISTRY_HOST}/app/gitfox:feat-dev"
  }

  stages {
    stage("Prepare") {
      steps {
        script {
          env.XIM_USERS = sh(returnStdout: true, script: 'git show -s --format=%ce').trim()
        }
      }
    }

    stage("Build Docker Image") {
      when {
        branch 'zentao'
      }
      steps {
        script {
          dockerBuildx(host=env.REGISTRY_HOST, credentialsId=env.DOCKER_CREDENTIALS_ID) {
              sh "docker buildx build --push --pull -t $BUILD_IMAGE ."
          }
        }
      }

      post {
        success {
          ximNotify(title: "Gitfox镜像构建成功", content: "$BUILD_IMAGE")
        }
      }

    }

    stage("Build Feat Docker Image") {
      when {
        branch 'feat'
      }
      steps {
        script {
          dockerBuildx(host=env.REGISTRY_HOST, credentialsId=env.DOCKER_CREDENTIALS_ID) {
              sh "docker buildx build --push --pull -t $BUILD_FEAT_IMAGE ."
          }
        }
      }

      post {
        success {
          ximNotify(title: "Gitfox上游特性分支镜像构建成功", content: "$BUILD_FEAT_IMAGE")
        }
      }

    }
  }

  post {
    failure {
       ximNotify(title: "Gitfox构建失败")
    }
  }
}
