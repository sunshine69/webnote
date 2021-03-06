pipeline {
    agent { label 'master' }

    options {
        ansiColor('xterm')
    }

    stages {
        stage('Checkout') {
            steps {
                script {
                env.BUILD_VERSION = VersionNumber projectStartDate: '2019-12-15', versionNumberString: "${BUILD_NUMBER}", versionPrefix: "1.", worstResultForIncrement: 'SUCCESS'
                echo "Version:  ${BUILD_VERSION}"

                GIT_REVISION = sh(returnStdout: true, script: """
                    git rev-parse --short HEAD"""
                ).trim()
                def PWD = pwd()
                echo "Check out REVISION: $GIT_REVISION on $PWD"
                DO_GATHER_ARTIFACT_BRANCH = (['master'].contains(GIT_BRANCH) ||
                    GIT_BRANCH ==~ /release\-[\d\-\.]+/ ||
                    GIT_BRANCH ==~ /[^\s]+enable_docker_image_push$/)
                checkout changelog: false, poll: false, scm: [$class: 'GitSCM', branches: [[name: '*/jenkins']], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: 'jenkins-helper']], submoduleCfg: [], userRemoteConfigs: [[credentialsId: 'github-personal-jenkins', url: 'https://github.com/sunshine69/jenkins-helper.git']]]
                utils = load("${WORKSPACE}/jenkins-helper/deployment.groovy")
                sh "git clean -fdx"
                }//script
            }//steps
        }//stage

        stage('Generate build scripts') {
            steps {
                script {
                  //utils.generate_add_user_script()
                  //utils.generate_aws_environment()
                  sh '''cat <<EOF > build.sh
#!/bin/sh -x
rm -f webnote-go-bin-*.tgz
BUILD_VERSION=${BUILD_VERSION} ./build-jenkins.sh
EOF
'''
                  sh 'chmod +x build.sh'
                }//script
            }//steps
        }

        stage('Run the command within the docker environment') {
            steps {
                script {
                    utils.run_build_script([
//Make sure you build this image ready - having user jenkins and cache go mod for that user.
                        'docker_image': 'golang-alpine-build-jenkins:latest',
                        'docker_entrypoint_opt': '--entrypoint sleep',
                        'docker_args_opt': '3600',
                        'docker_net_opt': '',
//define the name here so we can save a image cache in the script save-docker-image-cache.sh
                        'docker_extra_opt': '--name golang-alpine-build-jenkins',
//Uncomment these when you build with the golang-alpine from scratch. After we
//can commented out as the image is saved
                        //'outside_scripts': ['save-docker-image-cache.sh'],
                        //'extra_build_scripts': ['fix-godir-ownership.sh'],
                        //'run_as_user': ['fix-godir-ownership.sh': 'root'],
                    ])
                    if (GIT_BRANCH ==~ /(master|jenkins)/) {//Build for ARM x96
                    withCredentials([usernamePassword(credentialsId: 'github-personal-jenkins', passwordVariable: 'GITHUB_TOKEN', usernameVariable: 'GITHUB_USER')]) {
                    env.REPOSITORY = "webnote"
                    sh '''cat <<EOF > build-arm-auto-gen.sh
#!/bin/sh

cd ~/webnote
rm -f webnote-go-bin-*.tgz
git fetch http://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${REPOSITORY} refs/heads/${GIT_BRANCH}
git reset --hard
git checkout FETCH_HEAD
BUILD_VERSION=${BUILD_VERSION} ./build-jenkins.sh
ls webnote-go-bin-*.tgz
EOF
'''
                    sh 'chmod +x build-arm-auto-gen.sh'
                    sshagent(['jenkins-to-x96']) {
                        sh 'scp -P 1969 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null build-arm-auto-gen.sh stevek@192.168.20.20:build-arm-auto-gen.sh'
                        sh 'ssh -p 1969 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null stevek@192.168.20.20 ./build-arm-auto-gen.sh'
                        sh 'ssh -p 1969 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null stevek@192.168.20.20 rm -f build-arm-auto-gen.sh'
                        sh 'rm -f webnote-go-bin-*armv*.tgz'
                        sh "scp -P 1969 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null 'stevek@192.168.20.20:webnote/webnote-go-bin-*.tgz' ."
                        sh 'ssh -p 1969 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null stevek@192.168.20.20 rm -f webnote/webnote-go-bin-*.tgz'
                        sh 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@192.168.20.21 rm -f webnote-go-bin-*armv*.tgz'
                        sh "scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null webnote-go-bin-*armv*.tgz root@192.168.20.21:"
                        //it never returns even & in the deploy command no matter what I tried
                        //sh "ssh -t -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@192.168.20.21 ./deploy-webnote.sh webnote-go-bin-*armv*.tgz"
                        echo "You have to go there and deploy manually"
                    }//sshagent
                    }//withCred
                    }//If GIT_BRANCH
                }//script
            }//steps
        }//stage

        stage('Gather artifacts') {
            steps {
                script {
                    utils.save_build_data(['artifact_class': 'webnote'])

                    if (DO_GATHER_ARTIFACT_BRANCH) {
                      archiveArtifacts allowEmptyArchive: true, artifacts: 'webnote-go-bin-*.tgz', fingerprint: true, onlyIfSuccessful: true
                      if (GIT_BRANCH ==~ /master/ ) {
                        echo "Create a release as this is a master merge ..."
                        withCredentials([usernamePassword(credentialsId: 'github-personal-jenkins', passwordVariable: 'GITHUB_TOKEN', usernameVariable: 'GITHUB_USER')]) {
                            env.REPOSITORY = "webnote"
                            sh """
                            git tag v${BUILD_VERSION}; git push http://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${REPOSITORY} --tags
                            ARTIFACT_FILES=\$(ls webnote-go-bin-*.tgz)
                            GITHUB_USER=$GITHUB_USER REPOSITORY=${REPOSITORY} GITHUB_TOKEN=$GITHUB_TOKEN ARTIFACT_FILES=\${ARTIFACT_FILES} ./create-github-release.sh
                            """
                        }//withCred
                      }//If GIT_BRANCH
                    }//If DO_GATHER_ARTIFACT_BRANCH
                    else {
                      echo "Not collecting artifacts as branch"
                    }// If GIT_BRANCH
                } //script
            }
        }// Gather artifacts

    }
    post {
        always {
            script {
                utils.apply_maintenance_policy_per_branch()
                currentBuild.description = """Artifact version: ${BUILD_VERSION}
Artifact revision: ${GIT_REVISION}"""
            }
        }
        success {
            script {
              cleanWs cleanWhenFailure: false, cleanWhenNotBuilt: false, cleanWhenUnstable: false, deleteDirs: true, disableDeferredWipeout: true, notFailBuild: true, patterns: [[pattern: 'PerformanceTesting', type: 'INCLUDE'], [pattern: '*', type: 'INCLUDE']]
            }//script
        }
        failure {
            script {
                //slackSend baseUrl: 'https://xvt.slack.com/services/hooks/jenkins-ci/', botUser: true, channel: '#errcd-activity', message: "@here CRITICAL - ${JOB_NAME} (${BUILD_URL}) branch (${BRANCH_NAME}) revision (${GIT_REVISION}) on (${NODE_NAME})", teamDomain: 'xvt', tokenCredentialId: 'jenkins-ci-integration-token', color: "danger"
                echo 'Build failed'
            }//script
        }
    }
}
