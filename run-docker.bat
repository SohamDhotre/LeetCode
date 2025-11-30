@echo off
echo 🐳 Running LeetCode Sync in Docker...
echo.

REM Check if .env exists
if not exist .env (
    echo ❌ Error: .env file not found
    echo Please copy .env.example to .env and configure it
    exit /b 1
)

REM Run docker-compose in detached mode (background)
docker-compose up -d --build

echo.
echo ✅ LeetCode Sync Daemon started in background!
echo 📜 To view logs: docker-compose logs -f
echo 🛑 To stop: docker-compose down
