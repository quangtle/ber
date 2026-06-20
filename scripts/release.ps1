param(
    [string]$Version,
    [ValidateSet("windows-amd64", "linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64")]
    [string]$Target = "windows-amd64"
)

$ErrorActionPreference = "Stop"
$rootDir = Split-Path -Parent $PSScriptRoot

# Ensure PATH includes all tool locations
$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$env:Path = "$machinePath;$userPath"
$qtDir = "C:\Qt\6.8.2\msvc2022_64"
$env:Path = "$qtDir\bin;$env:Path"

# Detect version
if (-not $Version) {
    # ponytail: run via cmd to avoid PS 5.1 stderr-as-error wrapping
    $gitTag = (cmd /c "git describe --tags --abbrev=0 2>nul") -replace '\s+', ''
    $gitHash = (cmd /c "git rev-parse --short HEAD 2>nul") -replace '\s+', ''
    if ($gitTag -and ($gitTag -match '^v\d+\.\d+\.\d+')) {
        $Version = $gitTag.TrimStart('v')
    } else {
        $Version = "0.0.0-dev"
    }
    if ($gitHash) {
        $Version += "+$gitHash"
    }
}

Write-Host "=== ber Release Script ===" -ForegroundColor Cyan
Write-Host "Version: $Version"
Write-Host "Target:  $Target"

$releaseDir = Join-Path $rootDir "release"
$outDir = Join-Path $releaseDir "ber-$Version-$Target"

if (Test-Path $outDir) { Remove-Item -Recurse -Force $outDir }
New-Item -ItemType Directory -Path $outDir -Force | Out-Null

$parts = $Target.Split('-')
$os = $parts[0]
$arch = $parts[1]

$env:GOOS = $os
$env:GOARCH = $arch
if ($os -eq "windows") { $ext = ".exe" } else { $ext = "" }

Write-Host "`nBuilding server for $os/$arch..." -ForegroundColor Yellow
$serverDir = Join-Path $rootDir "server"
Push-Location $serverDir
try {
    $serverOut = Join-Path $outDir "ber-server$ext"
    go build -ldflags="-s -w -H=windowsgui -X main.Version=$Version" -o $serverOut .\cmd\ber-server\
    if ($LASTEXITCODE -ne 0) { throw "Server build failed" }
    Write-Host "  Server: $serverOut" -ForegroundColor Green
} finally {
    Pop-Location
}

if ($os -eq "windows") {
    Write-Host "`nBuilding client..." -ForegroundColor Yellow
    $clientDir = Join-Path $rootDir "client"
    $buildDir = Join-Path $clientDir "build-release"

    $vsPath = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools"
    $vcvars = "$vsPath\VC\Auxiliary\Build\vcvars64.bat"

    Push-Location $clientDir
    try {
        & cmd.exe /c "`"$vcvars`" > nul 2>&1 && cmake -S . -B `"$buildDir`" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=`"$qtDir`" && cmake --build `"$buildDir`" --config Release && cmake --install `"$buildDir`" --prefix `"$outDir`""
        if ($LASTEXITCODE -ne 0) { throw "Client build failed" }
        # windeployqt runs at post-build — copy only what the app actually needs
        Copy-Item "$buildDir/*.dll" $outDir -Force -ErrorAction SilentlyContinue
        Remove-Item (Join-Path $outDir "vc_redist.x64.exe") -Force -ErrorAction SilentlyContinue
        Remove-Item (Join-Path $outDir "Qt6Svg.dll") -Force -ErrorAction SilentlyContinue
        @('multimedia', 'platforms', 'styles') | ForEach-Object {
            $src = Join-Path $buildDir $_
            if (Test-Path $src) { Copy-Item $src (Join-Path $outDir $_) -Recurse -Force }
        }
        Write-Host "  Client: $outDir" -ForegroundColor Green
    } finally {
        Pop-Location
    }
} else {
    Write-Host "Client build not yet supported for $os, skipping." -ForegroundColor Yellow
}

Copy-Item (Join-Path $rootDir "README.md") (Join-Path $outDir "README.md") -Force

Write-Host "`nCreating archive..." -ForegroundColor Yellow
Push-Location $releaseDir
try {
    $archiveName = "ber-$Version-$Target"
    if ($os -eq "windows") {
        $archive = "$archiveName.zip"
        Compress-Archive -Path $archiveName -DestinationPath $archive -Force
    } else {
        $archive = "$archiveName.tar.gz"
        & tar czf $archive $archiveName
    }
    Write-Host "Archive: $releaseDir/$archive" -ForegroundColor Green
} finally {
    Pop-Location
}

Write-Host "`nRelease complete!" -ForegroundColor Cyan
Write-Host "Output: $outDir"
