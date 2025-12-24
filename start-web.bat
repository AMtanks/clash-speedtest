@echo off
chcp 65001 >nul
title Clash Speedtest Web UI

echo.
echo ========================================
echo   Clash Speedtest Web UI 启动器
echo ========================================
echo.

REM 检查是否存在可执行文件
if not exist "clash-speedtest.exe" (
    echo [错误] 找不到 clash-speedtest.exe
    echo 请先编译项目: go build -o clash-speedtest.exe ./cmd
    echo.
    pause
    exit /b 1
)

REM 默认端口
set PORT=8080

REM 检查端口是否被占用
netstat -ano | findstr ":%PORT%" >nul 2>&1
if %errorlevel% equ 0 (
    echo [警告] 端口 %PORT% 已被占用，尝试使用端口 8081...
    set PORT=8081
    
    REM 再次检查 8081
    netstat -ano | findstr ":%PORT%" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [警告] 端口 %PORT% 也被占用，尝试使用端口 8082...
        set PORT=8082
    )
)

echo [信息] 启动 Web 服务器...
echo [信息] 端口: %PORT%
echo [信息] 访问地址: http://localhost:%PORT%
echo.

REM 延迟 2 秒后自动打开浏览器
start "" cmd /c "timeout /t 2 /nobreak >nul & start http://localhost:%PORT%"

echo ========================================
echo   浏览器将在 2 秒后自动打开
echo   按 Ctrl+C 停止服务器
echo ========================================
echo.

REM 启动 Web 服务器
clash-speedtest.exe web --port %PORT%

if %errorlevel% neq 0 (
    echo.
    echo [错误] Web 服务器启动失败
    echo.
    pause
    exit /b 1
)

pause
