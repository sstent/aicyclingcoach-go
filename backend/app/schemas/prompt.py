from pydantic import BaseModel
from typing import Optional
from datetime import datetime


class PromptBase(BaseModel):
    action_type: str
    model: Optional[str] = None
    prompt_text: str
    version: int = 1
    active: bool = True


class PromptCreate(BaseModel):
    action_type: str
    prompt_text: str
    model: Optional[str] = None


class PromptUpdate(BaseModel):
    prompt_text: Optional[str] = None
    active: Optional[bool] = None


class Prompt(PromptBase):
    id: int
    created_at: datetime

    class Config:
        orm_mode = True