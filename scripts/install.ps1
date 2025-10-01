# Axle File Sync Installer for Windows
# Run with: irm https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "Installing Axle File Sync..." -ForegroundColor Cyan

# Configuration
$version = "0.1.0"
$repo = "parzi-val/axle-file-sync"
$installDir = "$env:LOCALAPPDATA\Axle"
$exeName = "axle.exe"

# Create install directory
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

# Download URL for Windows zip
$downloadUrl = "https://github.com/$repo/releases/download/v$version/axle-windows-amd64.zip"
$zipPath = Join-Path $env:TEMP "axle-windows-amd64.zip"

try {
    Write-Host "Downloading Axle v$version..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

    Write-Host "Extracting..." -ForegroundColor Yellow
    Expand-Archive -Path $zipPath -DestinationPath $installDir -Force

    $exePath = Join-Path $installDir $exeName

    # Clean up zip file
    Remove-Item $zipPath -Force

    Write-Host "Installed to $installDir" -ForegroundColor Green

    # Add to PATH if not already present
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$installDir*") {
        Write-Host "Adding Axle to PATH..." -ForegroundColor Yellow
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$currentPath;$installDir",
            "User"
        )
        Write-Host "Added to PATH. Restart your terminal to use 'axle' command." -ForegroundColor Green
    }

    Write-Host "`nAxle File Sync v$version installed successfully!" -ForegroundColor Green
    Write-Host "Run 'axle version' to verify installation" -ForegroundColor Cyan
    Write-Host "Run 'axle help' to get started" -ForegroundColor Cyan

} catch {
    Write-Host "Error: Failed to download Axle" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}