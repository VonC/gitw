@echo off
setlocal enabledelayedexpansion

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%"
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

if "%1" == "amd" (
    shift
    ssh -t -q wsl "bash -c 'source .senv; ./%dirname%'"
    goto:eof
)
if exist senv.bat (
    call senv.bat
)
rem @echo on
del debug.log 2>NUL
call "%dirname%.exe" %*