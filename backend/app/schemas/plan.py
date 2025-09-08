from pydantic import BaseModel
from datetime import datetime
from typing import List, Optional
from uuid import UUID

class PlanBase(BaseModel):
    user_id: UUID
    start_date: datetime
    end_date: datetime
    goal: str

class PlanCreate(PlanBase):
    rule_ids: List[UUID]

class Plan(PlanBase):
    id: UUID

    class Config:
        orm_mode = True