param(
    [switch]$Server,
    [switch]$Client,
    [switch]$Integration,
    [switch]$E2E,
    [switch]$All,
    [switch]$Coverage
)

$ErrorActionPreference = "Stop"

# Ensure PATH includes all tool locations
$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

$rootDir = Split-Path -Parent $PSScriptRoot
$qtDir = "C:\Qt\6.8.2\msvc2022_64"
$env:Path = "$qtDir\bin;$env:Path"

$runServer = $Server -or $All -or (-not $Client -and -not $Integration -and -not $E2E)
$runClient = $Client -or $All
$runIntegration = $Integration -or $All
$runE2E = $E2E -or $All

$failed = $false

if ($runServer) {
    Write-Host "`n=== Server Tests ===" -ForegroundColor Cyan
    $serverDir = Join-Path $rootDir "server"
    Push-Location $serverDir
    try {
        if ($Coverage) {
            go test -v -coverprofile=coverage.out -count=1 ./...
            go tool cover -html=coverage.out -o cover.html
            Write-Host "Coverage report: cover.html" -ForegroundColor Green
        } else {
            go test -v -count=1 ./...
        }
        if ($LASTEXITCODE -ne 0) { $failed = $true; Write-Host "Server tests FAILED" -ForegroundColor Red }
        else { Write-Host "Server tests PASSED" -ForegroundColor Green }
    } finally {
        Pop-Location
    }
}

if ($runClient) {
    Write-Host "`n=== Client Tests ===" -ForegroundColor Cyan
    $clientDir = Join-Path $rootDir "client"
    $buildDir = Join-Path $rootDir "build\client"
    $vsPath = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools"
    $vcvars = "$vsPath\VC\Auxiliary\Build\vcvars64.bat"

    if (-not (Test-Path $buildDir)) {
        New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
        Push-Location $clientDir
        try {
            & cmd.exe /c "`"$vcvars`" > nul 2>&1 && cmake -S . -B `"$buildDir`" -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_PREFIX_PATH=`"$qtDir`""
        } finally {
            Pop-Location
        }
    }

    Push-Location $clientDir
    try {
        & cmd.exe /c "`"$vcvars`" > nul 2>&1 && cmake --build `"$buildDir`" --target ber-client-tests --config Debug"
        if ($LASTEXITCODE -eq 0) {
            & cmd.exe /c "`"$vcvars`" > nul 2>&1 && set PATH=$env:Path && `"$buildDir\tests\ber-client-tests.exe`" --verbosity high"
            if ($LASTEXITCODE -ne 0) { $failed = $true; Write-Host "Client tests FAILED" -ForegroundColor Red }
            else { Write-Host "Client tests PASSED" -ForegroundColor Green }
        } else {
            $failed = $true; Write-Host "Client build FAILED" -ForegroundColor Red
        }
    } finally {
        Pop-Location
    }
}

if ($runIntegration) {
    Write-Host "`n=== Integration Tests ===" -ForegroundColor Cyan
    $serverDir = Join-Path $rootDir "server"
    Push-Location $serverDir
    try {
        go test -v -tags=integration -count=1 ./...
        if ($LASTEXITCODE -ne 0) { $failed = $true; Write-Host "Integration tests FAILED" -ForegroundColor Red }
        else { Write-Host "Integration tests PASSED" -ForegroundColor Green }
    } finally {
        Pop-Location
    }
}

if ($runE2E) {
    Write-Host "`n=== E2E Tests ===" -ForegroundColor Cyan
    $e2eDir = Join-Path $rootDir "tests/e2e"
    if (Test-Path $e2eDir) {
        Push-Location $e2eDir
        try {
            npx playwright test
            if ($LASTEXITCODE -ne 0) { $failed = $true; Write-Host "E2E tests FAILED" -ForegroundColor Red }
            else { Write-Host "E2E tests PASSED" -ForegroundColor Green }
        } finally {
            Pop-Location
        }
    } else {
        Write-Host "E2E tests directory not found, skipping" -ForegroundColor Yellow
    }
}

if ($failed) {
    Write-Host "`nSome tests FAILED!" -ForegroundColor Red
    exit 1
}

Write-Host "`nAll tests PASSED!" -ForegroundColor Cyan
