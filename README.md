# AI Cycling Coach

A single-user, self-hosted web application that provides AI-powered cycling training plan generation, workout analysis, and plan evolution based on actual ride data from Garmin Connect.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- 2GB+ available RAM
- 10GB+ available disk space

### Setup
1. Clone the repository
2. Copy environment file: `cp .env.example .env`
3. Edit `.env` with your credentials
4. Start services: `docker-compose up -d`

## 🐳 Container-First Development

This project follows strict containerization practices. All development occurs within Docker containers - never install packages directly on the host system.

### Key Rules

#### Containerization Rules
- ✅ All Python packages must be in `backend/requirements.txt`
- ✅ All system packages must be in `backend/Dockerfile`
- ✅ Never run `pip install` or `apt-get install` outside containers
- ✅ Use `docker-compose` for local development

#### Database Management
- ✅ Schema changes handled through Alembic migrations
- ✅ Migrations run automatically on container startup
- ✅ No raw SQL in application code - use SQLAlchemy ORM
- ✅ Migration rollback scripts available for emergencies

### Development Workflow

```bash
# Start development environment
docker-compose up -d

# View logs
docker-compose logs -f backend

# Run database migrations manually (if needed)
docker-compose exec backend alembic upgrade head

# Access backend container
docker-compose exec backend bash

# Stop services
docker-compose down
```

### Migration Management

#### Automatic Migrations
Migrations run automatically when containers start. The entrypoint script:
1. Runs `alembic upgrade head`
2. Verifies migration success
3. Starts the application

#### Manual Migration Operations
```bash
# Check migration status
docker-compose exec backend python scripts/migration_checker.py check-db

# Generate new migration
docker-compose exec backend alembic revision --autogenerate -m "description"

# Rollback migration
docker-compose exec backend python scripts/migration_rollback.py rollback
```

#### Migration Validation
```bash
# Validate deployment readiness
docker-compose exec backend python scripts/migration_checker.py validate-deploy

# Generate migration report
docker-compose exec backend python scripts/migration_checker.py report
```

### Database Backup & Restore

#### Creating Backups
```bash
# Create backup
docker-compose exec backend python scripts/backup_restore.py backup

# Create named backup
docker-compose exec backend python scripts/backup_restore.py backup my_backup
```

#### Restoring from Backup
```bash
# List available backups
docker-compose exec backend python scripts/backup_restore.py list

# Restore (with confirmation prompt)
docker-compose exec backend python scripts/backup_restore.py restore backup_file.sql

# Restore without confirmation
docker-compose exec backend python scripts/backup_restore.py restore backup_file.sql --yes
```

#### Cleanup
```bash
# Remove backups older than 30 days
docker-compose exec backend python scripts/backup_restore.py cleanup

# Remove backups older than N days
docker-compose exec backend python scripts/backup_restore.py cleanup 7
```

## 🔧 Configuration

### Environment Variables
```env
# Database
DATABASE_URL=postgresql://postgres:password@db:5432/cycling

# Garmin Connect
GARMIN_USERNAME=your_garmin_email@example.com
GARMIN_PASSWORD=your_secure_password

# AI Service
OPENROUTER_API_KEY=your_openrouter_api_key
AI_MODEL=anthropic/claude-3-sonnet-20240229

# Application
API_KEY=your_secure_random_api_key_here
```

### Health Checks

The application includes comprehensive health monitoring:

```bash
# Check overall health
curl http://localhost:8000/health

# Response includes:
# - Database connectivity
# - Migration status
# - Current vs head revision
# - Service availability
```

## 🏗️ Architecture

### Service Architecture
```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │
│   (React)       │◄──►│   (FastAPI)     │
│                 │    │                 │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   Garmin        │    │   PostgreSQL    │
│   Connect       │    │   Database      │
└─────────────────┘    └─────────────────┘
```

### Data Flow
1. Garmin activities synced via background tasks
2. AI analysis performed on workout data
3. Training plans evolved based on performance
4. User feedback incorporated for plan adjustments

## 🧪 Testing & Validation

### CI/CD Pipeline
GitHub Actions automatically validates:
- ✅ No uncommitted migration files
- ✅ No raw SQL in application code
- ✅ Proper dependency management
- ✅ Container build success
- ✅ Migration compatibility

### Local Validation
```bash
# Run all validation checks
docker-compose exec backend python scripts/migration_checker.py validate-deploy

# Check for raw SQL usage
grep -r "SELECT.*FROM\|INSERT.*INTO\|UPDATE.*SET\|DELETE.*FROM" backend/app/
```

## 📁 Project Structure

```
.
├── backend/
│   ├── Dockerfile              # Multi-stage container build
│   ├── requirements.txt        # Python dependencies
│   ├── scripts/
│   │   ├── migration_rollback.py   # Rollback utilities
│   │   ├── backup_restore.py      # Backup/restore tools
│   │   └── migration_checker.py   # Validation tools
│   └── app/
│       ├── main.py             # FastAPI application
│       ├── database.py         # Database configuration
│       ├── models/             # SQLAlchemy models
│       ├── routes/             # API endpoints
│       ├── services/           # Business logic
│       └── schemas/            # Pydantic schemas
├── frontend/
│   ├── Dockerfile
│   └── src/
├── docker-compose.yml          # Development services
├── .github/
│   └── workflows/
│       └── container-validation.yml  # CI/CD checks
└── .kilocode/
    └── rules/
        └── container-database-rules.md  # Development guidelines
```

## 🚨 Troubleshooting

### Common Issues

#### Migration Failures
```bash
# Check migration status
docker-compose exec backend alembic current

# View migration history
docker-compose exec backend alembic history

# Reset migrations (CAUTION: destroys data)
docker-compose exec backend alembic downgrade base
```

#### Database Connection Issues
```bash
# Check database health
docker-compose exec db pg_isready -U postgres

# View database logs
docker-compose logs db

# Restart database
docker-compose restart db
```

#### Container Build Issues
```bash
# Rebuild without cache
docker-compose build --no-cache backend

# View build logs
docker-compose build backend
```

### Health Monitoring

#### Service Health
```bash
# Check all services
docker-compose ps

# View service logs
docker-compose logs -f

# Check backend health
curl http://localhost:8000/health
```

#### Database Health
```bash
# Check database connectivity
docker-compose exec backend python -c "
from app.database import get_db
from sqlalchemy.ext.asyncio import AsyncSession
import asyncio

async def test():
    async with AsyncSession(get_db()) as session:
        result = await session.execute('SELECT 1')
        print('Database OK')

asyncio.run(test())
"
```

## 🔒 Security

- API key authentication for all endpoints
- Secure storage of Garmin credentials
- No sensitive data in application logs
- Container isolation prevents host system access
- Regular security updates via container rebuilds

## 📚 API Documentation

Once running, visit:
- **API Docs**: http://localhost:8000/docs
- **Alternative Docs**: http://localhost:8000/redoc

## 🤝 Contributing

1. Follow container-first development rules
2. Ensure all changes pass CI/CD validation
3. Update documentation for significant changes
4. Test migration compatibility before merging

### Development Guidelines

- Use SQLAlchemy ORM for all database operations
- Keep dependencies in `requirements.txt`
- Test schema changes in development environment
- Document migration changes in commit messages
- Run validation checks before pushing

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Note**: This application is designed for single-user, self-hosted deployment. All data remains on your local infrastructure with no external data sharing.