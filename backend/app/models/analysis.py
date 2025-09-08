from sqlalchemy import Column, Integer, String, ForeignKey, JSON, Boolean, DateTime
from sqlalchemy.orm import relationship
from .base import BaseModel


class Analysis(BaseModel):
    """Analysis model for AI-generated workout feedback."""
    __tablename__ = "analyses"

    workout_id = Column(Integer, ForeignKey("workouts.id"), nullable=False)
    analysis_type = Column(String(50), default='workout_review')
    jsonb_feedback = Column(JSON)  # AI-generated feedback
    suggestions = Column(JSON)  # AI-generated suggestions
    approved = Column(Boolean, default=False)

    # Relationships
    workout = relationship("Workout", back_populates="analyses")