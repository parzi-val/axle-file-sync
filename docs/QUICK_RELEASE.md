# 🚀 Quick Release Steps for v0.1.0

Everything is already configured! Just follow these steps:

## 1️⃣ Create GitHub Repository
Go to https://github.com/new
- Name: `axle-file-sync`
- Description: "Real-time file synchronization for hackathon teams"
- Public repository

## 2️⃣ Push Your Code
```bash
git init
git add .
git commit -m "feat: Initial release of Axle v0.1.0

Real-time file synchronization for hackathon teams"

git remote add origin https://github.com/parzi-val/axle-file-sync.git
git branch -M main
git push -u origin main
```

## 3️⃣ Create Release Tag
```bash
git tag -a v0.1.0 -m "Release version 0.1.0"
git push origin v0.1.0
```

## 4️⃣ Automatic Release
GitHub Actions will automatically:
- Build binaries for all platforms
- Create a release
- Upload the binaries

## 5️⃣ What the Install Scripts Do
After your release is created, users can install Axle with one command:

**Windows:**
```powershell
irm https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/install.ps1 | iex
```

**Mac/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/install.sh | bash
```

These scripts:
1. Download the binary from your GitHub release
2. Install it to the right location
3. Add to PATH
4. User can immediately run `axle` commands

That's it! 🎉