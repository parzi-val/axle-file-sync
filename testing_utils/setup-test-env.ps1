# PowerShell script to automatically setup test environment for Axle
# Clears alice_project and bob_project folders and initializes them with Axle

param()

# Set error action preference to stop on errors
$ErrorActionPreference = "Stop"

Write-Host "Setting up Axle test environment..." -ForegroundColor Green

# Get the directory where this script is located
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Define project directories
$AliceDir = "alice_project"
$BobDir = "bob_project"
$AxleExe = ".\axle.exe"

# Check if axle.exe exists
if (-not (Test-Path $AxleExe)) {
    Write-Host "Error: axle.exe not found in tests directory" -ForegroundColor Red
    Write-Host "Please copy axle.exe to the tests folder first" -ForegroundColor Red
    exit 1
}

Write-Host "Cleaning up existing directories..." -ForegroundColor Yellow

# Remove existing directories if they exist
if (Test-Path $AliceDir) {
    Write-Host "Removing existing $AliceDir directory..."
    Remove-Item -Recurse -Force $AliceDir
}

if (Test-Path $BobDir) {
    Write-Host "Removing existing $BobDir directory..."
    Remove-Item -Recurse -Force $BobDir
}

# Create fresh directories
Write-Host "Creating fresh directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $AliceDir | Out-Null
New-Item -ItemType Directory -Force -Path $BobDir | Out-Null

# Initialize Alice's project
Write-Host "Initializing Alice's project..." -ForegroundColor Cyan
Set-Location $AliceDir
& "..\axle.exe" init --team "dev-team" --username "alice"

Write-Host "Alice's project initialized successfully!" -ForegroundColor Green

# Initialize Bob's project  
Write-Host "Initializing Bob's project..." -ForegroundColor Cyan
Set-Location "..\$BobDir"
& "..\axle.exe" init --team "dev-team" --username "bob"

Write-Host "Bob's project initialized successfully!" -ForegroundColor Green

# Go back to tests directory
Set-Location $ScriptDir

Write-Host ""
Write-Host "✅ Test environment setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "1. Start Alice: cd alice_project && ..\axle.exe start"
Write-Host "2. Start Bob: cd bob_project && ..\axle.exe start"
Write-Host ""
Write-Host "Directory structure:" -ForegroundColor White
Write-Host "tests/"
Write-Host "  |-- axle.exe"
Write-Host "  |-- setup-test-env.ps1 (this script)"
Write-Host "  |-- setup-test-env.sh (bash version)"
Write-Host "  |-- alice_project/"
Write-Host "  |   |-- .git/"
Write-Host "  |   '-- axle_config.json"
Write-Host "  '-- bob_project/"
Write-Host "      |-- .git/"
Write-Host "      '-- axle_config.json"
Write-Host ""
