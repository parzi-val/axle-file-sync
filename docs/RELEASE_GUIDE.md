# Release Guide for Axle v0.1

## Pre-Release Checklist

### 1. Update Module Path (IMPORTANT!)

Before releasing, update the Go module to use your GitHub path:

```bash
# Replace 'yourusername' with your actual GitHub username
go mod edit -module github.com/yourusername/axle-file-sync
```

Then update all imports in the code:

```bash
# Update imports in all .go files
find . -name "*.go" -type f -exec sed -i 's|"axle/|"github.com/yourusername/axle-file-sync/|g' {} +
```

### 2. Update Version References

- Edit `cmd/version.go` - update the GitHub URL with your username
- Edit `install.ps1` - update repo variable with your username
- Edit `install.sh` - update repo variable with your username
- Edit `.github/workflows/release.yml` - this will use the repository context automatically

### 3. Clean Build Test

```bash
go build -o axle.exe
./axle.exe version
```

## Creating a GitHub Release

### Step 1: Create and Push Repository

```bash
# Initialize git if not already done
git init

# Add all files
git add .

# Commit
git commit -m "feat: Initial release of Axle v0.1.0

Real-time file synchronization for hackathon teams.

Features:
- Git-based sync with Redis pub/sub
- Conflict resolution strategies
- Team presence and chat
- Desktop notifications for priority messages
- Cross-platform support"

# Add remote (replace with your repo URL)
git remote add origin https://github.com/yourusername/axle-file-sync.git

# Push to main branch
git branch -M main
git push -u origin main
```

### Step 2: Create Release Tag

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release version 0.1.0"

# Push tag
git push origin v0.1.0
```

### Step 3: GitHub Actions Will Automatically:

Once you push the tag, the GitHub Actions workflow (`.github/workflows/release.yml`) will:

1. Build binaries for all platforms
2. Create a GitHub release
3. Upload the binaries

### Step 4: Manual Release (If GitHub Actions fails)

#### Build Binaries Locally:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o axle.exe
zip axle-windows-amd64.zip axle.exe

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o axle
tar czf axle-darwin-amd64.tar.gz axle

# macOS ARM (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o axle
tar czf axle-darwin-arm64.tar.gz axle

# Linux
GOOS=linux GOARCH=amd64 go build -o axle
tar czf axle-linux-amd64.tar.gz axle
```

#### Create Release on GitHub:

1. Go to https://github.com/yourusername/axle-file-sync/releases
2. Click "Create a new release"
3. Choose tag: v0.1.0
4. Release title: "Axle v0.1.0 - Initial Release"
5. Upload the binary files
6. Add release notes:

````markdown
## 🚀 Axle v0.1.0 - Initial Release

Real-time file synchronization for hackathon teams and rapid prototyping.

### ✨ Features

- **Real-time sync**: Git-based file synchronization with Redis pub/sub
- **Conflict resolution**: 5 strategies (theirs, mine, merge, backup, interactive)
- **Team presence**: See who's online and working
- **Team chat**: Built-in chat with priority notifications
- **Desktop notifications**: Get alerts for important messages
- **Cross-platform**: Windows, macOS, Linux support
- **Smart .gitignore**: Auto-detects project type and configures ignores
- **Dynamic batching**: Intelligent file change batching (1-5 seconds)

### 📦 Installation

#### Quick Install

**Windows:**

```powershell
irm https://raw.githubusercontent.com/yourusername/axle-file-sync/main/install.ps1 | iex
```
````

**macOS/Linux:**

```bash
curl -sSL https://raw.githubusercontent.com/yourusername/axle-file-sync/main/install.sh | bash
```

#### Manual Download

Download the appropriate binary for your platform below.

### 🚀 Quick Start

1. **Initialize team (first member):**

   ```bash
   axle init --team my-team --username alice
   ```

2. **Join team (other members):**

   ```bash
   axle join --team my-team --username bob
   ```

3. **Start syncing:**
   ```bash
   axle start
   ```

### 📖 Documentation

See [COMMANDS.md](https://github.com/yourusername/axle-file-sync/blob/main/COMMANDS.md) for full command reference.

### 🔧 Requirements

- Git (must be in PATH)
- Redis server (local or remote)
- Go 1.20+ (only for building from source)

### 🐛 Known Issues

- First sync between independent repos may show patch errors (self-resolving)
- Large binary files should be added to .gitignore

### 💡 Tips for Hackathon Teams

1. Use `--conflict merge` for most team members
2. Send priority messages with `axle chat -p "urgent message"`
3. Check team status with `axle team`
4. View sync stats with `axle stats`

```

7. Click "Publish release"

## Post-Release

### Update Installation Scripts
After creating the release, update the download URLs in:
- `install.ps1` - update version and download URL
- `install.sh` - update version and download URL

### Announce
Share your release on:
- Twitter/X
- Reddit (r/golang, r/programming)
- Hacker News
- Dev.to
- Your team/community channels

## Future Releases

For future releases (v0.2.0, etc.):
1. Update version in `cmd/version.go`
2. Update CHANGELOG.md
3. Create new tag: `git tag v0.2.0`
4. Push tag: `git push origin v0.2.0`
5. GitHub Actions handles the rest!
```
