pipeline {
    agent any

    parameters {
        booleanParam(name: 'autoDeploy', defaultValue: true, description: '是否自动部署到服务器')
    }

    stages {
        stage('build'){
            steps {
              echo 'fetch source code'
              git 'https://github.com/jiezi19971225/ahpuoj-backend'
              echo 'start build image'
              script{
                docker.withRegistry('https://ccr.ccs.tencentyun.com', 'dockerAccount') {
                  def customImage = docker.build("ccr.ccs.tencentyun.com/jiezi19971225/ahpuoj-backend:${env.BUILD_NUMBER}")
                  customImage.push()
                }
              }
            }
        }
        stage('deploy'){
          when {
            expression { params.autoDeploy == true }
          }
          steps {
            echo 'start deploy to school oj server'
            sshPublisher(publishers: [sshPublisherDesc(
              configName: 'schoolOJ',
              transfers: [sshTransfer(
                cleanRemote: false,
                excludes: '',
                execCommand: '''
                  echo "开始构建后操作"
                  cd /home/ahpuoj/ahpuojDocker/compose
                  docker-compose pull backend
                  docker-compose up -d backend
                  docker image prune -f --filter "dangling=true"
                ''',
                execTimeout: 120000,
                flatten: false,
                makeEmptyDirs: false,
                noDefaultExcludes: false,
                patternSeparator: '[, ]+',
                remoteDirectory: '/home/ahpuoj/ahpuojDocker/compose',
                remoteDirectorySDF: false,
                removePrefix: '',
                sourceFiles: '')],
              usePromotionTimestamp: false,
              useWorkspaceInPromotion: false,
              verbose: false)])
          }
        }
    }
    post {
        always {
            echo '🎉 done!!! 🎉'
        }
    }
}