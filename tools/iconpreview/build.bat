@echo off
rem Build the iconpreview tool (run from this directory).
rem rsrc must be on PATH: go install github.com/akavel/rsrc@latest
rem The manifest is what makes glyphs crisp (DPI PerMonitorV2) - always regenerate the syso.
rsrc -manifest app.manifest -o rsrc.syso
set CGO_ENABLED=1
go build -ldflags="-s -w -H windowsgui" -o iconpreview.exe
if errorlevel 1 (
  echo BUILD FAILED
  pause
  exit /b 1
)
echo BUILT iconpreview.exe
