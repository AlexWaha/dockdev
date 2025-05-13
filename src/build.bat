@echo off
echo Initializing Go build environment...

REM Check if go.mod exists
if not exist "go.mod" (
    echo Initializing Go modules...
    go mod init generator
    
    REM Add dependencies
    echo Adding required dependencies...
    go get golang.org/x/term
    go get github.com/joho/godotenv
    
    echo Go modules initialized and dependencies added.
) else (
    echo Updating Go dependencies...
    go mod tidy
)

REM Set GOOS and GOARCH for Linux build
echo Building for Linux...
set GOOS=linux
set GOARCH=amd64

REM Create dist folder if it doesn't exist
if not exist "..\dist" (
    mkdir "..\dist"
)

REM Build the binary into ../dist
go build -o ..\dist\dockdev ./cmd

REM Check result
if exist ..\dist\dockdev (
    echo Build successful: ..\dist\dockdev
) else (
    echo Build failed! No output binary found.
    exit /b 1
)

echo Build process completed.

echo.
echo NOTE: This application requires Docker Desktop to be installed and running.
echo The application will check for Docker Desktop availability at runtime.
echo If Docker Desktop is not running, the application will offer to start it.
echo.
