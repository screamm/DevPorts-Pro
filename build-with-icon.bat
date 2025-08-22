@echo off
echo Building DevPorts Pro with Windows icon...

REM Create icon if not exists
if not exist icon.ico (
    echo Creating icon files...
    go run create_icon.go
    go run convert_to_ico.go
)

REM Compile Windows resource if not exists
if not exist app.syso (
    echo Compiling Windows resources...
    windres -o app.syso app.rc
)

REM Build the executable
echo Building executable...
go build -ldflags "-H=windowsgui -w -s" -o devports-pro.exe

echo.
echo Build complete! 
echo The devports-pro.exe now includes the Windows icon.
echo.
echo If the icon doesn't show immediately in Windows Explorer:
echo 1. Delete the old devports-pro.exe
echo 2. Clear Windows icon cache (run: ie4uinit.exe -show)
echo 3. Rebuild and restart Windows Explorer
echo.
pause