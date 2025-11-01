@echo off
chcp 65001 >nul
echo ================================================================
echo Building clash-speedtest...
echo ================================================================

go build -ldflags "-s -w" -o clash-speedtest.exe ./cmd

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ================================================================
    echo Build successful!
    echo Output: clash-speedtest.exe
    echo ================================================================
) else (
    echo.
    echo ================================================================
    echo Build failed! Please check the error message above.
    echo ================================================================
    exit /b 1
)


