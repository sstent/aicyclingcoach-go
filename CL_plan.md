# AI-Assisted Cycling Coach - Updated Implementation Plan

## Overview
This document outlines the implementation plan for a single-user, self-hosted AI-assisted cycling coach application with Python backend, PostgreSQL database, GPX file storage, and web frontend.

## Architecture Components
- **Backend**: Python FastAPI (async)
- **Database**: PostgreSQL with versioning support
- **File Storage**: Local directory for GPX files (up to 200 files)
- **Frontend**: React/Next.js
- **AI Integration**: OpenRouter API
- **Garmin Integration**: garth or garmin-connect Python modules
- **Authentication**: Simple API key for single-user setup
- **Containerization**: Docker + Docker Compose (self-hosted)

## Implementation Phases

### Phase 1: Project Setup and Foundation ✅ (Week 1-2)
**Status: Complete**

1. **Initialize Project Structure**
   ```
   /
   ├── backend/
   │   ├── app/
   │   │   ├── __init__.py
   │   │   ├── main.py
   │   │   ├── models/
   │   │   ├── routes/
   │   │   ├── services/
   │   │   └── utils/
   │   ├── requirements.txt
   │   └── Dockerfile
   ├── frontend/
   │   ├── src/
   │   ├── public/
   │   ├── package.json
   │   └── Dockerfile
   ├── docker-compose.yml
   ├── .env.example
   └── README.md
   ```

2. **Docker Environment Setup**
   ```yaml
   version: '3.9'
   services:
     backend:
       build: ./backend
       ports:
         - "8000:8000"
       volumes:
         - gpx-data:/app/data/gpx
         - garmin-sessions:/app/data/sessions
       environment:
         - DATABASE_URL=postgresql://postgres:password@db:5432/cycling
         - GARMIN_USERNAME=${GARMIN_USERNAME}
         - GARMIN_PASSWORD=${GARMIN_PASSWORD}
         - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
         - AI_MODEL=${AI_MODEL:-claude-3-sonnet-20240229}
         - API_KEY=${API_KEY}
       depends_on:
         - db

     frontend:
       build: ./frontend
       ports:
         - "3000:3000"
       environment:
         - REACT_APP_API_URL=http://localhost:8000
         - REACT_APP_API_KEY=${API_KEY}

     db:
       image: postgres:15
       restart: always
       environment:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password
         POSTGRES_DB: cycling
       volumes:
         - postgres-data:/var/lib/postgresql/data

   volumes:
     gpx-data:
       driver: local
     garmin-sessions:
       driver: local
     postgres-data:
       driver: local
   ```

3. **Database Schema with Enhanced Versioning**
   ```sql
   -- Routes & Sections
   CREATE TABLE routes (
       id SERIAL PRIMARY KEY,
       name TEXT NOT NULL,
       created_at TIMESTAMP DEFAULT now()
   );

   CREATE TABLE sections (
       id SERIAL PRIMARY KEY,
       route_id INT REFERENCES routes(id),
       gpx_file_path TEXT NOT NULL,
       distance_m NUMERIC,
       grade_avg NUMERIC,
       min_gear TEXT,
       est_time_minutes NUMERIC,
       created_at TIMESTAMP DEFAULT now()
   );

   -- Rules with versioning and evolution tracking
   CREATE TABLE rules (
       id SERIAL PRIMARY KEY,
       name TEXT NOT NULL,
       user_defined BOOLEAN DEFAULT true,
       jsonb_rules JSONB NOT NULL,
       version INT DEFAULT 1,
       parent_rule_id INT REFERENCES rules(id),
       created_at TIMESTAMP DEFAULT now()
   );

   -- Plans with versioning and evolution tracking
   CREATE TABLE plans (
       id SERIAL PRIMARY KEY,
       jsonb_plan JSONB NOT NULL,
       version INT NOT NULL,
       parent_plan_id INT REFERENCES plans(id),
       created_at TIMESTAMP DEFAULT now()
   );

   -- Workouts with Garmin integration
   CREATE TABLE workouts (
       id SERIAL PRIMARY KEY,
       plan_id INT REFERENCES plans(id),
       garmin_activity_id TEXT UNIQUE NOT NULL,
       activity_type TEXT,
       start_time TIMESTAMP,
       duration_seconds INT,
       distance_m NUMERIC,
       avg_hr INT,
       max_hr INT,
       avg_power NUMERIC,
       max_power NUMERIC,
       avg_cadence NUMERIC,
       elevation_gain_m NUMERIC,
       metrics JSONB, -- Additional Garmin data
       created_at TIMESTAMP DEFAULT now()
   );

   -- Analyses with enhanced feedback structure
   CREATE TABLE analyses (
       id SERIAL PRIMARY KEY,
       workout_id INT REFERENCES workouts(id),
       analysis_type TEXT DEFAULT 'workout_review',
       jsonb_feedback JSONB,
       suggestions JSONB,
       approved BOOLEAN DEFAULT false,
       created_at TIMESTAMP DEFAULT now()
   );

   -- AI Prompts with versioning
   CREATE TABLE prompts (
       id SERIAL PRIMARY KEY,
       action_type TEXT, -- plan_generation, workout_analysis, rule_parsing, suggestions
       model TEXT,
       prompt_text TEXT,
       version INT DEFAULT 1,
       active BOOLEAN DEFAULT true,
       created_at TIMESTAMP DEFAULT now()
   );

   -- Garmin sync status tracking
   CREATE TABLE garmin_sync_log (
       id SERIAL PRIMARY KEY,
       last_sync_time TIMESTAMP,
       activities_synced INT DEFAULT 0,
       status TEXT, -- success, error, in_progress
       error_message TEXT,
       created_at TIMESTAMP DEFAULT now()
   );
   ```

### Phase 2: Core Backend Implementation ✅ (Week 3-5)
**Status: Complete**

1. **Database Models with SQLAlchemy**
2. **Basic API Endpoints**
3. **GPX File Handling**
4. **Basic Authentication Middleware**

### Phase 3: Enhanced Backend + Garmin Integration (Week 6-8)

#### Week 6: Garmin Integration
1. **Garmin Service Implementation**
   ```python
   # backend/app/services/garmin.py
   import os
   import garth
   from typing import List, Dict, Any, Optional
   from datetime import datetime, timedelta

   class GarminService:
       def __init__(self):
           self.username = os.getenv("GARMIN_USERNAME")
           self.password = os.getenv("GARMIN_PASSWORD")
           self.client: Optional[garth.Client] = None
           self.session_dir = "/app/data/sessions"

       async def authenticate(self):
           """Authenticate with Garmin Connect and persist session."""
           if not self.client:
               self.client = garth.Client()
               
           try:
               # Try to load existing session
               self.client.load(self.session_dir)
           except Exception:
               # Fresh authentication
               await self.client.login(self.username, self.password)
               self.client.save(self.session_dir)

       async def get_activities(self, limit: int = 10, start_date: datetime = None) -> List[Dict[str, Any]]:
           """Fetch recent activities from Garmin Connect."""
           if not self.client:
               await self.authenticate()
           
           if not start_date:
               start_date = datetime.now() - timedelta(days=7)
           
           activities = self.client.get_activities(limit=limit, start=start_date)
           return activities

       async def get_activity_details(self, activity_id: str) -> Dict[str, Any]:
           """Get detailed activity data including metrics."""
           if not self.client:
               await self.authenticate()
           
           details = self.client.get_activity(activity_id)
           return details
   ```

2. **Workout Sync Service**
   ```python
   # backend/app/services/workout_sync.py
   from sqlalchemy.ext.asyncio import AsyncSession
   from app.services.garmin import GarminService
   from app.models.workout import Workout
   from app.models.garmin_sync_log import GarminSyncLog

   class WorkoutSyncService:
       def __init__(self, db: AsyncSession):
           self.db = db
           self.garmin_service = GarminService()

       async def sync_recent_activities(self, days_back: int = 7):
           """Sync recent Garmin activities to database."""
           try:
               sync_log = GarminSyncLog(status="in_progress")
               self.db.add(sync_log)
               await self.db.commit()

               start_date = datetime.now() - timedelta(days=days_back)
               activities = await self.garmin_service.get_activities(
                   limit=50, start_date=start_date
               )

               synced_count = 0
               for activity in activities:
                   if await self.activity_exists(activity['activityId']):
                       continue

                   workout_data = await self.parse_activity_data(activity)
                   workout = Workout(**workout_data)
                   self.db.add(workout)
                   synced_count += 1

               sync_log.status = "success"
               sync_log.activities_synced = synced_count
               sync_log.last_sync_time = datetime.now()
               
               await self.db.commit()
               return synced_count

           except Exception as e:
               sync_log.status = "error"
               sync_log.error_message = str(e)
               await self.db.commit()
               raise

       async def activity_exists(self, garmin_activity_id: str) -> bool:
           """Check if activity already exists in database."""
           result = await self.db.execute(
               select(Workout).where(Workout.garmin_activity_id == garmin_activity_id)
           )
           return result.scalar_one_or_none() is not None

       async def parse_activity_data(self, activity: Dict[str, Any]) -> Dict[str, Any]:
           """Parse Garmin activity data into workout model format."""
           return {
               "garmin_activity_id": activity['activityId'],
               "activity_type": activity.get('activityType', {}).get('typeKey'),
               "start_time": datetime.fromisoformat(activity['startTimeLocal'].replace('Z', '+00:00')),
               "duration_seconds": activity.get('duration'),
               "distance_m": activity.get('distance'),
               "avg_hr": activity.get('averageHR'),
               "max_hr": activity.get('maxHR'),
               "avg_power": activity.get('avgPower'),
               "max_power": activity.get('maxPower'),
               "avg_cadence": activity.get('averageBikingCadenceInRevPerMinute'),
               "elevation_gain_m": activity.get('elevationGain'),
               "metrics": activity  # Store full Garmin data as JSONB
           }
   ```

3. **Background Tasks Setup**
   ```python
   # backend/app/main.py
   from fastapi import BackgroundTasks
   from app.services.workout_sync import WorkoutSyncService

   @app.post("/api/workouts/sync")
   async def trigger_garmin_sync(
       background_tasks: BackgroundTasks,
       db: AsyncSession = Depends(get_db)
   ):
       """Trigger background sync of recent Garmin activities."""
       sync_service = WorkoutSyncService(db)
       background_tasks.add_task(sync_service.sync_recent_activities, days_back=14)
       return {"message": "Garmin sync started"}

   @app.get("/api/workouts/sync-status")
   async def get_sync_status(db: AsyncSession = Depends(get_db)):
       """Get the latest sync status."""
       result = await db.execute(
           select(GarminSyncLog).order_by(GarminSyncLog.created_at.desc()).limit(1)
       )
       sync_log = result.scalar_one_or_none()
       return sync_log
   ```

#### Week 7: Enhanced AI Integration
1. **Prompt Management System**
   ```python
   # backend/app/services/prompt_manager.py
   class PromptManager:
       def __init__(self, db: AsyncSession):
           self.db = db

       async def get_active_prompt(self, action_type: str, model: str = None) -> Optional[str]:
           """Get the active prompt for a specific action type."""
           query = select(Prompt).where(
               Prompt.action_type == action_type,
               Prompt.active == True
           )
           if model:
               query = query.where(Prompt.model == model)
           
           result = await self.db.execute(query.order_by(Prompt.version.desc()))
           prompt = result.scalar_one_or_none()
           return prompt.prompt_text if prompt else None

       async def create_prompt_version(
           self, 
           action_type: str, 
           prompt_text: str, 
           model: str = None
       ) -> Prompt:
           """Create a new version of a prompt."""
           # Deactivate previous versions
           await self.db.execute(
               update(Prompt)
               .where(Prompt.action_type == action_type)
               .values(active=False)
           )
           
           # Get next version number
           result = await self.db.execute(
               select(func.max(Prompt.version))
               .where(Prompt.action_type == action_type)
           )
           max_version = result.scalar() or 0
           
           # Create new prompt
           new_prompt = Prompt(
               action_type=action_type,
               model=model,
               prompt_text=prompt_text,
               version=max_version + 1,
               active=True
           )
           
           self.db.add(new_prompt)
           await self.db.commit()
           return new_prompt
   ```

2. **Enhanced AI Service**
   ```python
   # backend/app/services/ai_service.py
   import asyncio
   from typing import Dict, Any, List
   import httpx
   from app.services.prompt_manager import PromptManager

   class AIService:
       def __init__(self, db: AsyncSession):
           self.db = db
           self.prompt_manager = PromptManager(db)
           self.api_key = os.getenv("OPENROUTER_API_KEY")
           self.model = os.getenv("AI_MODEL", "anthropic/claude-3-sonnet-20240229")
           self.base_url = "https://openrouter.ai/api/v1"

       async def analyze_workout(self, workout: Workout, plan: Optional[Dict] = None) -> Dict[str, Any]:
           """Analyze a workout using AI and generate feedback."""
           prompt_template = await self.prompt_manager.get_active_prompt("workout_analysis")
           
           if not prompt_template:
               raise ValueError("No active workout analysis prompt found")

           # Build context from workout data
           workout_context = {
               "activity_type": workout.activity_type,
               "duration_minutes": workout.duration_seconds / 60 if workout.duration_seconds else 0,
               "distance_km": workout.distance_m / 1000 if workout.distance_m else 0,
               "avg_hr": workout.avg_hr,
               "avg_power": workout.avg_power,
               "elevation_gain": workout.elevation_gain_m,
               "planned_workout": plan
           }

           prompt = prompt_template.format(**workout_context)
           
           response = await self._make_ai_request(prompt)
           return self._parse_workout_analysis(response)

       async def generate_plan(self, rules: List[Dict], goals: Dict[str, Any]) -> Dict[str, Any]:
           """Generate a training plan using AI."""
           prompt_template = await self.prompt_manager.get_active_prompt("plan_generation")
           
           context = {
               "rules": rules,
               "goals": goals,
               "current_fitness_level": goals.get("fitness_level", "intermediate")
           }
           
           prompt = prompt_template.format(**context)
           response = await self._make_ai_request(prompt)
           return self._parse_plan_response(response)

       async def parse_rules_from_natural_language(self, natural_language: str) -> Dict[str, Any]:
           """Parse natural language rules into structured format."""
           prompt_template = await self.prompt_manager.get_active_prompt("rule_parsing")
           prompt = prompt_template.format(user_rules=natural_language)
           
           response = await self._make_ai_request(prompt)
           return self._parse_rules_response(response)

       async def _make_ai_request(self, prompt: str) -> str:
           """Make async request to OpenRouter API with retry logic."""
           async with httpx.AsyncClient() as client:
               for attempt in range(3):  # Simple retry logic
                   try:
                       response = await client.post(
                           f"{self.base_url}/chat/completions",
                           headers={
                               "Authorization": f"Bearer {self.api_key}",
                               "Content-Type": "application/json",
                           },
                           json={
                               "model": self.model,
                               "messages": [{"role": "user", "content": prompt}],
                               "max_tokens": 2000,
                           },
                           timeout=30.0
                       )
                       response.raise_for_status()
                       data = response.json()
                       return data["choices"][0]["message"]["content"]
                       
                   except Exception as e:
                       if attempt == 2:  # Last attempt
                           raise AIServiceError(f"AI request failed after 3 attempts: {str(e)}")
                       await asyncio.sleep(2 ** attempt)  # Exponential backoff

       def _parse_workout_analysis(self, response: str) -> Dict[str, Any]:
           """Parse AI response for workout analysis."""
           # Implementation depends on your prompt design
           # This is a simplified example
           try:
               import json
               # Assume AI returns JSON
               clean_response = response.strip()
               if clean_response.startswith("```json"):
                   clean_response = clean_response[7:-3]
               return json.loads(clean_response)
           except json.JSONDecodeError:
               return {"raw_analysis": response, "structured": False}
   ```

#### Week 8: Plan Evolution & Analysis Pipeline
1. **Plan Evolution Service**
   ```python
   # backend/app/services/plan_evolution.py
   class PlanEvolutionService:
       def __init__(self, db: AsyncSession):
           self.db = db
           self.ai_service = AIService(db)

       async def evolve_plan_from_analysis(
           self, 
           analysis: Analysis, 
           current_plan: Plan
       ) -> Optional[Plan]:
           """Create a new plan version based on workout analysis."""
           if not analysis.approved:
               return None

           suggestions = analysis.suggestions
           if not suggestions:
               return None

           # Generate new plan incorporating suggestions
           evolution_context = {
               "current_plan": current_plan.jsonb_plan,
               "workout_analysis": analysis.jsonb_feedback,
               "suggestions": suggestions,
               "evolution_type": "workout_feedback"
           }

           new_plan_data = await self.ai_service.evolve_plan(evolution_context)
           
           # Create new plan version
           new_plan = Plan(
               jsonb_plan=new_plan_data,
               version=current_plan.version + 1,
               parent_plan_id=current_plan.id
           )

           self.db.add(new_plan)
           await self.db.commit()
           return new_plan
   ```

2. **Enhanced API Endpoints**
   ```python
   # backend/app/routes/workouts.py
   @router.post("/workouts/{workout_id}/analyze")
   async def analyze_workout(
       workout_id: int,
       background_tasks: BackgroundTasks,
       db: AsyncSession = Depends(get_db)
   ):
       """Trigger AI analysis of a specific workout."""
       workout = await get_workout_by_id(db, workout_id)
       if not workout:
           raise HTTPException(status_code=404, detail="Workout not found")

       ai_service = AIService(db)
       background_tasks.add_task(
           analyze_and_store_workout,
           db, workout, ai_service
       )
       
       return {"message": "Analysis started", "workout_id": workout_id}

   @router.post("/analyses/{analysis_id}/approve")
   async def approve_analysis(
       analysis_id: int,
       db: AsyncSession = Depends(get_db)
   ):
       """Approve analysis suggestions and trigger plan evolution."""
       analysis = await get_analysis_by_id(db, analysis_id)
       analysis.approved = True
       
       # Trigger plan evolution if suggestions exist
       if analysis.suggestions:
           evolution_service = PlanEvolutionService(db)
           current_plan = await get_current_active_plan(db)
           if current_plan:
               new_plan = await evolution_service.evolve_plan_from_analysis(
                   analysis, current_plan
               )
               return {"message": "Analysis approved", "new_plan_id": new_plan.id}
       
       await db.commit()
       return {"message": "Analysis approved"}
   ```

### Phase 4: Frontend Implementation (Week 9-11)

#### Week 9: Core Components
1. **Garmin Sync Interface**
   ```jsx
   // frontend/src/components/GarminSync.jsx
   import { useState, useEffect } from 'react';

   const GarminSync = () => {
     const [syncStatus, setSyncStatus] = useState(null);
     const [syncing, setSyncing] = useState(false);

     const triggerSync = async () => {
       setSyncing(true);
       try {
         await fetch('/api/workouts/sync', { method: 'POST' });
         // Poll for status updates
         pollSyncStatus();
       } catch (error) {
         console.error('Sync failed:', error);
         setSyncing(false);
       }
     };

     const pollSyncStatus = () => {
       const interval = setInterval(async () => {
         const response = await fetch('/api/workouts/sync-status');
         const status = await response.json();
         setSyncStatus(status);
         
         if (status.status !== 'in_progress') {
           setSyncing(false);
           clearInterval(interval);
         }
       }, 2000);
     };

     return (
       <div className="garmin-sync">
         <h3>Garmin Connect Sync</h3>
         <button 
           onClick={triggerSync} 
           disabled={syncing}
           className="sync-button"
         >
           {syncing ? 'Syncing...' : 'Sync Recent Activities'}
         </button>
         
         {syncStatus && (
           <div className="sync-status">
             <p>Last sync: {new Date(syncStatus.last_sync_time).toLocaleString()}</p>
             <p>Status: {syncStatus.status}</p>
             {syncStatus.activities_synced > 0 && (
               <p>Activities synced: {syncStatus.activities_synced}</p>
             )}
           </div>
         )}
       </div>
     );
   };
   ```

2. **Workout Analysis Interface**
   ```jsx
   // frontend/src/components/WorkoutAnalysis.jsx
   const WorkoutAnalysis = ({ workout, analysis }) => {
     const [approving, setApproving] = useState(false);

     const approveAnalysis = async () => {
       setApproving(true);
       try {
         const response = await fetch(`/api/analyses/${analysis.id}/approve`, {
           method: 'POST'
         });
         const result = await response.json();
         
         if (result.new_plan_id) {
           // Navigate to new plan or show success message
           console.log('New plan created:', result.new_plan_id);
         }
       } catch (error) {
         console.error('Approval failed:', error);
       } finally {
         setApproving(false);
       }
     };

     return (
       <div className="workout-analysis">
         <div className="workout-summary">
           <h3>{workout.activity_type} - {new Date(workout.start_time).toLocaleDateString()}</h3>
           <div className="metrics">
             <span>Duration: {Math.round(workout.duration_seconds / 60)}min</span>
             <span>Distance: {(workout.distance_m / 1000).toFixed(1)}km</span>
             {workout.avg_power && <span>Avg Power: {workout.avg_power}W</span>}
             {workout.avg_hr && <span>Avg HR: {workout.avg_hr}bpm</span>}
           </div>
         </div>

         {analysis && (
           <div className="analysis-content">
             <h4>AI Analysis</h4>
             <div className="feedback">
               {analysis.jsonb_feedback.summary}
             </div>
             
             {analysis.suggestions && (
               <div className="suggestions">
                 <h5>Suggestions</h5>
                 <ul>
                   {analysis.suggestions.map((suggestion, index) => (
                     <li key={index}>{suggestion}</li>
                   ))}
                 </ul>
                 
                 {!analysis.approved && (
                   <button 
                     onClick={approveAnalysis}
                     disabled={approving}
                     className="approve-button"
                   >
                     {approving ? 'Approving...' : 'Approve & Update Plan'}
                   </button>
                 )}
               </div>
             )}
           </div>
         )}
       </div>
     );
   };
   ```

#### Week 10: Data Visualization
1. **Workout Charts**
   ```jsx
   // Using recharts for workout data visualization
   import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';

   const WorkoutChart = ({ workoutData }) => {
     return (
       <div className="workout-charts">
         <h4>Workout Metrics</h4>
         <LineChart width={800} height={400} data={workoutData.time_series}>
           <CartesianGrid strokeDasharray="3 3" />
           <XAxis dataKey="time" />
           <YAxis yAxisId="left" />
           <YAxis yAxisId="right" orientation="right" />
           <Tooltip />
           <Legend />
           <Line yAxisId="left" type="monotone" dataKey="heart_rate" stroke="#8884d8" />
           <Line yAxisId="right" type="monotone" dataKey="power" stroke="#82ca9d" />
           <Line yAxisId="left" type="monotone" dataKey="cadence" stroke="#ffc658" />
         </LineChart>
       </div>
     );
   };
   ```

2. **Plan Timeline View**
   ```jsx
   // Plan visualization with version history
   const PlanTimeline = ({ plan, versions }) => {
     return (
       <div className="plan-timeline">
         <h3>Training Plan - Version {plan.version}</h3>
         
         {versions.length > 1 && (
           <div className="version-history">
             <h4>Version History</h4>
             {versions.map(version => (
               <div key={version.id} className="version-item">
                 <span>v{version.version}</span>
                 <span>{new Date(version.created_at).toLocaleDateString()}</span>
                 {version.parent_plan_id && <span>→ Evolved from analysis</span>}
               </div>
             ))}
           </div>
         )}
         
         <div className="plan-weeks">
           {plan.jsonb_plan.weeks.map((week, index) => (
             <div key={index} className="week-card">
               <h4>Week {index + 1}</h4>
               {week.workouts.map((workout, wIndex) => (
                 <div key={wIndex} className="workout-item">
                   <span>{workout.type}</span>
                   <span>{workout.duration}min</span>
                   <span>{workout.intensity}</span>
                 </div>
               ))}
             </div>
           ))}
         </div>
       </div>
     );
   };
   ```

#### Week 11: Integration & Polish
1. **Dashboard Overview**
2. **File Upload Improvements**
3. **Error Handling & Loading States**
4. **Responsive Design**

### Phase 5: Testing and Deployment (Week 12-13)

#### Week 12: Testing
1. **Backend Testing**
   ```python
   # tests/test_garmin_service.py
   import pytest
   from unittest.mock import AsyncMock, patch
   from app.services.garmin import GarminService

   @pytest.mark.asyncio
   async def test_garmin_authentication():
       with patch('garth.Client') as mock_client:
           service = GarminService()
           await service.authenticate()
           mock_client.return_value.login.assert_called_once()

   @pytest.mark.asyncio
   async def test_activity_sync(db_session):
       # Test workout sync functionality
       pass
   ```

2. **Integration Tests**
3. **Frontend Component Tests**

#### Week 13: Deployment Preparation
1. **Environment Configuration**
   ```bash
   # .env.production
   GARMIN_USERNAME=your_garmin_email
   GARMIN_PASSWORD=your_garmin_password
   OPENROUTER_API_KEY=your_openrouter_key
   AI_MODEL=anthropic/claude-3-sonnet-20240229
   API_KEY=your_secure_api_key
   ```

2. **Production Docker Setup**
3. **Backup Strategy for Database and GPX Files**
4. **Monitoring and Logging**

## Key Technical Decisions

### Single-User Simplifications
- **Authentication**: Simple API key instead of complex user management
- **File Storage**: Local filesystem (200 GPX files easily manageable)
- **Database**: Single tenant, no multi-user complexity
- **Deployment**: Self-hosted container, no cloud scaling needs

### Garmin Integration Strategy
- **garth library**: Python library for Garmin Connect API
- **Session persistence**: Store auth sessions in mounted volume
- **Background sync**: Async background tasks for activity fetching
- **Retry logic**: Handle API rate limits and temporary failures

### AI Integration Approach
- **Prompt versioning**: Database-stored prompts with version control
- **Async processing**: Non-blocking AI calls with background tasks
- **Cost management**: Simple retry logic, no complex rate limiting needed for single user
- **Response parsing**: Flexible parsing for different AI response formats

### Database Design Philosophy
- **Versioning everywhere**: Plans, rules, and prompts all support evolution
- **JSONB storage**: Flexible storage for AI responses and complex data
- **Audit trail**: Track plan evolution and analysis approval history

## Environment Variables
```bash
# Required environment variables
GARMIN_USERNAME=your_garmin_email
GARMIN_PASSWORD=your_garmin_password
OPENROUTER_API_KEY=your_openrouter_key
AI_MODEL=anthropic/claude-3-sonnet-20240229
API_KEY=your_secure_api_key

# Optional
DATABASE_URL=postgresql://postgres:password@db:5432/cycling
REACT_APP_API_URL=http://localhost:8000
REACT_APP_API_KEY=your_secure_api_key
```

## Python Standards and Best Practices

### Code Style and Structure
```python
# Example: Proper async service implementation
from typing import Optional, List, Dict, Any
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, update
import logging

logger = logging.getLogger(__name__)

class WorkoutAnalysisService:
    """Service for analyzing workout data with AI assistance."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
        self.ai_service = AIService(db)
    
    async def analyze_workout_performance(
        self, 
        workout_id: int, 
        comparison_metrics: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Analyze workout performance against planned metrics.
        
        Args:
            workout_id: The workout to analyze
            comparison_metrics: Optional baseline metrics for comparison
            
        Returns:
            Dict containing analysis results and suggestions
            
        Raises:
            WorkoutNotFoundError: If workout doesn't exist
            AIServiceError: If AI analysis fails
        """
        try:
            workout = await self._get_workout(workout_id)
            if not workout:
                raise WorkoutNotFoundError(f"Workout {workout_id} not found")
                
            analysis_data = await self._prepare_analysis_context(workout, comparison_metrics)
            ai_analysis = await self.ai_service.analyze_workout(analysis_data)
            
            # Store analysis results
            analysis_record = await self._store_analysis(workout_id, ai_analysis)
            
            logger.info(f"Successfully analyzed workout {workout_id}")
            return {
                "analysis_id": analysis_record.id,
                "feedback": ai_analysis.get("feedback"),
                "suggestions": ai_analysis.get("suggestions"),
                "performance_score": ai_analysis.get("score")
            }
            
        except Exception as e:
            logger.error(f"Failed to analyze workout {workout_id}: {str(e)}")
            raise
    
    async def _get_workout(self, workout_id: int) -> Optional[Workout]:
        """Retrieve workout by ID."""
        result = await self.db.execute(
            select(Workout).where(Workout.id == workout_id)
        )
        return result.scalar_one_or_none()
    
    async def _prepare_analysis_context(
        self, 
        workout: Workout, 
        comparison_metrics: Optional[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """Prepare context data for AI analysis."""
        context = {
            "workout_data": {
                "duration_minutes": workout.duration_seconds / 60 if workout.duration_seconds else 0,
                "distance_km": workout.distance_m / 1000 if workout.distance_m else 0,
                "avg_power": workout.avg_power,
                "avg_heart_rate": workout.avg_hr,
                "elevation_gain": workout.elevation_gain_m
            },
            "activity_type": workout.activity_type,
            "date": workout.start_time.isoformat() if workout.start_time else None
        }
        
        if comparison_metrics:
            context["baseline_metrics"] = comparison_metrics
            
        return context
```

### Error Handling Patterns
```python
# Custom exceptions for better error handling
class CyclingCoachError(Exception):
    """Base exception for cycling coach application."""
    pass

class WorkoutNotFoundError(CyclingCoachError):
    """Raised when a workout cannot be found."""
    pass

class GarminSyncError(CyclingCoachError):
    """Raised when Garmin synchronization fails."""
    pass

class AIServiceError(CyclingCoachError):
    """Raised when AI service requests fail."""
    pass

# Middleware for consistent error responses
@app.exception_handler(CyclingCoachError)
async def cycling_coach_exception_handler(request: Request, exc: CyclingCoachError):
    return JSONResponse(
        status_code=400,
        content={
            "error": exc.__class__.__name__,
            "message": str(exc),
            "timestamp": datetime.now().isoformat()
        }
    )
```

### Database Patterns
```python
# Proper async database patterns with context managers
from contextlib import asynccontextmanager

class DatabaseService:
    """Base service class with database session management."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    @asynccontextmanager
    async def transaction(self):
        """Context manager for database transactions."""
        try:
            yield self.db
            await self.db.commit()
        except Exception:
            await self.db.rollback()
            raise
    
    async def get_or_create(self, model_class, **kwargs):
        """Get existing record or create new one."""
        result = await self.db.execute(
            select(model_class).filter_by(**kwargs)
        )
        instance = result.scalar_one_or_none()
        
        if not instance:
            instance = model_class(**kwargs)
            self.db.add(instance)
            await self.db.flush()  # Get ID without committing
            
        return instance
```

## Sample Prompt Templates

### Workout Analysis Prompt
```sql
INSERT INTO prompts (action_type, model, prompt_text, version, active) VALUES (
'workout_analysis',
'anthropic/claude-3-sonnet-20240229',
'Analyze the following cycling workout data and provide structured feedback:

Workout Details:
- Activity Type: {activity_type}
- Duration: {duration_minutes} minutes  
- Distance: {distance_km} km
- Average Power: {avg_power}W
- Average Heart Rate: {avg_hr} bpm
- Elevation Gain: {elevation_gain}m

Please provide your analysis in the following JSON format:
{{
  "performance_summary": "Brief overall assessment",
  "strengths": ["strength 1", "strength 2"],
  "areas_for_improvement": ["area 1", "area 2"],
  "training_suggestions": ["suggestion 1", "suggestion 2"],
  "next_workout_recommendations": {{
    "intensity": "easy/moderate/hard",
    "focus": "endurance/power/recovery",
    "duration_minutes": 60
  }},
  "performance_score": 8.5
}}

Focus on actionable insights and specific recommendations for improvement.',
1,
true
);
```

### Plan Generation Prompt
```sql
INSERT INTO prompts (action_type, model, prompt_text, version, active) VALUES (
'plan_generation',
'anthropic/claude-3-sonnet-20240229',
'Create a personalized cycling training plan based on the following information:

Training Rules:
{rules}

Goals:
{goals}

Generate a 4-week training plan in the following JSON format:
{{
  "plan_overview": {{
    "duration_weeks": 4,
    "focus": "endurance/power/mixed",
    "weekly_hours": 8
  }},
  "weeks": [
    {{
      "week_number": 1,
      "focus": "base building",
      "workouts": [
        {{
          "day": "monday",
          "type": "easy_ride",
          "duration_minutes": 60,
          "intensity": "zone_1_2",
          "description": "Easy recovery ride"
        }}
      ]
    }}
  ],
  "progression_notes": "How the plan builds over the weeks"
}}

Ensure all workouts respect the training rules provided.',
1,
true
);
```

## Deployment Configuration

### Production Docker Compose
```yaml
# docker-compose.prod.yml
version: '3.9'
services:
  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile.prod
    restart: unless-stopped
    ports:
      - "8000:8000"
    volumes:
      - gpx-data:/app/data/gpx:rw
      - garmin-sessions:/app/data/sessions:rw
      - ./logs:/app/logs:rw
    environment:
      - DATABASE_URL=postgresql://postgres:${POSTGRES_PASSWORD}@db:5432/cycling
      - GARMIN_USERNAME=${GARMIN_USERNAME}
      - GARMIN_PASSWORD=${GARMIN_PASSWORD}
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
      - AI_MODEL=${AI_MODEL}
      - API_KEY=${API_KEY}
      - LOG_LEVEL=INFO
    depends_on:
      - db
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - REACT_APP_API_URL=http://localhost:8000
      - NODE_ENV=production
    depends_on:
      - backend

  db:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: cycling
    volumes:
      - postgres-data:/var/lib/postgresql/data:rw
      - ./backups:/backups:rw
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  gpx-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /home/user/cycling-coach/data/gpx
  garmin-sessions:
    driver: local  
    driver_opts:
      type: none
      o: bind
      device: /home/user/cycling-coach/data/sessions
  postgres-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /home/user/cycling-coach/data/postgres

networks:
  default:
    name: cycling-coach
```

### Backup Script
```bash
#!/bin/bash
# backup.sh - Daily backup script

BACKUP_DIR="/home/user/cycling-coach/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup database
docker exec cycling-coach-db-1 pg_dump -U postgres cycling > "$BACKUP_DIR/db_backup_$DATE.sql"

# Backup GPX files
tar -czf "$BACKUP_DIR/gpx_backup_$DATE.tar.gz" -C /home/user/cycling-coach/data/gpx .

# Backup Garmin sessions
tar -czf "$BACKUP_DIR/sessions_backup_$DATE.tar.gz" -C /home/user/cycling-coach/data/sessions .

# Keep only last 30 days of backups
find "$BACKUP_DIR" -name "*backup*" -type f -mtime +30 -delete

echo "Backup completed: $DATE"
```

### Health Monitoring
```python
# backend/app/routes/health.py
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from app.database import get_db
from app.services.garmin import GarminService

router = APIRouter()

@router.get("/health")
async def health_check(db: AsyncSession = Depends(get_db)):
    """Health check endpoint for monitoring."""
    health_status = {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "services": {}
    }
    
    # Check database
    try:
        await db.execute(text("SELECT 1"))
        health_status["services"]["database"] = "healthy"
    except Exception as e:
        health_status["services"]["database"] = f"error: {str(e)}"
        health_status["status"] = "unhealthy"
    
    # Check Garmin service
    try:
        garmin_service = GarminService()
        # Simple connectivity check without full auth
        health_status["services"]["garmin"] = "configured"
    except Exception as e:
        health_status["services"]["garmin"] = f"error: {str(e)}"
    
    # Check file system
    try:
        gpx_dir = "/app/data/gpx"
        if os.path.exists(gpx_dir) and os.access(gpx_dir, os.W_OK):
            health_status["services"]["file_storage"] = "healthy"
        else:
            health_status["services"]["file_storage"] = "error: directory not writable"
            health_status["status"] = "unhealthy"
    except Exception as e:
        health_status["services"]["file_storage"] = f"error: {str(e)}"
        health_status["status"] = "unhealthy"
    
    if health_status["status"] == "unhealthy":
        raise HTTPException(status_code=503, detail=health_status)
    
    return health_status
```

## Post-Deployment Setup

### Initial Data Setup
```python
# scripts/init_prompts.py
"""Initialize default AI prompts in the database."""
import asyncio
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from app.models.prompt import Prompt

async def init_default_prompts():
    """Initialize the database with default AI prompts."""
    engine = create_async_engine(DATABASE_URL)
    
    async with AsyncSession(engine) as session:
        # Add default prompts for each action type
        default_prompts = [
            # Workout analysis prompt (from above)
            # Plan generation prompt (from above)  
            # Rule parsing prompt
        ]
        
        for prompt_data in default_prompts:
            prompt = Prompt(**prompt_data)
            session.add(prompt)
        
        await session.commit()
    
    print("Default prompts initialized successfully")

if __name__ == "__main__":
    asyncio.run(init_default_prompts())
```

### Maintenance Tasks
```python
# scripts/maintenance.py
"""Maintenance tasks for the cycling coach application."""

async def cleanup_old_analyses():
    """Remove analyses older than 6 months."""
    cutoff_date = datetime.now() - timedelta(days=180)
    
    async with AsyncSession(engine) as session:
        result = await session.execute(
            delete(Analysis).where(Analysis.created_at < cutoff_date)
        )
        await session.commit()
        print(f"Deleted {result.rowcount} old analyses")

async def optimize_database():
    """Run database maintenance tasks."""
    async with AsyncSession(engine) as session:
        await session.execute(text("VACUUM ANALYZE"))
        await session.commit()
        print("Database optimization completed")
```

This comprehensive implementation plan addresses all the key requirements for your single-user, self-hosted AI-assisted cycling coach application. The plan includes:

1. **Complete Garmin integration** using environment variables and the garth library
2. **Enhanced database schema** with proper versioning for plans and rules
3. **Robust AI integration** with prompt management and error handling
4. **Production-ready deployment** configuration with health checks and backups
5. **Comprehensive testing strategy** for both backend and frontend
6. **Maintenance and monitoring** tools for long-term operation

The plan is well-suited for your scale (single user, 200 GPX files) and deployment target (self-hosted container), with practical simplifications that avoid unnecessary complexity while maintaining professional software engineering standards.