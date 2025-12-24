@echo off
chcp 65001 >nul
title 编译 Clash Speedtest

echo.
echo ========================================
echo   编译 Clash Speedtest
echo ========================================
echo.

echo [1/2] 整理依赖...
go mod tidy
if %errorlevel% neq 0 (
    echo [错误] 依赖整理失败
    pause
    exit /b 1
)

echo [2/2] 编译程序...
go build -ldflags "-s -w" -o clash-speedtest.exe ./cmd
if %errorlevel% neq 0 (
    echo [错误] 编译失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo   ✅ 编译成功！
echo ========================================
echo.
echo 可执行文件: clash-speedtest.exe
echo.
echo 使用方法:
echo   1. CLI 模式: clash-speedtest.exe -h
echo   2. Web UI:   双击 start-web.bat
echo.

pause
