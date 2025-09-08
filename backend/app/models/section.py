from sqlalchemy import Column, String, Float, ForeignKey
from sqlalchemy.orm import relationship
from .base import BaseModel

class Section(BaseModel):
    __tablename__ = "sections"
    
    route_id = Column(ForeignKey("routes.id"), nullable=False)
    gpx_file_path = Column(String(255), nullable=False)
    distance_m = Column(Float, nullable=False)
    grade_avg = Column(Float)
    min_gear = Column(String(50))
    est_time_minutes = Column(Float)
    
    route = relationship("Route", back_populates="sections")