# Axle File Sync Uninstaller for Windows
# Run with: irm https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "Uninstalling Axle File Sync..." -ForegroundColor Yellow

# Configuration
$installDir = "$env:LOCALAPPDATA\Axle"
$exeName = "axle.exe"

# Check if Axle is installed
if (!(Test-Path $installDir)) {
    Write-Host "Axle is not installed at $installDir" -ForegroundColor Red
    exit 1
}

# Remove from PATH
Write-Host "Removing from PATH..." -ForegroundColor Yellow
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -like "*$installDir*") {
    $newPath = ($currentPath.Split(';') | Where-Object { $_ -ne $installDir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Removed from PATH" -ForegroundColor Green
}

# Remove Windows registry entry (if exists)
$uninstallKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Axle"
if (Test-Path $uninstallKey) {
    Write-Host "Removing registry entry..." -ForegroundColor Yellow
    Remove-Item $uninstallKey -Force
    Write-Host "Registry entry removed" -ForegroundColor Green
}

# Remove installation directory
Write-Host "Removing Axle directory..." -ForegroundColor Yellow
Remove-Item -Path $installDir -Recurse -Force
Write-Host "Directory removed" -ForegroundColor Green

Write-Host "`nAxle File Sync has been uninstalled successfully!" -ForegroundColor Green
Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Cyan