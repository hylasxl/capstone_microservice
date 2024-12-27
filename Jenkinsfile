pipeline {
    agent any

    environment {
        DOCKER_IMAGE = 'vietpham18/capstone_gateway'
        DOCKER_TAG = 'latest'
    }

    stages {
        stage('Clone Repository') {
            steps {
                git branch: 'master', url: 'https://github.com/hylasxl/capstone_microservice'
            }
        }
        stage('Build Docker Image') {
                    steps {
                        script {
                            // Change to the gateway directory and build the image
                            sh 'cd gateway && docker build -t vietpham18/capstone_gateway:latest .'
                        }
                    }
                }

        stage('Push  to Docker Hub') {
            steps {
                script {
                    docker.withRegistry('https://index.docker.io/v1/', 'docker-hub-credentials') {
                        docker.image("${DOCKER_IMAGE}:${DOCKER_TAG}").push()
                    }
                }
            }
        }

        stage('Run Full Project') {
            steps {
                script {
                    echo 'Running all services with Docker Compose...'
                    sh '''
                        docker-compose down || echo "No containers to stop"
                        docker-compose up -d --build
                    '''
                }
            }
        }

        stage('Send Notification') {
            steps {
                script {
                    try {
                        sh '''
                            curl -s -X POST https://api.telegram.org/bot7686490744:AAF3MWixwEm0e6SJZu520Uu8pNmYNB2q7VU/sendMessage \
                            -d chat_id=-4696233151 \
                            -d text="All services started successfully!"
                        '''
                    } catch (Exception e) {
                        sh '''
                            curl -s -X POST https://api.telegram.org/bot7686490744:AAF3MWixwEm0e6SJZu520Uu8pNmYNB2q7VU/sendMessage \
                            -d chat_id=-4696233151 \
                            -d text="Error starting services: ${e.message}"
                        '''
                        throw e
                    }
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
