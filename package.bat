@echo off
setlocal enabledelayedexpansion

echo ==============================================================================
echo   Game Trainer Hub - Release Packaging Script (Installer + Portable)
echo ==============================================================================

set "PATH=C:\Users\bigant\go\bin;C:\Program Files\Go\bin;C:\Program Files\nodejs;C:\Program Files\Git\cmd;C:\Program Files (x86)\NSIS;C:\Program Files\NSIS;C:\Windows\System32;C:\Windows;%PATH%"

cd /d "%~dp0"

:: 1. Safety check: close running instances
echo [1/6] Checking for running instances...
taskkill /F /T /IM GameTrainerHub.exe >nul 2>nul

:: 2. Ensure output directory exists
if not exist "bin" mkdir "bin"

:: 3. Build frontend
echo [2/6] Building Frontend Assets...
cd frontend
call npm.cmd run build
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend build failed with exit code %ERRORLEVEL%!
    exit /b %ERRORLEVEL%
)
cd ..

:: 4. Build Wails executable
echo [3/6] Building Production Executable with Wails...
wails build -clean -s
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Wails build failed with exit code %ERRORLEVEL%!
    exit /b %ERRORLEVEL%
)

if not exist "build\bin\GameTrainerHub.exe" (
    echo [ERROR] Expected build\bin\GameTrainerHub.exe not found!
    exit /b 1
)

copy /y "build\bin\GameTrainerHub.exe" "bin\GameTrainerHub.exe" >nul
copy /y "build\bin\GameTrainerHub.exe" "GameTrainerHub.exe" >nul

:: 5. Create Portable ZIP
echo [4/6] Creating Portable ZIP package (GameTrainerHub-Portable-x64.zip)...
if exist "bin\temp_portable" rd /s /q "bin\temp_portable"
mkdir "bin\temp_portable"
copy /y "bin\GameTrainerHub.exe" "bin\temp_portable\GameTrainerHub.exe" >nul
copy /y "build\installer\NOTICE.txt" "bin\temp_portable\NOTICE.txt" >nul
copy /y "build\installer\README.txt" "bin\temp_portable\README.txt" >nul
copy /y "build\installer\version.json" "bin\temp_portable\version.json" >nul

if exist "bin\GameTrainerHub-Portable-x64.zip" del /f /q "bin\GameTrainerHub-Portable-x64.zip"
powershell -NoProfile -Command "Compress-Archive -Path 'bin\temp_portable\*' -DestinationPath 'bin\GameTrainerHub-Portable-x64.zip' -Force"
rd /s /q "bin\temp_portable"

if not exist "bin\GameTrainerHub-Portable-x64.zip" (
    echo [ERROR] Failed to create Portable ZIP package!
    exit /b 1
)

:: 6. Compile NSIS Installer
echo [5/6] Compiling NSIS Installer (GameTrainerHub-Setup-x64.exe)...
set "MAKENSIS_EXE=makensis.exe"
if exist "C:\Program Files (x86)\NSIS\makensis.exe" set "MAKENSIS_EXE=C:\Program Files (x86)\NSIS\makensis.exe"
if exist "C:\Program Files\NSIS\makensis.exe" set "MAKENSIS_EXE=C:\Program Files\NSIS\makensis.exe"

"%MAKENSIS_EXE%" /V3 /INPUTCHARSET UTF8 build\installer\installer.nsi
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] NSIS installer compilation failed with exit code %ERRORLEVEL%!
    exit /b %ERRORLEVEL%
)

if not exist "bin\GameTrainerHub-Setup-x64.exe" (
    echo [ERROR] Expected bin\GameTrainerHub-Setup-x64.exe not found!
    exit /b 1
)

:: 7. Generate Checksums
echo [6/6] Computing SHA256 Checksums...
powershell -NoProfile -Command ^
    "$installer = Get-FileHash -Algorithm SHA256 'bin\GameTrainerHub-Setup-x64.exe';" ^
    "$portable = Get-FileHash -Algorithm SHA256 'bin\GameTrainerHub-Portable-x64.zip';" ^
    "$exe = Get-FileHash -Algorithm SHA256 'bin\GameTrainerHub.exe';" ^
    "$lines = @(" ^
    "  \"GameTrainerHub Packaging Release Checksums (SHA256)\"," ^
    "  \"Generated at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')\"," ^
    "  \"--------------------------------------------------------------------------------\"," ^
    "  \"$($installer.Hash)  GameTrainerHub-Setup-x64.exe ($([math]::Round((Get-Item $installer.Path).Length / 1MB, 2)) MB)\"," ^
    "  \"$($portable.Hash)  GameTrainerHub-Portable-x64.zip ($([math]::Round((Get-Item $portable.Path).Length / 1MB, 2)) MB)\"," ^
    "  \"$($exe.Hash)  GameTrainerHub.exe ($([math]::Round((Get-Item $exe.Path).Length / 1MB, 2)) MB)\"" ^
    ");" ^
    "$lines | Out-File -FilePath 'bin\checksums.txt' -Encoding utf8;" ^
    "$lines | ForEach-Object { Write-Host $_ }"

echo ==============================================================================
echo   Build and Packaging Completed Successfully!
echo   Deliverables located in: %~dp0bin
echo ==============================================================================
