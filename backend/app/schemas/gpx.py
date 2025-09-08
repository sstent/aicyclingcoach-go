from pydantic import BaseModel
from typing import Optional, List

class GPXData(BaseModel):
    total_distance: float
    elevation_gain: float
    points: List[dict]

class RouteCreate(BaseModel):
    name: str
    description: Optional[str] = None
    total_distance: float
    elevation_gain: float
    gpx_file_path: str

class Route(BaseModel):
    id: str
    name: str
    description: Optional[str] = None
    total_distance: float
    elevation_gain: float
    gpx_file_path: str

    class Config:
        orm_mode = True