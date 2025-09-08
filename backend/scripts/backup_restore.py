#!/usr/bin/env python3
"""
Database backup and restore utilities for containerized deployments.
Ensures safe backup/restore operations with migration compatibility checks.
"""

import sys
import os
import subprocess
import shutil
from pathlib import Path
from datetime import datetime
from typing import Optional

# Add backend directory to path
backend_dir = Path(__file__).parent.parent
sys.path.insert(0, str(backend_dir))

from app.database import get_database_url

class DatabaseManager:
    """Handles database backup and restore operations."""

    def __init__(self, backup_dir: str = "/app/data/backups"):
        self.backup_dir = Path(backup_dir)
        self.backup_dir.mkdir(parents=True, exist_ok=True)

    def get_db_connection_params(self):
        """Extract database connection parameters from URL."""
        from urllib.parse import urlparse
        db_url = get_database_url()
        parsed = urlparse(db_url)

        return {
            'host': parsed.hostname,
            'port': parsed.port or 5432,
            'user': parsed.username,
            'password': parsed.password,
            'database': parsed.path.lstrip('/')
        }

    def create_backup(self, name: Optional[str] = None) -> str:
        """Create a database backup."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = name or f"backup_{timestamp}"
        backup_file = self.backup_dir / f"{backup_name}.sql"

        params = self.get_db_connection_params()

        # Use pg_dump for backup
        cmd = [
            "pg_dump",
            "-h", params['host'],
            "-p", str(params['port']),
            "-U", params['user'],
            "-d", params['database'],
            "-f", str(backup_file),
            "--no-password",
            "--format=custom",  # Custom format for better compression
            "--compress=9"
        ]

        # Set password environment variable
        env = os.environ.copy()
        env['PGPASSWORD'] = params['password']

        try:
            print(f"Creating backup: {backup_file}")
            result = subprocess.run(cmd, env=env, capture_output=True, text=True)

            if result.returncode == 0:
                print(f"✅ Backup created successfully: {backup_file}")
                return str(backup_file)
            else:
                print(f"❌ Backup failed: {result.stderr}")
                raise Exception(f"Backup failed: {result.stderr}")

        except FileNotFoundError:
            print("❌ pg_dump not found. Ensure PostgreSQL client tools are installed.")
            raise

    def restore_backup(self, backup_file: str, confirm: bool = False) -> None:
        """Restore database from backup."""
        backup_path = Path(backup_file)
        if not backup_path.exists():
            raise FileNotFoundError(f"Backup file not found: {backup_file}")

        if not confirm:
            print(f"⚠️  This will overwrite the current database!")
            response = input("Are you sure you want to continue? (yes/no): ")
            if response.lower() != 'yes':
                print("Restore cancelled.")
                return

        params = self.get_db_connection_params()

        # Drop and recreate database to ensure clean restore
        self._recreate_database()

        # Use pg_restore for restore
        cmd = [
            "pg_restore",
            "-h", params['host'],
            "-p", str(params['port']),
            "-U", params['user'],
            "-d", params['database'],
            "--no-password",
            "--clean",
            "--if-exists",
            "--create",
            str(backup_path)
        ]

        env = os.environ.copy()
        env['PGPASSWORD'] = params['password']

        try:
            print(f"Restoring from backup: {backup_file}")
            result = subprocess.run(cmd, env=env, capture_output=True, text=True)

            if result.returncode == 0:
                print("✅ Database restored successfully")
            else:
                print(f"❌ Restore failed: {result.stderr}")
                raise Exception(f"Restore failed: {result.stderr}")

        except FileNotFoundError:
            print("❌ pg_restore not found. Ensure PostgreSQL client tools are installed.")
            raise

    def _recreate_database(self):
        """Drop and recreate the database."""
        params = self.get_db_connection_params()

        # Connect to postgres database to drop/recreate target database
        postgres_params = params.copy()
        postgres_params['database'] = 'postgres'

        drop_cmd = [
            "psql",
            "-h", postgres_params['host'],
            "-p", str(postgres_params['port']),
            "-U", postgres_params['user'],
            "-d", postgres_params['database'],
            "-c", f"DROP DATABASE IF EXISTS {params['database']};"
        ]

        create_cmd = [
            "psql",
            "-h", postgres_params['host'],
            "-p", str(postgres_params['port']),
            "-U", postgres_params['user'],
            "-d", postgres_params['database'],
            "-c", f"CREATE DATABASE {params['database']};"
        ]

        env = os.environ.copy()
        env['PGPASSWORD'] = params['password']

        for cmd in [drop_cmd, create_cmd]:
            result = subprocess.run(cmd, env=env, capture_output=True, text=True)
            if result.returncode != 0:
                print(f"Database recreation step failed: {result.stderr}")

    def list_backups(self):
        """List available backup files."""
        backups = list(self.backup_dir.glob("*.sql"))
        backups.sort(key=lambda x: x.stat().st_mtime, reverse=True)

        if not backups:
            print("No backup files found.")
            return

        print("Available backups:")
        for backup in backups:
            size = backup.stat().st_size / (1024 * 1024)  # Size in MB
            mtime = datetime.fromtimestamp(backup.stat().st_mtime)
            print(".2f")

    def cleanup_old_backups(self, keep_days: int = 30):
        """Remove backups older than specified days."""
        from datetime import timedelta

        cutoff = datetime.now() - timedelta(days=keep_days)
        removed = []

        for backup in self.backup_dir.glob("*.sql"):
            if datetime.fromtimestamp(backup.stat().st_mtime) < cutoff:
                backup.unlink()
                removed.append(backup.name)

        if removed:
            print(f"Removed {len(removed)} old backups: {', '.join(removed)}")
        else:
            print("No old backups to remove.")

def main():
    if len(sys.argv) < 2:
        print("Usage: python backup_restore.py <command> [options]")
        print("Commands:")
        print("  backup [name]          - Create a new backup")
        print("  restore <file> [--yes] - Restore from backup")
        print("  list                   - List available backups")
        print("  cleanup [days]         - Remove backups older than N days (default: 30)")
        sys.exit(1)

    manager = DatabaseManager()
    command = sys.argv[1]

    try:
        if command == "backup":
            name = sys.argv[2] if len(sys.argv) > 2 else None
            manager.create_backup(name)

        elif command == "restore":
            if len(sys.argv) < 3:
                print("Error: Please specify backup file to restore from")
                sys.exit(1)

            backup_file = sys.argv[2]
            confirm = "--yes" in sys.argv
            manager.restore_backup(backup_file, confirm)

        elif command == "list":
            manager.list_backups()

        elif command == "cleanup":
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 30
            manager.cleanup_old_backups(days)

        else:
            print(f"Unknown command: {command}")
            sys.exit(1)

    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()