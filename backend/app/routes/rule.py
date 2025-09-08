from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from app.database import get_db
from app.models import Rule
from app.schemas.rule import RuleCreate, Rule as RuleSchema
from uuid import UUID

router = APIRouter(prefix="/rules", tags=["Rules"])

@router.post("/", response_model=RuleSchema)
async def create_rule(
    rule: RuleCreate,
    db: AsyncSession = Depends(get_db)
):
    db_rule = Rule(**rule.dict())
    db.add(db_rule)
    await db.commit()
    await db.refresh(db_rule)
    return db_rule

@router.get("/{rule_id}", response_model=RuleSchema)
async def read_rule(
    rule_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    rule = await db.get(Rule, rule_id)
    if not rule:
        raise HTTPException(status_code=404, detail="Rule not found")
    return rule

@router.get("/", response_model=list[RuleSchema])
async def read_rules(
    db: AsyncSession = Depends(get_db)
):
    result = await db.execute(sa.select(Rule))
    return result.scalars().all()

@router.put("/{rule_id}", response_model=RuleSchema)
async def update_rule(
    rule_id: UUID,
    rule: RuleCreate,
    db: AsyncSession = Depends(get_db)
):
    db_rule = await db.get(Rule, rule_id)
    if not db_rule:
        raise HTTPException(status_code=404, detail="Rule not found")
    
    for key, value in rule.dict().items():
        setattr(db_rule, key, value)
    
    await db.commit()
    await db.refresh(db_rule)
    return db_rule

@router.delete("/{rule_id}")
async def delete_rule(
    rule_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    rule = await db.get(Rule, rule_id)
    if not rule:
        raise HTTPException(status_code=404, detail="Rule not found")
    
    await db.delete(rule)
    await db.commit()
    return {"detail": "Rule deleted"}