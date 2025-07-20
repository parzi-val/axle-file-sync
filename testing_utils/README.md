# Testing Utils

This folder contains utilities to help you test Axle file synchronization locally.

## Setup Scripts

### `setup-test-env.sh` (Linux/macOS)
Bash script that creates a complete test environment with:
- Two separate git repositories (alice_project and bob_project)  
- Redis server setup and configuration
- Environment configuration files
- Sample files for testing synchronization

**Usage:**
```bash
chmod +x setup-test-env.sh
./setup-test-env.sh
```

### `setup-test-env.ps1` (Windows)
PowerShell script that provides the same functionality as the bash script but for Windows systems.

**Usage:**
```powershell
# Run with execution policy bypass if needed
powershell -ExecutionPolicy Bypass -File setup-test-env.ps1
```

## What the Scripts Do

1. **Create Test Directories**: Sets up `alice_project` and `bob_project` as separate git repositories
2. **Initialize Git Repos**: Each project gets initialized with git and has initial commits
3. **Create Config Files**: Generates `.env` files with proper Redis and team configurations
4. **Add Sample Content**: Creates initial files that you can modify to test synchronization
5. **Setup Instructions**: Provides clear next steps for running Axle in each project

## Testing Workflow

After running the setup script:

1. **Start Redis** (if not already running):
   - Linux/macOS: `redis-server`
   - Windows: Start Redis service or run `redis-server.exe`

2. **Open two terminal sessions**:
   - Terminal 1: Navigate to `alice_project` and run `axle`
   - Terminal 2: Navigate to `bob_project` and run `axle`

3. **Test synchronization**:
   - Make changes in alice_project (edit files, create new files)
   - Watch the changes appear automatically in bob_project
   - Try the reverse: make changes in bob_project and see them in alice_project

## Requirements

- Git installed and available in PATH
- Redis server running locally
- Axle binary built and available in PATH
- Network connectivity between test instances

## Troubleshooting

- If Redis connection fails, ensure Redis server is running on default port 6379
- If synchronization isn't working, check the console logs for error messages
- Make sure both Axle instances are using the same team ID (configured in .env files)
- Ensure file permissions allow Axle to watch and modify files in the test directories
