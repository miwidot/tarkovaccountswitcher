@echo off
set PATH=C:\mingw64\bin;%PATH%
set CGO_ENABLED=1
cd /d C:\Users\miwi\Documents\foobar\v2
"C:\Program Files\Go\bin\go.exe" build -v -ldflags="-H windowsgui -s -w" -o "Tarkov Account Switcher.exe" . > build_output.txt 2>&1
if %ERRORLEVEL% EQU 0 (
    echo BUILD SUCCESS >> build_output.txt
    dir "Tarkov Account Switcher.exe" >> build_output.txt
) else (
    echo BUILD FAILED with error %ERRORLEVEL% >> build_output.txt
)
