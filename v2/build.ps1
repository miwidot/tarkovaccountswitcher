# Build script for Tarkov Account Switcher v2 (Go)
#
# Prerequisites:
#   1. Go 1.21+ installed
#   2. GCC/MinGW-w64 installed (required for Fyne)
#      - Download from: https://github.com/niXman/mingw-builds-binaries/releases
#      - Extract to C:\mingw64 (or similar)
#      - Add to PATH: $env:PATH += ";C:\mingw64\bin"
#   3. Set CGO_ENABLED=1

Write-Host "Building Tarkov Account Switcher v2..." -ForegroundColor Cyan

# Enable CGO
$env:CGO_ENABLED = "1"

# Check for gcc
$gcc = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $gcc) {
    Write-Host "ERROR: GCC not found in PATH!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install MinGW-w64:" -ForegroundColor Yellow
    Write-Host "  1. Download from: https://github.com/niXman/mingw-builds-binaries/releases"
    Write-Host "  2. Extract to C:\mingw64"
    Write-Host "  3. Add to PATH: `$env:PATH += `";C:\mingw64\bin`""
    exit 1
}

Write-Host "Using GCC: $($gcc.Source)" -ForegroundColor Green

# Build with Windows GUI flags and strip symbols
go build -ldflags="-H windowsgui -s -w" -o "Tarkov Account Switcher.exe"

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item "Tarkov Account Switcher.exe").Length / 1MB
    Write-Host ""
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Output: Tarkov Account Switcher.exe ($([math]::Round($size, 2)) MB)"
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
