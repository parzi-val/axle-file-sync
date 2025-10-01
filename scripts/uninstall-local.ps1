# Axle File Sync Local Uninstaller for Windows
# This script is included with the installation

param(
    [switch]$Silent = $false
)

$ErrorActionPreference = "Stop"

if (-not $Silent) {
    Write-Host "=====================================" -ForegroundColor Yellow
    Write-Host "   Axle File Sync Uninstaller" -ForegroundColor Yellow
    Write-Host "=====================================" -ForegroundColor Yellow
    Write-Host ""

    $confirmation = Read-Host "Are you sure you want to uninstall Axle? (y/n)"
    if ($confirmation -ne 'y') {
        Write-Host "Uninstall cancelled." -ForegroundColor Cyan
        exit 0
    }
}

Write-Host "Uninstalling Axle File Sync..." -ForegroundColor Yellow

# Get current directory (should be the install directory)
$installDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Remove from PATH
Write-Host "Removing from PATH..." -ForegroundColor Yellow
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -like "*$installDir*") {
    $newPath = ($currentPath.Split(';') | Where-Object { $_ -ne $installDir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Removed from PATH" -ForegroundColor Green
}

# Remove Windows registry entry
$uninstallKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Axle"
if (Test-Path $uninstallKey) {
    Write-Host "Removing registry entry..." -ForegroundColor Yellow
    Remove-Item $uninstallKey -Force
    Write-Host "Registry entry removed" -ForegroundColor Green
}

# Schedule directory removal (can't delete while script is running from it)
$tempScript = Join-Path $env:TEMP "axle-uninstall-cleanup.ps1"
@"
Start-Sleep -Seconds 2
Remove-Item -Path '$installDir' -Recurse -Force
Remove-Item -Path '$tempScript' -Force
"@ | Out-File -FilePath $tempScript -Encoding UTF8

Write-Host "Cleaning up installation files..." -ForegroundColor Yellow

# Start the cleanup script
Start-Process powershell -ArgumentList "-WindowStyle Hidden -ExecutionPolicy Bypass -File `"$tempScript`"" -WindowStyle Hidden

Write-Host ""
Write-Host "Axle File Sync has been uninstalled successfully!" -ForegroundColor Green
Write-Host "The installation directory will be removed in a few seconds." -ForegroundColor Cyan
Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Cyan