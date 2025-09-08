from pydantic import BaseModel
from typing import Optional

class RuleBase(BaseModel):
    name: str
    description: Optional[str] = None
    condition: str
    priority: int = 0

class RuleCreate(RuleBase):
    pass

class Rule(RuleBase):
    id: str

    class Config:
        orm_mode = True