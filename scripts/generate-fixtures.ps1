param(
    [string]$OutputDir
)

$ErrorActionPreference = "Stop"

if (-not $OutputDir) {
    $rootDir = Split-Path -Parent $PSScriptRoot
    $OutputDir = Join-Path $rootDir "tests\fixtures"
}

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "Generating test fixtures in $OutputDir" -ForegroundColor Cyan

$ffmpeg = Get-Command ffmpeg -ErrorAction SilentlyContinue
if (-not $ffmpeg) {
    Write-Host "FFmpeg not found. Install FFmpeg and ensure it's on PATH." -ForegroundColor Red
    exit 1
}

# Generate test videos (1 second each, small resolution)
$fixtures = @(
    @{ Name = "sample_h264.mp4"; Args = "-t 1 -f lavfi -i testsrc2=size=640x480:rate=1 -c:v libx264 -preset ultrafast -pix_fmt yuv420p" },
    @{ Name = "sample_h265.mkv"; Args = "-t 1 -f lavfi -i testsrc2=size=640x480:rate=1 -c:v libx265 -preset ultrafast -pix_fmt yuv420p" },
    @{ Name = "sample_av1.webm"; Args = "-t 1 -f lavfi -i testsrc2=size=640x480:rate=1 -c:v libaom-av1 -crf 63 -strict experimental" },
    @{ Name = "test_audio.mp3"; Args = "-t 1 -f lavfi -i sine=frequency=440:duration=1 -codec:a libmp3lame -b:a 128k" }
)

foreach ($f in $fixtures) {
    $output = Join-Path $OutputDir $f.Name
    if (Test-Path $output) {
        Write-Host "  Skipping $($f.Name) (exists)" -ForegroundColor Yellow
        continue
    }
    Write-Host "  Generating $($f.Name)..." -ForegroundColor Yellow
    $cmd = "$($f.Args) -y `"$output`""
    Start-Process -Wait -NoNewWindow -FilePath ffmpeg -ArgumentList "-loglevel quiet $cmd"
}

Write-Host "Done! Generated $(($fixtures | Where-Object { -not (Test-Path (Join-Path $OutputDir $_.Name)) }).Count) new files." -ForegroundColor Green
