from sqlalchemy import Column, Integer, DateTime, String, Text
from .base import BaseModel


class GarminSyncLog(BaseModel):
    """Log model for tracking Garmin sync operations."""
    __tablename__ = "garmin_sync_log"

    last_sync_time = Column(DateTime)
    activities_synced = Column(Integer, default=0)
    status = Column(String(20))  # success, error, in_progress
    error_message = Column(Text)