@echo off
echo Mengompilasi untuk semua OS...

:: Windows x64
set GOOS=windows
set GOARCH=amd64
go build -o bin/GoScRcPy_Win_x64.exe main.go

:: Windows x86 (32-bit)
set GOOS=windows
set GOARCH=386
go build -o bin/GoScRcPy_Win_x86.exe main.go

:: Linux x64
set GOOS=linux
set GOARCH=amd64
go build -o bin/GoScRcPy_Linux_x64 main.go

:: macOS Intel
set GOOS=darwin
set GOARCH=amd64
go build -o bin/GoScRcPy_Mac_Intel main.go

:: macOS M1/M2/M3 (ARM)
set GOOS=darwin
set GOARCH=arm64
go build -o bin/GoScRcPy_Mac_M_Series main.go

echo Selesai! Cek folder /bin
pause