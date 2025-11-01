#!/bin/bash

# Deploy Telegram Bot to Google Cloud Run
# Usage: ./scripts/deploy-cloudrun.sh [action]
# Actions: setup, build, deploy, logs, all

set -e

# Configuration
PROJECT_ID=$(gcloud config get-value project)
REGION="europe-west1"
SERVICE_NAME="territory-bot"
REPOSITORY_NAME="territory-bot-repo"
IMAGE_NAME="territory-bot"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    # Check if gcloud is installed
    if ! command -v gcloud &> /dev/null; then
        log_error "gcloud CLI is not installed. Please install it first:"
        log_error "  macOS: brew install google-cloud-sdk"
        exit 1
    fi
    
    # Check if project is set
    if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "(unset)" ]; then
        log_error "No GCP project set. Please run:"
        log_error "  gcloud config set project YOUR_PROJECT_ID"
        exit 1
    fi
    
    log_info "Requirements check passed. Project: $PROJECT_ID"
}

setup_artifact_registry() {
    log_info "Setting up Artifact Registry..."
    
    # Check if repository already exists
    if gcloud artifacts repositories describe $REPOSITORY_NAME \
        --location=$REGION &> /dev/null; then
        log_warn "Repository $REPOSITORY_NAME already exists. Skipping creation."
    else
        log_info "Creating Artifact Registry repository..."
        gcloud artifacts repositories create $REPOSITORY_NAME \
            --repository-format=docker \
            --location=$REGION \
            --description="Docker repository for Territory Service Bot"
        log_info "Repository created successfully!"
    fi
    
    # Configure Docker to use gcloud as credential helper
    log_info "Configuring Docker authentication..."
    gcloud auth configure-docker ${REGION}-docker.pkg.dev
    
    log_info "Artifact Registry setup complete!"
}

build_and_push() {
    log_info "Building and pushing Docker image..."
    
    IMAGE_TAG="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY_NAME}/${IMAGE_NAME}:latest"
    
    log_info "Building image: $IMAGE_TAG (no cache)"
    docker build --no-cache -t $IMAGE_TAG .
    
    log_info "Pushing image to Artifact Registry..."
    docker push $IMAGE_TAG
    
    log_info "Image pushed successfully!"
    echo $IMAGE_TAG
}

deploy_to_cloudrun() {
    log_info "Deploying to Cloud Run..."
    
    IMAGE_TAG="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY_NAME}/${IMAGE_NAME}:latest"
    
    # Check required environment variables
    if [ -z "$TS_TELEGRAM_BOT_TOKEN" ]; then
        log_error "TS_TELEGRAM_BOT_TOKEN environment variable is required"
        log_error "Please set it before deploying:"
        log_error "  export TS_TELEGRAM_BOT_TOKEN='your-bot-token'"
        exit 1
    fi
    
    if [ -z "$TS_POSTGRESQL_HOST" ]; then
        log_error "TS_POSTGRESQL_HOST environment variable is required"
        log_error "Please set database connection variables:"
        log_error "  export TS_POSTGRESQL_HOST='your-db-host'"
        log_error "  export TS_POSTGRESQL_USER='your-db-user'"
        log_error "  export TS_POSTGRESQL_PASSWORD='your-db-password'"
        log_error "  export TS_POSTGRESQL_DATABASE='your-db-name'"
        exit 1
    fi
    
    log_info "Deploying service: $SERVICE_NAME"
    
    gcloud run deploy $SERVICE_NAME \
        --image=$IMAGE_TAG \
        --region=$REGION \
        --platform=managed \
        --cpu=1 \
        --memory=512Mi \
        --min-instances=1 \
        --max-instances=1 \
        --timeout=3600 \
        --cpu-throttling \
        --no-cpu-boost \
        --set-env-vars="TS_LOG_LEVEL=${TS_LOG_LEVEL:-info}" \
        --set-env-vars="TS_POSTGRESQL_HOST=${TS_POSTGRESQL_HOST}" \
        --set-env-vars="TS_POSTGRESQL_USER=${TS_POSTGRESQL_USER}" \
        --set-env-vars="TS_POSTGRESQL_PASSWORD=${TS_POSTGRESQL_PASSWORD}" \
        --set-env-vars="TS_POSTGRESQL_DATABASE=${TS_POSTGRESQL_DATABASE}" \
        --set-env-vars="TS_TELEGRAM_BOT_TOKEN=${TS_TELEGRAM_BOT_TOKEN}" \
        --no-allow-unauthenticated
    
    log_info "Deployment complete!"
    log_info "View logs with: make cloud-logs"
}

stream_logs() {
    log_info "Streaming logs from Cloud Run..."
    log_info "Press Ctrl+C to stop streaming"
    echo ""
    gcloud beta logging tail "resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME" \
        --format="value(timestamp,severity,textPayload,jsonPayload.message)" \
        --project=$PROJECT_ID
}

show_status() {
    log_info "Service status:"
    gcloud run services describe $SERVICE_NAME --region=$REGION --format="table(
        status.url,
        status.conditions[0].type,
        status.conditions[0].status,
        metadata.annotations.run.googleapis.com/lastModifier
    )"
}

# Main script
case "${1:-}" in
    setup)
        check_requirements
        setup_artifact_registry
        ;;
    build)
        check_requirements
        build_and_push
        ;;
    deploy)
        check_requirements
        deploy_to_cloudrun
        ;;
    logs)
        check_requirements
        stream_logs
        ;;
    status)
        check_requirements
        show_status
        ;;
    all)
        check_requirements
        setup_artifact_registry
        build_and_push
        deploy_to_cloudrun
        ;;
    *)
        echo "Usage: $0 {setup|build|deploy|logs|status|all}"
        echo ""
        echo "Actions:"
        echo "  setup  - Create Artifact Registry repository"
        echo "  build  - Build and push Docker image"
        echo "  deploy - Deploy to Cloud Run"
        echo "  logs   - Stream Cloud Run logs"
        echo "  status - Show service status"
        echo "  all    - Run setup, build, and deploy"
        echo ""
        echo "Required environment variables:"
        echo "  TS_TELEGRAM_BOT_TOKEN"
        echo "  TS_POSTGRESQL_HOST"
        echo "  TS_POSTGRESQL_USER"
        echo "  TS_POSTGRESQL_PASSWORD"
        echo "  TS_POSTGRESQL_DATABASE"
        echo ""
        echo "Optional environment variables:"
        echo "  TS_LOG_LEVEL (default: info)"
        exit 1
        ;;
esac

