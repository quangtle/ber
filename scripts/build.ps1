param(
    [switch]$ServerOnly,
    [switch]$ClientOnly,
    [ValidateSet("Debug", "Release")]
    [string]$Config = "Release"
)

$ErrorActionPreference = "Stop"

$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"

$rootDir = Split-Path -Parent $PSScriptRoot
$buildDir = Join-Path $rootDir "build"
$qtDir = "C:\Qt\6.8.2\msvc2022_64"
$env:Path = "$qtDir\bin;$env:Path"

if (-not (Test-Path $buildDir)) {
    New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
}

Write-Host "=== ber Build Script ===" -ForegroundColor Cyan

if (-not $ClientOnly) {
    Write-Host "`nBuilding server..." -ForegroundColor Yellow
    $serverOut = Join-Path $buildDir "ber-server.exe"
    Push-Location (Join-Path $rootDir "server")
    try {
        go build -ldflags="-s -w -H=windowsgui" -o $serverOut .\cmd\ber-server\
        if ($LASTEXITCODE -ne 0) { throw "Server build failed" }
        Write-Host "  $serverOut" -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

if (-not $ServerOnly) {
    Write-Host "`nBuilding client..." -ForegroundColor Yellow
    $clientDir = Join-Path $rootDir "client"
    $clientBuildDir = Join-Path $buildDir "client"
    $vsPath = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools"
    $vcvars = "$vsPath\VC\Auxiliary\Build\vcvars64.bat"

    Push-Location $clientDir
    try {
        & cmd.exe /c "`"$vcvars`" > nul 2>&1 && cmake -S . -B `"$clientBuildDir`" -G Ninja -DCMAKE_BUILD_TYPE=$Config -DCMAKE_PREFIX_PATH=`"$qtDir`" && cmake --build `"$clientBuildDir`" --config $Config"
        if ($LASTEXITCODE -ne 0) { throw "Client build failed" }
        Write-Host "  $(Join-Path $clientBuildDir 'ber-client.exe')" -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

Write-Host "`nBuild complete!" -ForegroundColor Cyan
