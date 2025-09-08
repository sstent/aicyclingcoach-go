from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, JSON, Boolean
from sqlalchemy.orm import relationship
from .base import BaseModel


class Workout(BaseModel):
    """Workout model for Garmin activity data."""
    __tablename__ = "workouts"

    plan_id = Column(Integer, ForeignKey("plans.id"), nullable=True)
    garmin_activity_id = Column(String(255), unique=True, nullable=False)
    activity_type = Column(String(50))
    start_time = Column(DateTime, nullable=False)
    duration_seconds = Column(Integer)
    distance_m = Column(Float)
    avg_hr = Column(Integer)
    max_hr = Column(Integer)
    avg_power = Column(Float)
    max_power = Column(Float)
    avg_cadence = Column(Float)
    elevation_gain_m = Column(Float)
    metrics = Column(JSON)  # Store full Garmin data as JSONB

    # Relationships
    plan = relationship("Plan", back_populates="workouts")
    analyses = relationship("Analysis", back_populates="workout", cascade="all, delete-orphan")