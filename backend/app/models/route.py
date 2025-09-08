from sqlalchemy import Column, String, Float, ForeignKey
from sqlalchemy.orm import relationship
from .base import BaseModel

class Route(BaseModel):
    __tablename__ = "routes"
    
    name = Column(String(100), nullable=False)
    description = Column(String(500))
    total_distance = Column(Float, nullable=False)
    elevation_gain = Column(Float, nullable=False)
    gpx_file_path = Column(String(255), nullable=False)
    
    sections = relationship("Section", back_populates="route", cascade="all, delete-orphan")