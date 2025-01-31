@echo off
set OUTPUT_DIR=build

if not exist %OUTPUT_DIR% (
	mkdir %OUTPUT_DIR%
)

set GOOS=linux
set GOARCH=amd64
go build -o build/ysptp_linux_amd64

set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -o build/ysptp_linux_armv7

set GOOS=linux
set GOARCH=arm64
go build -o build/ysptp_linux_arm64

set GOOS=windows
set GOARCH=amd64
go build -o build/ysptp_windows_amd64.exe

echo Build completed!
pause