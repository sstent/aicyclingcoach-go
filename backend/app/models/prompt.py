from sqlalchemy import Column, Integer, String, Text, Boolean, DateTime
from .base import BaseModel


class Prompt(BaseModel):
    """Prompt model for AI prompt versioning and management."""
    __tablename__ = "prompts"

    action_type = Column(String(50), nullable=False)  # plan_generation, workout_analysis, rule_parsing, suggestions
    model = Column(String(100))  # AI model identifier
    prompt_text = Column(Text, nullable=False)
    version = Column(Integer, default=1)
    active = Column(Boolean, default=True)