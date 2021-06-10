@echo off
setlocal enabledelayedexpansion

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%"
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

if "%1" == "amd" (
    shift
    ssh -t -q wsl "bash -c 'ptest=1 PAGER_LOG=debug.log ./%dirname%'"
    goto:eof
)
if exist senv.bat (
    call senv.bat
)
@echo on
call "%dirname%.exe" %*