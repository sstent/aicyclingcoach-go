from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    DATABASE_URL: str
    GPX_STORAGE_PATH: str
    AI_MODEL: str = "openrouter/auto"
    
    class Config:
        env_file = ".env"

settings = Settings()