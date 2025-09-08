from fastapi import FastAPI, Depends, Request, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from .database import get_db, get_database_url
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from alembic.config import Config
from alembic.migration import MigrationContext
from alembic.script import ScriptDirectory
from .routes import gpx as gpx_routes
from .routes import rule as rule_routes
from .routes import plan as plan_routes
from .routes import workouts as workout_routes
from .routes import prompts as prompt_routes
from .config import settings

app = FastAPI(
    title="AI Cycling Coach API",
    description="Backend service for AI-assisted cycling training platform",
    version="0.1.0"
)

# API Key Authentication Middleware
@app.middleware("http")
async def api_key_auth(request: Request, call_next):
    if request.url.path.startswith("/docs") or request.url.path.startswith("/redoc") or request.url.path == "/health":
        return await call_next(request)
        
    api_key = request.headers.get("X-API-KEY")
    if api_key != settings.API_KEY:
        raise HTTPException(status_code=401, detail="Invalid API Key")
    
    return await call_next(request)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(gpx_routes.router)
app.include_router(rule_routes.router)
app.include_router(plan_routes.router)
app.include_router(workout_routes.router, prefix="/workouts", tags=["workouts"])
app.include_router(prompt_routes.router, prefix="/prompts", tags=["prompts"])

async def check_migration_status():
    """Check if database migrations are up to date."""
    try:
        # Get Alembic configuration
        config = Config("alembic.ini")
        config.set_main_option("sqlalchemy.url", get_database_url())
        script = ScriptDirectory.from_config(config)

        # Get current database revision
        from sqlalchemy import create_engine
        engine = create_engine(get_database_url())
        with engine.connect() as conn:
            context = MigrationContext.configure(conn)
            current_rev = context.get_current_revision()

        # Get head revision
        head_rev = script.get_current_head()

        return {
            "current_revision": current_rev,
            "head_revision": head_rev,
            "migrations_up_to_date": current_rev == head_rev
        }
    except Exception as e:
        return {
            "error": str(e),
            "migrations_up_to_date": False
        }

@app.get("/health")
async def health_check(db: AsyncSession = Depends(get_db)):
    """Enhanced health check with migration verification."""
    health_status = {
        "status": "healthy",
        "version": "0.1.0",
        "timestamp": "2024-01-15T10:30:00Z"  # Should be dynamic
    }

    # Database connection check
    try:
        await db.execute(text("SELECT 1"))
        health_status["database"] = "connected"
    except Exception as e:
        health_status["status"] = "unhealthy"
        health_status["database"] = f"error: {str(e)}"

    # Migration status check
    migration_info = await check_migration_status()
    health_status["migrations"] = migration_info

    if not migration_info.get("migrations_up_to_date", False):
        health_status["status"] = "unhealthy"

    return health_status

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)