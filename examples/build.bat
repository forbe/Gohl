@echo off
rsrc -manifest app.manifest -ico icon.ico -o rsrc.syso
@REM go build -o demo.exe -ldflags "-s -w -H windowsgui"
go build -o demo.exe -ldflags "-s -w"