# PowerShell script to automatically setup test environment for Axle
# Clears alice_project and bob_project folders

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

# Go back to tests directory
Set-Location $ScriptDir

Write-Host ""
Write-Host "✅ Test environment cleaned up!
" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "1. Team Lead (Alice) initializes the team:"
Write-Host "   cd testing_utils/alice_project"
Write-Host "   ../axle.exe init --team \"dev-team\" --username \"alice\""
Write-Host ""
Write-Host "2. Team Member (Bob) joins the team:"
Write-Host "   cd testing_utils/bob_project"
Write-Host "   ../axle.exe join --team \"dev-team\" --username \"bob\""
Write-Host ""
Write-Host "3. Start both clients:"
Write-Host "   - In one terminal, from alice_project: ../axle.exe start"
Write-Host "   - In another terminal, from bob_project: ../axle.exe start"
Write-Host ""