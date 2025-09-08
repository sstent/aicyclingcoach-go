import psutil
from datetime import datetime
import logging
from typing import Dict, Any
from sqlalchemy import text
from app.database import get_db
from app.models.garmin_sync_log import GarminSyncLog, SyncStatus
import requests
from app.config import settings

logger = logging.getLogger(__name__)

class HealthMonitor:
    def __init__(self):
        self.warning_thresholds = {
            'cpu_percent': 80,
            'memory_percent': 75,
            'disk_percent': 85
        }

    def check_system_health(self) -> Dict[str, Any]:
        """Check vital system metrics and log warnings"""
        metrics = {
            'timestamp': datetime.utcnow().isoformat(),
            'cpu': psutil.cpu_percent(),
            'memory': psutil.virtual_memory().percent,
            'disk': psutil.disk_usage('/').percent,
            'services': self._check_service_health()
        }
        
        self._log_anomalies(metrics)
        return metrics

    def _check_service_health(self) -> Dict[str, str]:
        """Check critical application services"""
        return {
            'database': self._check_database(),
            'garmin_sync': self._check_garmin_sync(),
            'ai_service': self._check_ai_service()
        }

    def _check_database(self) -> str:
        try:
            with get_db() as db:
                db.execute(text("SELECT 1"))
            return "ok"
        except Exception as e:
            logger.error(f"Database check failed: {str(e)}")
            return "down"

    def _check_garmin_sync(self) -> str:
        try:
            last_sync = GarminSyncLog.get_latest()
            if last_sync and last_sync.status == SyncStatus.FAILED:
                return "warning"
            return "ok"
        except Exception as e:
            logger.error(f"Garmin sync check failed: {str(e)}")
            return "down"

    def _check_ai_service(self) -> str:
        try:
            response = requests.get(
                f"{settings.AI_SERVICE_URL}/ping", 
                timeout=5,
                headers={"Authorization": f"Bearer {settings.OPENROUTER_API_KEY}"}
            )
            return "ok" if response.ok else "down"
        except Exception as e:
            logger.error(f"AI service check failed: {str(e)}")
            return "down"

    def _log_anomalies(self, metrics: Dict[str, Any]):
        alerts = []
        for metric, value in metrics.items():
            if metric in self.warning_thresholds and value > self.warning_thresholds[metric]:
                alerts.append(f"{metric} {value}%")
        
        if alerts:
            logger.warning(f"System thresholds exceeded: {', '.join(alerts)}")