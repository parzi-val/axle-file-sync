# Axle Commands Reference

## Overview
Axle is a real-time file synchronization tool designed for hackathon teams and rapid prototyping groups. It uses Git for version control and Redis for real-time communication.

## Installation
```bash
go build -o axle.exe
```

## Core Commands

### `axle init`
Initialize a new Axle team and create the first instance.

```bash
axle init --team <team-id> --username <username> [options]
```

**Required Flags:**
- `--team` - Unique team identifier
- `--username` - Your username for this instance

**Optional Flags:**
- `--password` - Team password (will prompt if not provided)
- `--host` - Redis server host (default: localhost)
- `--port` - Redis server port (default: 6379)

**Example:**
```bash
axle init --team hackathon-2024 --username alice --password secret123
```

---

### `axle join`
Join an existing Axle team.

```bash
axle join --team <team-id> --username <username> [options]
```

**Required Flags:**
- `--team` - Team identifier to join
- `--username` - Your username for this instance

**Optional Flags:**
- `--password` - Team password (will prompt if not provided)
- `--host` - Redis server host (default: localhost)
- `--port` - Redis server port (default: 6379)

**Example:**
```bash
axle join --team hackathon-2024 --username bob --password secret123
```

---

### `axle start`
Start the Axle daemon for file synchronization.

```bash
axle start [options]
```

**Optional Flags:**
- `--conflict` - Conflict resolution strategy (default: merge)
  - `theirs` - Always accept incoming changes
  - `mine` - Always keep local changes
  - `merge` - Create merge conflict markers (recommended)
  - `backup` - Create .backup files before applying changes
  - `interactive` - Open conflicts in IDE (VS Code)

**Examples:**
```bash
axle start                    # Use default merge strategy
axle start --conflict theirs  # Always accept remote changes
axle start --conflict merge   # Create conflict markers for manual resolution
```

**Notes:**
- Press `Ctrl+C` to stop the daemon gracefully
- The daemon will automatically batch file changes for efficiency
- Monitors all files except those in .gitignore and .git directory

---

### `axle chat`
Send a message to your team.

```bash
axle chat <message>
```

**Example:**
```bash
axle chat "Just pushed the new API endpoints!"
```

---

### `axle team`
Display team information and member presence.

```bash
axle team
```

**Output includes:**
- Team ID
- Team members and their status (online/offline)
- Last seen timestamps for offline members
- IP addresses of connected nodes

---

### `axle stats`
Display comprehensive synchronization statistics.

```bash
axle stats
```

**Output includes:**
- Git repository statistics
  - Total commits
  - Current branch
  - Recent commit history
- File statistics
  - Total files tracked
  - Total repository size
  - Largest files
  - Recently modified files
- Team presence information
- Sync activity summary

---

### `axle ignore`
Manage ignored file patterns.

```bash
axle ignore <action> [pattern]
```

**Actions:**
- `add <pattern>` - Add a pattern to ignore list
- `remove <pattern>` - Remove a pattern from ignore list
- `list` - Show all ignored patterns
- `auto` - Auto-detect and configure based on project type

**Examples:**
```bash
axle ignore add "*.log"        # Ignore all log files
axle ignore add "node_modules" # Ignore node_modules directory
axle ignore list               # Show all ignored patterns
axle ignore auto               # Auto-configure for detected stack
```

**Auto-detection supports:**
- Node.js/JavaScript projects
- Python projects
- Go projects
- Rust projects
- Java projects
- Ruby projects
- .NET projects

---

### `axle help`
Display help information.

```bash
axle help [command]
```

**Examples:**
```bash
axle help        # Show general help
axle help start  # Show help for start command
```

---

## Environment Variables

You can set these environment variables instead of using flags:

- `AXLE_REDIS_HOST` - Redis server host
- `AXLE_REDIS_PORT` - Redis server port
- `AXLE_TEAM_ID` - Default team ID
- `AXLE_USERNAME` - Default username

---

## Configuration File

Axle creates an `axle_config.json` file in your project root with the following structure:

```json
{
  "team_id": "hackathon-2024",
  "username": "alice",
  "root_dir": "/path/to/project",
  "redis_host": "localhost",
  "redis_port": 6379,
  "ignore_patterns": [".git", "axle_config.json", "*.log"],
  "conflict_strategy": "merge"
}
```

This file is automatically added to `.git/info/exclude` to prevent it from being committed.

---

## Conflict Resolution Strategies

### `theirs` Strategy
- Automatically accepts all incoming changes
- Overwrites local changes with remote version
- Best for: Team members who primarily consume updates

### `mine` Strategy
- Keeps all local changes
- Ignores incoming patches that conflict
- Best for: Team leads making authoritative changes

### `merge` Strategy (Recommended)
- Attempts automatic merging when possible
- Creates Git-style conflict markers when automatic merge fails
- Integrates with VS Code's merge conflict UI
- Best for: Most team members who need visibility into conflicts

### `backup` Strategy
- Creates `.backup` files before applying changes
- Applies incoming changes after backing up current version
- Best for: Cautious users who want to preserve all versions

### `interactive` Strategy
- Opens conflicts in VS Code for manual resolution
- Similar to merge but actively opens the IDE
- Best for: Active development with immediate conflict resolution

---

## Workflow Examples

### Starting a Hackathon Project

**Team Lead (Alice):**
```bash
cd my-hackathon-project
axle init --team hackathon-2024 --username alice
axle start --conflict merge
```

**Team Members (Bob, Carol):**
```bash
cd my-hackathon-project
axle join --team hackathon-2024 --username bob
axle start --conflict merge
```

### Handling Conflicts

When conflicts occur with `--conflict merge`:
1. Axle creates standard Git conflict markers in files
2. VS Code automatically detects and highlights conflicts
3. Use VS Code's merge conflict UI to resolve
4. Save the file and Axle will sync the resolution

### Checking Team Status
```bash
axle team     # See who's online
axle stats    # See sync statistics
axle chat "Need help with the API integration!"
```

---

## Troubleshooting

### Common Issues

**"Failed to connect to Redis"**
- Ensure Redis server is running: `redis-server`
- Check host and port settings
- Verify network connectivity

**"Patch failed to apply"**
- Check for uncommitted changes: `git status`
- Try different conflict strategy: `axle start --conflict theirs`
- Ensure you have the latest version with independent repo support

**"File size exceeds limit"**
- Default limit is 10MB per file
- Add large files to ignore list: `axle ignore add "*.mp4"`
- Binary files are automatically skipped

**Performance Issues**
- Axle uses dynamic batching (1-5 seconds based on activity)
- High-activity periods automatically extend batch window
- Check ignored patterns to exclude unnecessary files

---

## Tips for Hackathon Teams

1. **Start with shared understanding**: Have all team members use `--conflict merge` initially
2. **Designate authoritative sources**: Team lead can use `--conflict mine` for critical files
3. **Use chat actively**: `axle chat` helps coordinate changes
4. **Monitor presence**: `axle team` shows who's actively working
5. **Auto-configure ignores**: Run `axle ignore auto` to detect your stack
6. **Graceful shutdown**: Always use Ctrl+C to ensure pending changes are synced

---

## System Requirements

- Git (must be in PATH)
- Redis server (local or remote)
- Go 1.20+ (for building from source)
- Supported OS: Windows, macOS, Linux

---

## Security Considerations

- Team passwords are hashed with bcrypt before storage
- Patches are validated to prevent path traversal attacks
- Each node gets a unique ID for presence tracking
- Redis channels are namespaced by team ID
- Local config files are excluded from Git

---

## Version
Current version: 0.1.0 (Beta)