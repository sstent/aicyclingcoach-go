from fastapi import APIRouter
from app.services.health_monitor import HealthMonitor

router = APIRouter()
monitor = HealthMonitor()

@router.get("/health")
async def get_health():
    return monitor.check_system_health()