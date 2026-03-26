@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

set APP_NAME=adnx_dns
set VERSION=1.0.0
set MAIN_PKG=.\cmd\server

set ROOT=%~dp0
cd /d "%ROOT%"

set DIST_DIR=%ROOT%dist
set RELEASE_DIR=%DIST_DIR%\%APP_NAME%_linux_amd64
set BIN_NAME=%APP_NAME%
set BIN_PATH=%RELEASE_DIR%\%BIN_NAME%
set TAR_PATH=%DIST_DIR%\%APP_NAME%_linux_amd64.tar.gz
set ZIP_PATH=%DIST_DIR%\%APP_NAME%_linux_amd64.zip

echo.
echo ==========================================
echo      %APP_NAME% Linux 构建打包脚本
echo ==========================================

echo.
echo [1/7] 检查环境...
where go >nul 2>nul
if errorlevel 1 (
    echo 未找到 go，请先安装 Go。
    exit /b 1
)

where powershell >nul 2>nul
if errorlevel 1 (
    echo 未找到 powershell，无法执行压缩打包。
    exit /b 1
)

echo.
echo [2/7] 检查必要文件...
if not exist go.mod (
    echo 缺少 go.mod
    exit /b 1
)

if not exist .env.example (
    echo 缺少 .env.example
    exit /b 1
)

if not exist README.md (
    echo 缺少 README.md
    exit /b 1
)

if not exist schema.sql (
    echo 缺少 schema.sql
    exit /b 1
)

if not exist "%MAIN_PKG%" (
    echo 缺少主程序目录: %MAIN_PKG%
    exit /b 1
)

echo.
echo [3/7] 清理旧产物...
if exist "%RELEASE_DIR%" rmdir /s /q "%RELEASE_DIR%"
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"
if not exist "%RELEASE_DIR%" mkdir "%RELEASE_DIR%"
if exist "%ZIP_PATH%" del /f /q "%ZIP_PATH%"
if exist "%TAR_PATH%" del /f /q "%TAR_PATH%"

echo.
echo [4/7] 拉取依赖...
call go mod tidy
if errorlevel 1 (
    echo go mod tidy 失败
    exit /b 1
)

echo.
echo [5/7] 编译 Linux amd64...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

call go build -trimpath -ldflags "-s -w -X main.Version=%VERSION%" -o "%BIN_PATH%" %MAIN_PKG%
if errorlevel 1 (
    echo 编译失败
    exit /b 1
)

if not exist "%BIN_PATH%" (
    echo 编译失败，未生成 %BIN_PATH%
    exit /b 1
)

echo.
echo [6/7] 复制必要文件...
copy /y ".env.example" "%RELEASE_DIR%\.env.example" >nul
copy /y "README.md" "%RELEASE_DIR%\README.md" >nul
copy /y "schema.sql" "%RELEASE_DIR%\schema.sql" >nul

if exist "RELEASE.md" copy /y "RELEASE.md" "%RELEASE_DIR%\RELEASE.md" >nul

if exist "configs" (
    xcopy "configs" "%RELEASE_DIR%\configs" /e /i /y >nul
)

echo File: %BIN_NAME%> "%RELEASE_DIR%\SHA256SUMS.txt"
certutil -hashfile "%BIN_PATH%" SHA256 >> "%RELEASE_DIR%\SHA256SUMS.txt"
if errorlevel 1 (
    echo 生成 SHA256 失败
    exit /b 1
)

echo.
echo [7/7] 打包...
powershell -NoProfile -ExecutionPolicy Bypass -Command "Compress-Archive -Path '%RELEASE_DIR%\*' -DestinationPath '%ZIP_PATH%' -Force"
if errorlevel 1 (
    echo ZIP 打包失败
    exit /b 1
)

powershell -NoProfile -ExecutionPolicy Bypass -Command "tar -czf '%TAR_PATH%' -C '%DIST_DIR%' '%APP_NAME%_linux_amd64'"
if errorlevel 1 (
    echo tar.gz 打包失败
    exit /b 1
)

echo.
echo ==========================================
echo 构建完成
echo VERSION: %VERSION%
echo BIN: %BIN_PATH%
echo ZIP: %ZIP_PATH%
echo TAR.GZ: %TAR_PATH%
echo ==========================================
pause
exit /b 0