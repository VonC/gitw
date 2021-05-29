@echo off
setlocal enabledelayedexpansion

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%"
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

if not exist buildversion.exe (
    go build version/cmd/buildversion.go
    if errorlevel 1 (
        echo ERROR BUILD BUILD VERSION 1>&2
        exit /b 1
    )
)
buildversion.exe
if errorlevel 1 (
    echo ERROR RUN BUILD VERSION 1>&2
    exit /b 1
)

if "%1" == "amd" (
    set GOARCH=amd64
    set GOOS=linux
    go build
    if errorlevel 1 (
        echo ERROR BUILD 1>&2
        exit /b 1
    )
    goto:eof
)

go build
if errorlevel 1 (
    echo ERROR BUILD 1>&2
    exit /b 1
)
if "%1" neq "" ( %dirname% %* )