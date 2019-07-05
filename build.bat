@echo off

if exist build.bat goto ok
echo install.bat must be run from its folder
goto end

: ok

SET dist=go_win_amd64

SET GOPATH=%~dp0;%GOPATH%

go build -o bin/%dist%.exe main

:end
echo finished