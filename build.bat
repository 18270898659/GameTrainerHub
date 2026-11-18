@echo off
setlocal
echo ========================================================
echo   Building GameTrainerHub Windows Desktop Application
echo ========================================================

set "PATH=C:\Users\bigant\go\bin;C:\Program Files\Go\bin;C:\Program Files\nodejs;C:\Windows\System32;C:\Windows;%PATH%"

cd /d "%~dp0"

echo [1/3] Building Frontend...
cd frontend
call npm.cmd run build
if %ERRORLEVEL% NEQ 0 (
    echo Frontend build failed!
    exit /b %ERRORLEVEL%
)
cd ..

echo [2/3] Compiling Application with Wails...
where wails >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    wails build -clean -s
    if %ERRORLEVEL% EQU 0 (
        copy /y "build\bin\GameTrainerHub.exe" "GameTrainerHub.exe"
    )
) else (
    echo Wails CLI not in PATH, compiling with go build tags...
    go build -tags "desktop,production" -ldflags "-s -w" -o GameTrainerHub.exe main.go
)

if %ERRORLEVEL% EQU 0 (
    echo ========================================================
    echo   Build Succeeded: GameTrainerHub.exe
    echo ========================================================
) else (
    echo Build failed!
)

