# Axle

Axle is a real-time collaborative file synchronization system that enables teams to work together on shared directories. It uses Git for version control and patch generation, combined with Redis pub/sub for real-time communication between team members.

## Features

- **Real-time file synchronization** across team members
- **Git-based version control** with automatic commit and patch generation
- **Redis pub/sub messaging** for instant change notifications
- **Batch processing** to efficiently handle multiple file changes
- **File system watching** with debouncing to prevent duplicate events
- **Team-based collaboration** with configurable team channels
- **Selective file ignoring** with customizable ignore patterns

## Prerequisites

- **Go 1.24.0** or higher
- **Git** installed and available in PATH
- **Redis server** running and accessible

## Installation

1. Clone the repository:

```bash
git clone https://github.com/parzi-val/axle-file-sync
cd axle-file-sync
```

2. Install dependencies:

```bash
go mod tidy
```

3. Build the project:

```bash
go build -o axle.exe
```

## Usage

1. **Start Redis server** (if not already running):

```bash
redis-server
```

2. **Initialize a new team**:

```bash
./axle.exe init --team <team-name> --username <username>
```

You will be prompted to create a password for the team. Alternatively, you can provide it with the `--password` flag.

3. **Have team members join**:

Other team members can join by running:
```bash
./axle.exe join --team <team-name> --username <other-username>
```
They will be prompted for the team password.

4. **Start Axle Daemon**:

```bash
./axle.exe start
```

5. **Send chat messages** to your team (optional):

```bash
./axle.exe chat "Your message here"
```

## How It Works

1. **File Monitoring**: Axle watches your specified directory for file changes using filesystem events
2. **Git Integration**: Changes are automatically committed to a local Git repository with descriptive messages
3. **Patch Generation**: Git patches are created for each commit to capture the exact changes
4. **Redis Publishing**: Change metadata and patches are published to team-specific Redis channels
5. **Real-time Sync**: Other team members receive and apply patches automatically to stay in sync

## Configuration

Axle supports ignoring specific files and directories through configurable patterns. Common patterns include:

- `.git` (automatically ignored)
- `node_modules`
- `*.log`
- `temp/`

## Team Collaboration

Multiple users can collaborate on the same project by:

1. Using the same **Team ID**
2. Connecting to the same **Redis server**
3. Working in directories with the same **Git repository structure**

Changes made by one team member are automatically synchronized to all other members in real-time.
