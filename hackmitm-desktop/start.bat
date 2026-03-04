@echo off
chcp 65001 >nul

echo 🚀 HackMITM 快速启动脚本
echo =========================
echo.

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ⚠️  Go语言环境未安装
    echo 请访问 https://golang.org/dl/ 下载安装Go
    echo.
    echo 📋 快速启动 (无插件模式):
    echo hackmitm.exe -config configs/config-no-plugins.json
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set go_version=%%i
echo ✅ 检测到Go环境: %go_version%
echo.

echo 请选择启动模式:
echo 1) 快速启动 (无插件，推荐首次使用)
echo 2) 完整功能 (包含插件，需要构建)
echo.
set /p choice=请输入选择 [1-2]: 

if "%choice%"=="1" (
    echo.
    echo 🚀 启动基础版本...
    hackmitm.exe -config configs/config-no-plugins.json
) else if "%choice%"=="2" (
    echo.
    echo 🔧 构建插件...
    if exist "plugins/Makefile" (
        cd plugins
        make examples
        cd ..
        echo ✅ 插件构建完成
    ) else (
        echo ❌ 插件Makefile不存在
        pause
        exit /b 1
    )
    
    echo.
    echo 🚀 启动完整版本...
    hackmitm.exe -config configs/config.json
) else (
    echo ❌ 无效选择
    pause
    exit /b 1
) 