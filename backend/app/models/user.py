from .base import BaseModel
from sqlalchemy.orm import relationship

class User(BaseModel):
    __tablename__ = "users"
    
    plans = relationship("Plan", back_populates="user")