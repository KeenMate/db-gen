# db-gen Build Script for PowerShell
# Builds db-gen for Windows and Linux

param(
    [string]$Target = "all"
)

$BinaryName = "db-gen"
$BuildDir = "build"
$BuildFlags = "-trimpath"
$LdFlags = "-s -w"

function Build-Windows {
    Write-Host "Building for Windows (amd64)..." -ForegroundColor Cyan
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir | Out-Null
    }

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build $BuildFlags -ldflags $LdFlags -o "$BuildDir\$BinaryName-win.exe" .

    if ($LASTEXITCODE -eq 0) {
        Write-Host "Windows build complete: $BuildDir\$BinaryName-win.exe" -ForegroundColor Green
    } else {
        Write-Host "Windows build failed" -ForegroundColor Red
        exit 1
    }
}

function Build-Linux {
    Write-Host "Building for Linux (amd64)..." -ForegroundColor Cyan
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir | Out-Null
    }

    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build $BuildFlags -ldflags $LdFlags -o "$BuildDir\$BinaryName-linux" .

    if ($LASTEXITCODE -eq 0) {
        Write-Host "Linux build complete: $BuildDir\$BinaryName-linux" -ForegroundColor Green
    } else {
        Write-Host "Linux build failed" -ForegroundColor Red
        exit 1
    }
}

function Clean-Build {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
    if (Test-Path $BuildDir) {
        Remove-Item -Path $BuildDir -Recurse -Force
    }
    if (Test-Path "$BinaryName.exe") {
        Remove-Item -Path "$BinaryName.exe" -Force
    }
    Write-Host "Clean complete" -ForegroundColor Green
}

# Main execution
switch ($Target.ToLower()) {
    "windows" {
        Build-Windows
    }
    "linux" {
        Build-Linux
    }
    "clean" {
        Clean-Build
    }
    "all" {
        Clean-Build
        Build-Windows
        Build-Linux
    }
    default {
        Write-Host "Usage: .\build.ps1 [all|windows|linux|clean]" -ForegroundColor Yellow
        Write-Host "  all     - Build for Windows and Linux (default)"
        Write-Host "  windows - Build for Windows only"
        Write-Host "  linux   - Build for Linux only"
        Write-Host "  clean   - Remove build artifacts"
        exit 1
    }
}

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
