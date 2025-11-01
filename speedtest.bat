@echo off
chcp 65001 >nul
REM ================================================================
REM  Clash Speedtest - Node Speed Test Script
REM ================================================================
REM  Usage: speedtest.bat [keyword]
REM  
REM  Examples:
REM    speedtest.bat           Test all nodes
REM    speedtest.bat HK        Test Hong Kong nodes
REM    speedtest.bat US        Test US nodes
REM ================================================================

setlocal enabledelayedexpansion

REM ================================================================
REM  Configuration Section - Modify these values as needed
REM ================================================================

REM Clash API Configuration
set CLASH_ADDR=http://127.0.0.1:9090
set CLASH_SECRET=set-your-secret

REM Output Configuration
set FORMAT=png
REM Options: txt (text table) or png (image output)

REM Ping Configuration
set ENABLE_PING=true
set PING_INTERVAL=30
REM Ping interval in seconds (recommended: 30-60)

set PING_TIMEOUT=5000
REM Ping timeout in milliseconds (recommended: 3000-8000)

REM Concurrent Testing Configuration
set CONCURRENT_NODES=3
REM Number of nodes to test simultaneously (1-5)
REM   1 = Serial (slowest, most stable)
REM   2 = Recommended (2x faster, stable)
REM   3 = Fast (2.5x faster)
REM   5 = Fastest (5x faster, may be unstable)

REM Download Configuration
set THREADS=1
REM Number of download threads per node (1-3)

set SIZE=20
REM Download size per thread in MB (10-20)

REM First parameter as filter keyword
set FILTER=%~1

REM Check if executable exists
if not exist "clash-speedtest.exe" (
    echo [ERROR] clash-speedtest.exe not found
    echo Please compile first: go build -ldflags "-s -w" -o clash-speedtest.exe ./cmd
    pause
    exit /b 1
)

echo ================================================================
echo   Clash Speedtest - Node Speed Test Tool (Concurrent Mode)
echo ================================================================
echo.

REM Display configuration
echo [Configuration]
echo   Clash Address:     %CLASH_ADDR%
if defined CLASH_SECRET (
    echo   Clash Secret:      [Set]
) else (
    echo   Clash Secret:      [Not Set]
)
echo.
echo [Test Settings]
echo   Output Format:     %FORMAT%
echo   Concurrent Nodes:  %CONCURRENT_NODES% (simultaneous testing)
echo   Download Threads:  %THREADS%
echo   Download Size:     %SIZE% MB per thread
if defined FILTER (
    echo   Filter Keyword:    "%FILTER%"
) else (
    echo   Filter Keyword:    [All Nodes]
)
echo.
if "%ENABLE_PING%"=="true" (
    echo [Ping Monitor]
    echo   Status:            Enabled
    echo   Interval:          %PING_INTERVAL% seconds
    echo   Timeout:           %PING_TIMEOUT% ms
) else (
    echo [Ping Monitor]
    echo   Status:            Disabled
)
echo.

REM Build command
set CMD=clash-speedtest.exe --clash-addr "%CLASH_ADDR%"

if defined CLASH_SECRET (
    set CMD=!CMD! --clash-secret "%CLASH_SECRET%"
)

if defined FILTER (
    set CMD=!CMD! -i "%FILTER%"
)

set CMD=!CMD! -f %FORMAT% -t %THREADS% -r %SIZE% -n %CONCURRENT_NODES%

if "%ENABLE_PING%"=="true" (
    set CMD=!CMD! -p --ping-interval %PING_INTERVAL% --ping-timeout %PING_TIMEOUT%
)

set CMD=!CMD! -y

REM Display command
echo ================================================================
echo [Command]
echo   %CMD%
echo ================================================================
echo.
echo Starting speed test...
echo.

REM Execute speed test
%CMD%

set EXIT_CODE=%ERRORLEVEL%

echo.
echo ================================================================
if %EXIT_CODE% EQU 0 (
    echo   Speed test completed successfully!
    echo   Results saved to: output directory
) else (
    echo   Speed test failed with error code: %EXIT_CODE%
)
echo ================================================================
echo.
pause