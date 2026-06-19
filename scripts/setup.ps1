param(
    [switch]$ServerOnly,
    [switch]$ClientOnly
)

$ErrorActionPreference = "Stop"

$installDir = Split-Path -Parent $PSScriptRoot

Write-Host "=== ber Development Environment Setup ===" -ForegroundColor Cyan
Write-Host "Target: Windows (amd64)"
Write-Host ""

function Install-Tool {
    param([string]$Name, [string]$WingetId)
    Write-Host "Installing $Name..." -ForegroundColor Yellow
    & winget install --id $WingetId --exact --accept-source-agreements --accept-package-agreements --source winget --disable-interactivity 2>&1
    if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne -1978335189) {
        Write-Host "  $Name install had issues (may already be installed)" -ForegroundColor Yellow
    } else {
        Write-Host "  $Name installed" -ForegroundColor Green
    }
}

# --- Common tools ---
if (-not $ClientOnly) {
    Install-Tool "Go" "GoLang.Go"
    Install-Tool "FFmpeg" "Gyan.FFmpeg"
}

if (-not $ServerOnly) {
    Install-Tool "CMake" "Kitware.CMake"
    Install-Tool "Ninja" "Ninja-build.Ninja"

    # Visual Studio Build Tools with C++ workload
    Write-Host ""
    Write-Host "Checking Visual Studio 2022 Build Tools..." -ForegroundColor Yellow
    $vsPath = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools"
    $vcvars = "$vsPath\VC\Auxiliary\Build\vcvars64.bat"
    if (-not (Test-Path $vcvars)) {
        Write-Host "Installing Visual Studio 2022 Build Tools with C++ workload..." -ForegroundColor Yellow
        Write-Host "  (This downloads several GB — will take 10-30 minutes)" -ForegroundColor Yellow

        # Download bootstrapper
        $bootstrapper = "$env:TEMP\vs_BuildTools.exe"
        if (-not (Test-Path $bootstrapper)) {
            Write-Host "  Downloading bootstrapper..." -ForegroundColor Yellow
            Invoke-WebRequest -Uri "https://aka.ms/vs/17/release/vs_BuildTools.exe" -OutFile $bootstrapper
        }

        # Install C++ workload
        $proc = Start-Process -Wait -PassThru -NoNewWindow -FilePath $bootstrapper -ArgumentList "--installPath `"$vsPath`" --add Microsoft.VisualStudio.Workload.VCTools --add Microsoft.VisualStudio.Component.VC.Tools.x86.x64 --add Microsoft.VisualStudio.Component.Windows11SDK.22621 --quiet --norestart --includeRecommended --wait"
        if ($proc.ExitCode -eq 0) {
            Write-Host "  Visual Studio Build Tools installed" -ForegroundColor Green
        } else {
            Write-Host "  Visual Studio install exit code: $($proc.ExitCode)" -ForegroundColor Yellow
            Write-Host "  (You may need to reboot and re-run, or install manually)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  Visual Studio Build Tools already installed" -ForegroundColor Green
    }

    # Qt6
    Write-Host ""
    Write-Host "Installing Qt6..." -ForegroundColor Yellow
    $qtDir = "C:\Qt\6.8.2\msvc2022_64"
    if (-not (Test-Path "$qtDir\bin\qmake6.exe")) {
        # Install aqtinstall
        $python = Get-Command python -ErrorAction SilentlyContinue
        if (-not $python) {
            Install-Tool "Python" "Python.Python.3.12"
            $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
        }
        pip install aqtinstall 2>&1 | Out-Null
        python -m aqt install-qt windows desktop 6.8.2 win64_msvc2022_64 --modules qtmultimedia --outputdir C:\Qt 2>&1
        Write-Host "  Qt6 installed at: $qtDir" -ForegroundColor Green
    } else {
        Write-Host "  Qt6 already installed" -ForegroundColor Green
    }
}

# --- Refresh PATH ---
$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
$qtDir = "C:\Qt\6.8.2\msvc2022_64"
$env:Path = "$qtDir\bin;$env:Path"

# --- Verify ---
Write-Host ""
Write-Host "=== Verification ===" -ForegroundColor Cyan
$tools = @(
    @{Name="Go"; Cmd="go version"}
    @{Name="CMake"; Cmd="cmake --version"}
    @{Name="Ninja"; Cmd="ninja --version"}
    @{Name="FFmpeg"; Cmd="ffmpeg -version"}
    @{Name="MSVC cl"; Cmd="cmd /c `"`"$vcvars`" > nul 2>&1 && cl 2>&1 | findstr `"Optimizing`""}
)
foreach ($t in $tools) {
    $result = & cmd /c $($t.Cmd) 2>&1 | Out-String
    if ($result -match "not recognized|not found|NOT FOUND|error") {
        Write-Host "  $($t.Name): NOT FOUND" -ForegroundColor Red
    } else {
        Write-Host "  $($t.Name): OK" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Cyan
Write-Host "Next steps:"
Write-Host "  1. .\scripts\build.ps1"
Write-Host "  2. .\scripts\test.ps1 -All"
Write-Host "  3. .\scripts\release.ps1"
