from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.database import get_db
from app.models import Plan, PlanRule, Rule
from app.schemas.plan import PlanCreate, Plan as PlanSchema
from uuid import UUID

router = APIRouter(prefix="/plans", tags=["Training Plans"])

@router.post("/", response_model=PlanSchema)
async def create_plan(
    plan: PlanCreate,
    db: AsyncSession = Depends(get_db)
):
    # Create plan
    db_plan = Plan(
        user_id=plan.user_id,
        start_date=plan.start_date,
        end_date=plan.end_date,
        goal=plan.goal
    )
    db.add(db_plan)
    await db.flush()  # Flush to get plan ID
    
    # Add rules to plan
    for rule_id in plan.rule_ids:
        db_plan_rule = PlanRule(plan_id=db_plan.id, rule_id=rule_id)
        db.add(db_plan_rule)
    
    await db.commit()
    await db.refresh(db_plan)
    return db_plan

@router.get("/{plan_id}", response_model=PlanSchema)
async def read_plan(
    plan_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    plan = await db.get(Plan, plan_id)
    if not plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    return plan

@router.get("/", response_model=list[PlanSchema])
async def read_plans(
    db: AsyncSession = Depends(get_db)
):
    result = await db.execute(select(Plan))
    return result.scalars().all()

@router.put("/{plan_id}", response_model=PlanSchema)
async def update_plan(
    plan_id: UUID,
    plan: PlanCreate,
    db: AsyncSession = Depends(get_db)
):
    db_plan = await db.get(Plan, plan_id)
    if not db_plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    
    # Update plan fields
    db_plan.user_id = plan.user_id
    db_plan.start_date = plan.start_date
    db_plan.end_date = plan.end_date
    db_plan.goal = plan.goal
    
    # Update rules
    await db.execute(PlanRule.delete().where(PlanRule.plan_id == plan_id))
    for rule_id in plan.rule_ids:
        db_plan_rule = PlanRule(plan_id=plan_id, rule_id=rule_id)
        db.add(db_plan_rule)
    
    await db.commit()
    await db.refresh(db_plan)
    return db_plan

@router.delete("/{plan_id}")
async def delete_plan(
    plan_id: UUID,
    db: AsyncSession = Depends(get_db)
):
    plan = await db.get(Plan, plan_id)
    if not plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    
    await db.delete(plan)
    await db.commit()
    return {"detail": "Plan deleted"}