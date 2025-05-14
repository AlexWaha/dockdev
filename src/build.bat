@echo off
echo === Go Build Process ===

REM Init Go modules if needed
if not exist "go.mod" (
    echo Initializing go.mod...
    go mod init generator
    go get golang.org/x/term
    go get github.com/joho/godotenv
) else (
    echo Tidying modules...
    go mod tidy
)

REM Set build target for Linux
set GOOS=linux
set GOARCH=amd64

REM Create dist if not exists
if not exist "..\dist" (
    mkdir "..\dist"
)

echo Building binary...
go build -o ..\dist\dockdev ./cmd

if exist "..\dist\dockdev" (
    echo Build successful: ..\dist\dockdev
) else (
    echo Build failed!
    exit /b 1
)
