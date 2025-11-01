# Territory Service Bot API

Telegram bot for managing territory assignments in congregations.

## Quick Start

### Local Development

```bash
# 1. Start PostgreSQL
make docker-up

# 2. Seed test data
make seed

# 3. Create .env file
cat > .env << 'EOF'
TS_LOG_LEVEL=debug
TS_POSTGRESQL_USER=admin
TS_POSTGRESQL_PASSWORD=postgres
TS_POSTGRESQL_HOST=localhost
TS_POSTGRESQL_DATABASE=api
TS_TELEGRAM_BOT_TOKEN=your_bot_token_here
EOF

# 4. Run the app
make run
```

Or use the automated script:
```bash
./scripts/quick-start.sh
```

### Cloud Deployment (Google Cloud Run + Supabase)

```bash
# 1. Setup Supabase (free tier)
# - Create project at supabase.com
# - Note connection pooler credentials (port 6543)

# 2. Setup Google Cloud
gcloud auth login
gcloud projects create territory-bot-prod
gcloud config set project territory-bot-prod
gcloud services enable run.googleapis.com artifactregistry.googleapis.com

# 3. Set environment variables
export TS_TELEGRAM_BOT_TOKEN="your-bot-token"
export TS_POSTGRESQL_HOST="aws-0-eu-west-1.pooler.supabase.com"
export TS_POSTGRESQL_USER="postgres.xxxxx"
export TS_POSTGRESQL_PASSWORD="your-password"
export TS_POSTGRESQL_DATABASE="postgres"

# 4. Deploy
make cloud-setup    # One-time setup
make cloud-all      # Build and deploy
```

## Essential Commands

### Development
```bash
make help           # Show all commands
make dev            # Complete setup + run
make build          # Build application
make run            # Run application
make test           # Run tests
```

### Docker
```bash
make docker-up      # Start PostgreSQL
make docker-down    # Stop PostgreSQL
make docker-clean   # Remove all data
```

### Database
```bash
make seed           # Seed test data
make seed-clean     # Clean + seed
make inspect        # View test data
make psql           # Connect to database
```

### Cloud (GCP + Supabase)
```bash
make cloud-setup    # Initialize Artifact Registry
make cloud-build    # Build and push image
make cloud-deploy   # Deploy to Cloud Run
make cloud-all      # Build + deploy
make cloud-logs     # Stream logs
make cloud-status   # Check service status
```

### Database Migration
```bash
# Migrate data from local to Supabase
SUPABASE_DB_HOST=db.xxx.supabase.co \
SUPABASE_DB_PASSWORD=yourpassword \
make migrate
```

## Test Data

The seed script creates:
- **3 Congregations**: Lorem Central, Ipsum North, Dolor South
- **5 Territory Groups**: Lorem District, Ipsum Heights, Dolor Hills, Amet Center, Sit Quarter
- **4 Users**: John Doe (Admin), Jane Smith (Publisher), Bob Johnson (Admin), Alice Williams (Publisher)
- **6 Territories**: Various streets and locations (available and in-use)
- **3 Territory Notes** with sample activity

View with: `make inspect`

## Database Connection

**Local:**
- Host: localhost
- Port: 5432
- Database: api
- User: admin
- Password: postgres

**Supabase:**
- Use Connection Pooler settings (port 6543, not 5432)
- Find in: Settings → Database → Connection Pooler

## Environment Variables

Required variables:
```bash
TS_LOG_LEVEL              # debug, info, error
TS_POSTGRESQL_USER        # Database user
TS_POSTGRESQL_PASSWORD    # Database password
TS_POSTGRESQL_HOST        # Database host
TS_POSTGRESQL_DATABASE    # Database name
TS_TELEGRAM_BOT_TOKEN     # Bot token from @BotFather
```

## Troubleshooting

**Database won't start:**
```bash
lsof -i :5432          # Check if port is in use
make logs              # View database logs
make docker-clean      # Nuclear option
```

**Bot not responding:**
```bash
make cloud-logs        # Check Cloud Run logs
# Verify bot token is correct
# Verify database connection (use pooler port 6543 for Supabase)
```

**Need fresh start:**
```bash
make docker-clean
make docker-up
make seed
```
